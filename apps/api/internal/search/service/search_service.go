package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"kun-galgame-api/internal/galgame/client"
	galgameDto "kun-galgame-api/internal/galgame/dto"
	galgameService "kun-galgame-api/internal/galgame/service"
	"kun-galgame-api/internal/search/dto"
	"kun-galgame-api/internal/search/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"
)

type SearchService struct {
	repo          *repository.SearchRepository
	galgameClient *client.GalgameClient
	enricher      *galgameService.GalgameEnricher
	userClient    *userclient.Client
}

func NewSearchService(
	repo *repository.SearchRepository,
	galgameClient *client.GalgameClient,
	enricher *galgameService.GalgameEnricher,
	userClient *userclient.Client,
) *SearchService {
	return &SearchService{repo: repo, galgameClient: galgameClient, enricher: enricher, userClient: userClient}
}

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
	pollSet := s.repo.FindTopicIDsWithPoll(topicIDs)

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
			IsPollTopic:      pollSet[r.ID],
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

	_ = page
	users, err := s.userClient.SearchUsers(ctx, raw, limit)
	if err != nil {
		return nil, errors.ErrInternal(fmt.Sprintf("用户搜索失败: %v", err))
	}

	items := make([]dto.UserItem, 0, len(users))
	for _, u := range users {
		if u.Status != 0 {
			continue
		}
		items = append(items, dto.UserItem{
			ID:     u.ID,
			Name:   u.Name,
			Avatar: u.Avatar,
			Bio:    u.Bio,
		})
	}
	return &dto.PaginatedResult[dto.UserItem]{Items: items, Total: int64(len(items))}, nil
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
