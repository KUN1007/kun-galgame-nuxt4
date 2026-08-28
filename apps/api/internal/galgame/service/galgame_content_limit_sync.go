package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/repository"

	"github.com/redis/go-redis/v9"
)

const (
	contentLimitSyncChunk   = 500
	contentLimitSyncTimeout = 15 * time.Minute

	// Losing this key is not a fault: an absent cursor drains the feed from the
	// head of the inventory, which is a full re-mirror. Deleting it is also how
	// an operator forces one.
	contentLimitCursorKey = "catalog:changes:content-limit:cursor"

	// 100 pages a tick is 10000 works, and the ten-minute beat gives 60000 an
	// hour — enough that the 110k-work night of a machine re-grade drains inside
	// two hours, and enough to walk the whole inventory in about the same time.
	contentLimitMirrorPages = 100

	// A local row catalog has no work for cannot be resolved by asking again a
	// minute later, but it can be resolved by asking again after a submission is
	// approved. The nightly full sweep used to be what re-asked; the mirror
	// channel never will, because a row that is missing upstream is exactly the
	// row the channel has nothing to say about.
	contentLimitOrphanTTL = 6 * time.Hour
)

// Keeps galgame.content_limit in step with catalog's editorial verdict.
//
// Two lanes, and they answer different questions. The mirror channel carries
// every catalog-side change — an editor flipping display_nsfw reaches the local
// SFW lists within a tick. The pending fill carries the local side: a stub row
// created by a user this minute has no catalog change to its name and would
// otherwise stay NULL until its work happens to be touched upstream.
type GalgameContentLimitSync struct {
	galgameClient *client.GalgameClient
	galgameRepo   *repository.GalgameRepository
	rdb           *redis.Client
	maxPages      int

	running sync.Mutex
	// Local rows catalog has no work for. They stay NULL, so without this the
	// ten-minute pass would ask about the same orphans forever.
	unresolvedMu sync.Mutex
	unresolved   map[int]time.Time
}

func NewGalgameContentLimitSync(
	galgameClient *client.GalgameClient,
	galgameRepo *repository.GalgameRepository,
	rdb *redis.Client,
) *GalgameContentLimitSync {
	return &GalgameContentLimitSync{
		galgameClient: galgameClient,
		galgameRepo:   galgameRepo,
		rdb:           rdb,
		maxPages:      contentLimitMirrorPages,
		unresolved:    map[int]time.Time{},
	}
}

// RunMirror drains catalog's changes feed from the stored cursor and applies the
// display verdict of every changed work the forum mirrors.
//
// The cursor advances one page at a time and only after that page's rows are in
// the database, so a failure mid-drain keeps the pages already applied and
// resumes at the first one that is not — the whole-round abort of the old sweep
// cost every chunk after the first error.
//
// The last page of a drain has no next_cursor, so the stored cursor stays on the
// page before it and that page is read again next tick. That re-read is the only
// way to notice rows appended since, and re-applying it is free: the UPDATE is
// keyed IS DISTINCT FROM.
func (s *GalgameContentLimitSync) RunMirror() {
	if s.galgameClient == nil || s.rdb == nil {
		return
	}
	if !s.running.TryLock() {
		slog.Info("content_limit 同步仍在进行, 跳过本轮", "mode", "信道")
		return
	}
	defer s.running.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), contentLimitSyncTimeout)
	defer cancel()

	cursor, err := s.rdb.Get(ctx, contentLimitCursorKey).Result()
	if err != nil && err != redis.Nil {
		slog.Warn("读取 catalog 变更游标失败 (本轮跳过)", "error", err)
		return
	}
	bootstrap := cursor == ""

	var changed, gone, matched, updated int64
	pages := 0
	for ; pages < s.maxPages; pages++ {
		page, appErr := s.galgameClient.CatalogChanges(ctx, cursor, client.CatalogChangesLimit)
		if appErr != nil {
			slog.Warn("catalog 变更信道拉取失败, 游标保持不动", "pages", pages, "error", appErr.Message)
			break
		}
		if len(page.Items) == 0 && page.NextCursor == "" {
			break
		}
		ids := make([]int64, 0, len(page.Items))
		for _, it := range page.Items {
			if it.Gone {
				gone++
				continue
			}
			ids = append(ids, it.ID)
		}
		changed += int64(len(page.Items))

		limits, appErr := s.galgameClient.ContentLimitsByCatalogIDs(ctx, ids)
		if appErr != nil {
			slog.Warn("catalog 变更条目水合失败, 游标保持不动", "pages", pages, "error", appErr.Message)
			break
		}
		matched += int64(len(limits))
		if len(limits) > 0 {
			n, err := s.galgameRepo.SetContentLimits(groupByContentLimit(limits))
			if err != nil {
				slog.Warn("content_limit 入库失败, 游标保持不动", "pages", pages, "error", err)
				break
			}
			updated += n
		}

		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
		if err := s.rdb.Set(ctx, contentLimitCursorKey, cursor, 0).Err(); err != nil {
			slog.Warn("写入 catalog 变更游标失败, 下轮将重放本页", "error", err)
			break
		}
	}

	if changed == 0 {
		return
	}
	slog.Info("galgame content_limit 信道同步完成", "bootstrap", bootstrap,
		"pages", pages, "changed", changed, "gone", gone, "matched", matched, "updated", updated)
}

