package service

import (
	"context"
	"log/slog"
	"sync"

	galgameDto "kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/internal/search/dto"
	toolsetDto "kun-galgame-api/internal/toolset/dto"
	"kun-galgame-api/pkg/errors"
)

// One lane's slice of the overview. Long-form lanes (topics, games) get more
// rows than the ones that are only there to say "your keyword also lives here".
const (
	overviewTopicLimit   = 4
	overviewGalgameLimit = 6
	overviewEntityLimit  = 5
	overviewUserLimit    = 6
	overviewReplyLimit   = 3
	overviewToolsetLimit = 3
)

const entitySearchDefaultLimit = 24

// Overview answers the search page's default tab: every lane at once, so the
// reader sees which categories their keyword actually lives in before choosing
// one. Lanes run concurrently and a failed lane is dropped, the same trade the
// command palette makes — catalog and OAuth are remote.
func (s *SearchService) Overview(
	ctx context.Context,
	raw string,
	isSFW bool,
) (*dto.OverviewResult, *errors.AppError) {
	if _, appErr := tokenize(raw); appErr != nil {
		return nil, appErr
	}

	var (
		wg       sync.WaitGroup
		topics   *dto.PaginatedResult[dto.TopicItem]
		galgames *dto.PaginatedResult[galgameDto.GalgameCard]
		users    *dto.PaginatedResult[dto.UserItem]
		replies  *dto.PaginatedResult[dto.ReplyItem]
		comments *dto.PaginatedResult[dto.CommentItem]
		entities []galgameDto.EntitySearchGroup
		toolsets []toolsetDto.ToolsetCard
		toolsetN int64
	)
	run := func(name string, lane func()) {
		wg.Add(1)
		go func() {
			// fiber's recover middleware only wraps the handler goroutine, so a
			// panic in a lane would take the process down with it instead of
			// failing this one request.
			defer func() {
				if r := recover(); r != nil {
					slog.Error("search overview lane panicked", "lane", name, "panic", r)
				}
			}()
			defer wg.Done()
			lane()
		}()
	}
	run("topic", func() { topics, _ = s.SearchTopics(ctx, raw, 1, overviewTopicLimit) })
	run("galgame", func() {
		galgames, _ = s.SearchGalgames(ctx, raw, 1, overviewGalgameLimit, false)
	})
	run("user", func() { users, _ = s.SearchUsers(ctx, raw, 1, overviewUserLimit) })
	run("reply", func() { replies, _ = s.SearchReplies(ctx, raw, 1, overviewReplyLimit) })
	run("comment", func() { comments, _ = s.SearchComments(ctx, raw, 1, overviewReplyLimit) })
	run("entity", func() {
		entities, _ = s.SearchEntities(ctx, raw, "", overviewEntityLimit, isSFW)
	})
	run("toolset", func() {
		toolsets, toolsetN = s.SearchToolsets(ctx, raw, 1, overviewToolsetLimit)
	})
	wg.Wait()

	res := &dto.OverviewResult{
		Topics:   []dto.TopicItem{},
		Galgames: []galgameDto.GalgameCard{},
		Entities: []galgameDto.EntitySearchGroup{},
		Users:    []dto.UserItem{},
		Replies:  []dto.ReplyItem{},
		Comments: []dto.CommentItem{},
		Toolsets: toolsets,
	}
	if topics != nil {
		res.Topics, res.Totals.Topic = topics.Items, topics.Total
	}
	if galgames != nil {
		res.Galgames, res.Totals.Galgame = galgames.Items, galgames.Total
	}
	if users != nil {
		res.Users, res.Totals.User = users.Items, users.Total
	}
	if replies != nil {
		res.Replies, res.Totals.Reply = replies.Items, replies.Total
	}
	if comments != nil {
		res.Comments, res.Totals.Comment = comments.Items, comments.Total
	}
	if entities != nil {
		res.Entities = entities
		for _, group := range entities {
			res.Totals.Entity += group.Total
		}
	}
	res.Totals.Toolset = toolsetN
	return res, nil
}

func (s *SearchService) SearchEntities(
	ctx context.Context,
	raw string,
	family string,
	limit int,
	isSFW bool,
) ([]galgameDto.EntitySearchGroup, *errors.AppError) {
	if _, appErr := tokenize(raw); appErr != nil {
		return nil, appErr
	}
	if s.entityService == nil {
		return nil, errors.ErrInternal("Galgame 资料库搜索未启用")
	}
	if limit <= 0 {
		limit = entitySearchDefaultLimit
	}
	return s.entityService.Search(ctx, raw, family, limit, isSFW)
}

func (s *SearchService) SearchToolsets(
	ctx context.Context,
	raw string,
	page, limit int,
) ([]toolsetDto.ToolsetCard, int64) {
	if s.toolset == nil {
		return []toolsetDto.ToolsetCard{}, 0
	}
	return s.toolset.GetList(ctx, &toolsetDto.ToolsetListRequest{
		Page:  page,
		Limit: limit,
		Query: raw,
	})
}
