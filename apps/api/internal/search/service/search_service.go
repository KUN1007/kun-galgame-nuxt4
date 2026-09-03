package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"kun-galgame-api/internal/galgame/client"
	galgameDto "kun-galgame-api/internal/galgame/dto"
	galgameService "kun-galgame-api/internal/galgame/service"
	"kun-galgame-api/internal/search/dto"
	"kun-galgame-api/internal/search/repository"
	toolsetService "kun-galgame-api/internal/toolset/service"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/role"
	"kun-galgame-api/pkg/userclient"
)

type SearchService struct {
	repo          *repository.SearchRepository
	galgameClient *client.GalgameClient
	enricher      *galgameService.GalgameEnricher
	userClient    *userclient.Client
	entityService *galgameService.EntitySearchService
	toolset       *toolsetService.ToolsetService
	resource      *galgameService.ResourceService
}

func NewSearchService(
	repo *repository.SearchRepository,
	galgameClient *client.GalgameClient,
	enricher *galgameService.GalgameEnricher,
	userClient *userclient.Client,
	entityService *galgameService.EntitySearchService,
	toolset *toolsetService.ToolsetService,
	resource *galgameService.ResourceService,
) *SearchService {
	return &SearchService{
		repo:          repo,
		galgameClient: galgameClient,
		enricher:      enricher,
		userClient:    userClient,
		entityService: entityService,
		toolset:       toolset,
		resource:      resource,
	}
}

// OAuth's /users/search takes q and limit only — it has no offset, so the page
// is cut here out of one capped fetch rather than asked for upstream.
const userSearchMax = 50

func tokenize(raw string) ([]string, *errors.AppError) {
	keywords := strings.Fields(strings.TrimSpace(raw))
	if len(keywords) == 0 {
		return nil, errors.ErrBadRequest("搜索关键词不能为空")
	}
	return keywords, nil
}

