package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/repository"
)

const (
	contentLimitSyncChunk   = 500
	contentLimitSyncTimeout = 15 * time.Minute
)

// Keeps galgame.content_limit in step with catalog's editorial verdict. There
// is no feed for that field — the claim-event feed carries claim state only —
// so a full sweep is the only way an editor's flip reaches the local lists.
type GalgameContentLimitSync struct {
	galgameClient *client.GalgameClient
	galgameRepo   *repository.GalgameRepository

	running sync.Mutex
	// Local rows catalog has no work for. They stay NULL, so without this the
	// ten-minute pass would ask about the same orphans forever.
	unresolvedMu sync.Mutex
	unresolved   map[int]bool
}

func NewGalgameContentLimitSync(
	galgameClient *client.GalgameClient,
	galgameRepo *repository.GalgameRepository,
) *GalgameContentLimitSync {
	return &GalgameContentLimitSync{
		galgameClient: galgameClient,
		galgameRepo:   galgameRepo,
		unresolved:    map[int]bool{},
	}
}

func (s *GalgameContentLimitSync) RunAll() {
	s.run("全量", s.galgameRepo.ContentLimitSyncIDs(false, 0))
}

// Rows created since the last sweep, so a game published at noon is filtered
// correctly the same day instead of leaving a hole in every SFW page until
// tomorrow. Uncapped because the one run that is not nearly empty is the first
// one after the column ships, which is exactly the run that should not be
// spread over an hour.
func (s *GalgameContentLimitSync) RunPending() {
	s.run("增量", s.resolvable(s.galgameRepo.ContentLimitSyncIDs(true, 0)))
}

func (s *GalgameContentLimitSync) run(mode string, ids []int) {
	if s.galgameClient == nil || len(ids) == 0 {
		return
	}
	if !s.running.TryLock() {
		slog.Info("content_limit 同步仍在进行, 跳过本轮", "mode", mode)
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
			slog.Warn("catalog content_limit 拉取失败, 本轮中止", "mode", mode, "offset", start, "error", appErr.Message)
			break
		}
		n, err := s.galgameRepo.SetContentLimits(groupByContentLimit(limits))
		if err != nil {
			slog.Warn("content_limit 入库失败, 本轮中止", "mode", mode, "offset", start, "error", err)
			break
		}
		s.rememberUnresolved(chunk, limits)
		seen += int64(len(limits))
		updated += n
	}
	slog.Info("galgame content_limit 同步完成",
		"mode", mode, "requested", len(ids), "resolved", seen, "updated", updated)
}

func (s *GalgameContentLimitSync) resolvable(ids []int) []int {
	s.unresolvedMu.Lock()
	defer s.unresolvedMu.Unlock()
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if !s.unresolved[id] {
			out = append(out, id)
		}
	}
	return out
}

func (s *GalgameContentLimitSync) rememberUnresolved(asked []int, got map[int]string) {
	s.unresolvedMu.Lock()
	defer s.unresolvedMu.Unlock()
	for _, id := range asked {
		if _, ok := got[id]; ok {
			delete(s.unresolved, id)
			continue
		}
		s.unresolved[id] = true
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