// RunPending resolves rows created since the last pass, so a game published at
// noon is filtered correctly the same day instead of leaving a hole in every SFW
// page until its catalog work next changes.
func (s *GalgameContentLimitSync) RunPending() {
	if s.galgameClient == nil || s.galgameRepo == nil {
		return
	}
	ids := s.resolvable(s.galgameRepo.ContentLimitMissingIDs())
	if len(ids) == 0 {
		return
	}
	if !s.running.TryLock() {
		slog.Info("content_limit 同步仍在进行, 跳过本轮", "mode", "增量")
		return
	}
	defer s.running.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), contentLimitSyncTimeout)
	defer cancel()

	var seen, updated int64
	for start := 0; start < len(ids); start += contentLimitSyncChunk {
		chunk := ids[start:min(start+contentLimitSyncChunk, len(ids))]
		limits, appErr := s.galgameClient.ContentLimitsByGIDs(ctx, chunk)
		if appErr != nil {
			slog.Warn("catalog content_limit 拉取失败, 本轮中止", "offset", start, "error", appErr.Message)
			break
		}
		n, err := s.galgameRepo.SetContentLimits(groupByContentLimit(limits))
		if err != nil {
			slog.Warn("content_limit 入库失败, 本轮中止", "offset", start, "error", err)
			break
		}
		s.rememberUnresolved(chunk, limits)
		seen += int64(len(limits))
		updated += n
	}
	slog.Info("galgame content_limit 增量同步完成",
		"requested", len(ids), "resolved", seen, "updated", updated)
}

func (s *GalgameContentLimitSync) resolvable(ids []int) []int {
	now := time.Now()
	s.unresolvedMu.Lock()
	defer s.unresolvedMu.Unlock()
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if until, ok := s.unresolved[id]; ok && now.Before(until) {
			continue
		}
		out = append(out, id)
	}
	return out
}

func (s *GalgameContentLimitSync) rememberUnresolved(asked []int, got map[int]string) {
	until := time.Now().Add(contentLimitOrphanTTL)
	s.unresolvedMu.Lock()
	defer s.unresolvedMu.Unlock()
	for _, id := range asked {
		if _, ok := got[id]; ok {
			delete(s.unresolved, id)
			continue
		}
		s.unresolved[id] = until
	}
}

func groupByContentLimit(limits map[int]string) map[string][]int {
	out := make(map[string][]int, 2)
	for gid, limit := range limits {
		switch limit {
		case "sfw", "nsfw":
			out[limit] = append(out[limit], gid)
		}
	}
	return out
}