func (s *SearchService) SearchTopics(ctx context.Context, raw string, page, limit int) (*dto.PaginatedResult[dto.TopicItem], *errors.AppError) {
	keywords, appErr := tokenize(raw)
	if appErr != nil {
		return nil, appErr
	}
	rows, total := s.repo.SearchTopics(keywords, page, limit)

	uids := userclient.CollectIDs(rows, func(r repository.TopicRow) int { return r.UserID })
	userMap := s.userClient.Hydrate(ctx, uids)

	topicIDs := make([]int, len(rows))
	for i, r := range rows {
		topicIDs[i] = r.ID
	}
	sectionMap := map[int][]string{}
	for _, sct := range s.repo.FindTopicSections(topicIDs) {
		sectionMap[sct.TopicID] = append(sectionMap[sct.TopicID], sct.SectionName)
	}
	miniApps := s.repo.FindTopicMiniApps(topicIDs)

	items := make([]dto.TopicItem, 0, len(rows))
	for _, r := range rows {
		u := userMap[r.UserID]
		if !userclient.IsRenderable(u) {
			continue
		}
		sections := sectionMap[r.ID]
		if sections == nil {
			sections = []string{}
		}
		items = append(items, dto.TopicItem{
			ID: r.ID, Title: r.Title, View: r.View, Status: r.Status,
			LikeCount: r.LikeCount, ReplyCount: r.ReplyCount,
			CommentCount:     r.CommentCount,
			HasBestAnswer:    r.BestAnswerID != nil,
			MiniApps:         miniApps[r.ID],
			IsNSFWTopic:      r.IsNSFW,
			Section:          sections,
			UpvoteTime:       r.UpvoteTime,
			StatusUpdateTime: r.StatusUpdateTime,
			User:             dto.UserBrief{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
		})
	}
	return &dto.PaginatedResult[dto.TopicItem]{Items: items, Total: total}, nil
}

func (s *SearchService) SearchUsers(
	ctx context.Context,
	raw string,
	page, limit int,
) (*dto.PaginatedResult[dto.UserItem], *errors.AppError) {
	if _, appErr := tokenize(raw); appErr != nil {
		return nil, appErr
	}
	if s.userClient == nil {
		return nil, errors.ErrInternal("用户搜索未启用")
	}

	users, err := s.userClient.SearchUsers(ctx, raw, userSearchMax)
	if err != nil {
		return nil, errors.ErrInternal(fmt.Sprintf("用户搜索失败: %v", err))
	}

	items := make([]dto.UserItem, 0, len(users))
	for _, u := range users {
		if u.Status != 0 {
			continue
		}
		items = append(items, dto.UserItem{
			ID:      u.ID,
			Name:    u.Name,
			Avatar:  u.Avatar,
			Bio:     u.Bio,
			Roles:   role.Union(u.Roles, u.SiteRoles),
			Created: parseUserCreated(u.CreatedAt),
		})
	}

	total := int64(len(items))
	start := (page - 1) * limit
	if start < 0 || start >= len(items) {
		return &dto.PaginatedResult[dto.UserItem]{Items: []dto.UserItem{}, Total: total}, nil
	}
	items = items[start:min(start+limit, len(items))]

	ids := make([]int, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	topics, replies := s.repo.CountUserPosts(ids)
	for i := range items {
		items[i].TopicCount, items[i].ReplyCount = topics[items[i].ID], replies[items[i].ID]
	}

	return &dto.PaginatedResult[dto.UserItem]{Items: items, Total: total}, nil
}

func parseUserCreated(raw string) *time.Time {
	if raw == "" {
		return nil
	}
	at, err := time.Parse(time.RFC3339, raw)
	if err != nil || at.IsZero() {
		return nil
	}
	return &at
}

func (s *SearchService) SearchReplies(ctx context.Context, raw string, page, limit int) (*dto.PaginatedResult[dto.ReplyItem], *errors.AppError) {
	keywords, appErr := tokenize(raw)
	if appErr != nil {
		return nil, appErr
	}
	rows, total := s.repo.SearchReplies(keywords, page, limit)

	uids := userclient.CollectIDs(rows, func(r repository.ReplyRow) int { return r.UserID })
	for _, r := range rows {
		if r.TopicUserID > 0 {
			uids = append(uids, r.TopicUserID)
		}
	}
	userMap := s.userClient.Hydrate(ctx, uids)

	items := make([]dto.ReplyItem, 0, len(rows))
	for _, r := range rows {
		u := userMap[r.UserID]
		if !userclient.IsRenderable(u) {
			continue
		}
		if r.TopicUserID > 0 && !userclient.IsRenderable(userMap[r.TopicUserID]) {
			continue
		}
		items = append(items, dto.ReplyItem{
			ID: r.ID, TopicID: r.TopicID, TopicTitle: r.TopicTitle,
			Content: r.Content, Floor: r.Floor,
			User:    dto.UserBrief{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			Created: r.Created,
		})
	}
	return &dto.PaginatedResult[dto.ReplyItem]{Items: items, Total: total}, nil
}

func (s *SearchService) SearchGalgames(
	ctx context.Context,
	raw string,
	page, limit int,
	isSFW bool,
) (*dto.PaginatedResult[galgameDto.GalgameCard], *errors.AppError) {
	if _, appErr := tokenize(raw); appErr != nil {
		return nil, appErr
	}
	if s.galgameClient == nil || s.enricher == nil {
		return nil, errors.ErrInternal("Galgame 搜索未启用")
	}

	q := url.Values{
		"q":       {raw},
		"page":    {strconv.Itoa(page)},
		"limit":   {strconv.Itoa(limit)},
		"include": {galgameService.CatalogCardInclude},
		"sort":    {"relevance"},
	}
	client.ApplyWorksGate(q, isSFW)

	res, appErr := s.galgameClient.CatalogWorksSearch(ctx, q)
	if appErr != nil {
		return nil, appErr
	}
	items := make([]galgameDto.NextMoeGalgameItem, 0, len(res.Items))
	for i := range res.Items {
		if !client.CatalogItemRenderable(&res.Items[i]) {
			continue
		}
		items = append(items, client.CatalogItemToNextMoeItem(ctx, &res.Items[i]))
	}

	return &dto.PaginatedResult[galgameDto.GalgameCard]{
		Items: s.enricher.ToCards(ctx, items),
		Total: res.Total,
	}, nil
}

func (s *SearchService) SearchComments(ctx context.Context, raw string, page, limit int) (*dto.PaginatedResult[dto.CommentItem], *errors.AppError) {
	keywords, appErr := tokenize(raw)
	if appErr != nil {
		return nil, appErr
	}
	rows, total := s.repo.SearchComments(keywords, page, limit)

	uids := userclient.CollectIDs(rows, func(r repository.CommentRow) int { return r.UserID })
	for _, r := range rows {
		if r.TopicUserID > 0 {
			uids = append(uids, r.TopicUserID)
		}
	}
	userMap := s.userClient.Hydrate(ctx, uids)

	items := make([]dto.CommentItem, 0, len(rows))
	for _, r := range rows {
		u := userMap[r.UserID]
		if !userclient.IsRenderable(u) {
			continue
		}
		if r.TopicUserID > 0 && !userclient.IsRenderable(userMap[r.TopicUserID]) {
			continue
		}
		items = append(items, dto.CommentItem{
			ID: r.ID, TopicID: r.TopicID, TopicTitle: r.TopicTitle,
			Content: r.Content,
			User:    dto.UserBrief{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			Created: r.Created,
		})
	}
	return &dto.PaginatedResult[dto.CommentItem]{Items: items, Total: total}, nil
}

// quickSearchLimit is per lane: the palette is a preview, the search page is the
// whole result.
const quickSearchLimit = 5

// QuickSearch answers the command palette. A lane that fails is dropped rather
// than failing the request — catalog and OAuth are remote, and a palette that
// blanks out because one of them is slow is worse than one showing topics only.
func (s *SearchService) QuickSearch(
	ctx context.Context,
	raw string,
) (*dto.QuickSearchResult, *errors.AppError) {
	if _, appErr := tokenize(raw); appErr != nil {
		return nil, appErr
	}

	var (
		wg       sync.WaitGroup
		topics   *dto.PaginatedResult[dto.TopicItem]
		galgames *dto.PaginatedResult[galgameDto.GalgameCard]
		users    *dto.PaginatedResult[dto.UserItem]
	)
	run := func(lane func()) {
		wg.Add(1)
		go func() {
			// fiber's recover middleware only wraps the handler goroutine, so a
			// panic in a lane would take the process down with it instead of
			// failing this one request.
			defer func() {
				if r := recover(); r != nil {
					slog.Error("quick search lane panicked", "panic", r)
				}
			}()
			defer wg.Done()
			lane()
		}()
	}
	run(func() { topics, _ = s.SearchTopics(ctx, raw, 1, quickSearchLimit) })
	run(func() {
		galgames, _ = s.SearchGalgames(ctx, raw, 1, quickSearchLimit, false)
	})
	run(func() { users, _ = s.SearchUsers(ctx, raw, 1, quickSearchLimit) })
	wg.Wait()

	res := &dto.QuickSearchResult{
		Topics:   []dto.TopicItem{},
		Galgames: []galgameDto.GalgameCard{},
		Users:    []dto.UserItem{},
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
	return res, nil
}
