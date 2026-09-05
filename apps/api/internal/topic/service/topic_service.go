package service

import (
	"context"

	"kun-galgame-api/internal/infrastructure/markdown"
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/topic/dto"
	"kun-galgame-api/internal/topic/repository"
	userRepo "kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type TopicService struct {
	topicRepo    *repository.TopicRepository
	listRepo     *repository.TopicListRepository
	taxonomyRepo *repository.TopicTaxonomyRepository
	rdb          *redis.Client
	userClient   *userclient.Client
	stateRepo    *userRepo.StateRepository
}

func NewTopicService(
	topicRepo *repository.TopicRepository,
	listRepo *repository.TopicListRepository,
	taxonomyRepo *repository.TopicTaxonomyRepository,
	rdb *redis.Client,
	userClient *userclient.Client,
	stateRepo *userRepo.StateRepository,
) *TopicService {
	return &TopicService{
		topicRepo:    topicRepo,
		listRepo:     listRepo,
		taxonomyRepo: taxonomyRepo,
		rdb:          rdb,
		userClient:   userClient,
		stateRepo:    stateRepo,
	}
}

func (s *TopicService) GetMyInteractions(userID int) dto.MyTopicInteractions {
	favorited, reactions, err := s.topicRepo.UserTopicInteractions(userID)
	if err != nil {
		return dto.MyTopicInteractions{Favorited: []int{}, Reactions: map[int][]string{}}
	}
	return dto.MyTopicInteractions{Favorited: favorited, Reactions: reactions}
}

const topicUpvoteRecordLimit = 50

func (s *TopicService) GetTopicUpvotes(ctx context.Context, topicID int) ([]dto.TopicUpvoteRecord, *errors.AppError) {
	rows, err := s.topicRepo.FetchTopicUpvotes(topicID, topicUpvoteRecordLimit)
	if err != nil {
		return nil, errors.ErrInternal("操作失败")
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	userMap := s.userClient.Hydrate(ctx, ids)
	out := make([]dto.TopicUpvoteRecord, 0, len(rows))
	for _, row := range rows {
		u := userMap[row.UserID]
		if !userclient.IsRenderable(u) {
			continue
		}
		out = append(out, dto.TopicUpvoteRecord{
			ID:          row.ID,
			User:        dto.KunUser{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			Description: row.Description,
			Created:     row.Created,
		})
	}
	return out, nil
}

const topicReactionHistoryLimit = 300

func (s *TopicService) GetTopicReactionHistory(ctx context.Context, topicID int) ([]dto.ReactionHistoryItem, *errors.AppError) {
	rows, err := s.topicRepo.GetTopicReactionHistory(topicID, topicReactionHistoryLimit)
	if err != nil {
		return nil, errors.ErrInternal("操作失败")
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	userMap := s.userClient.Hydrate(ctx, ids)
	out := make([]dto.ReactionHistoryItem, 0, len(rows))
	for _, row := range rows {
		u := userMap[row.UserID]
		out = append(out, dto.ReactionHistoryItem{
			User:     dto.KunUser{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			Reaction: row.Reaction,
			Created:  row.Created,
		})
	}
	return out, nil
}

func (s *TopicService) GetList(
	ctx context.Context,
	req *dto.ListTopicsRequest,
	isNSFW, authenticated bool,
) ([]dto.TopicCard, int64, *errors.AppError) {
	rows, total, err := s.listRepo.FindList(
		req.Page, req.Limit,
		req.SortField, req.SortOrder, req.Category,
		isNSFW, authenticated,
	)
	if err != nil {
		return nil, 0, errors.ErrInternal("获取话题列表失败")
	}

	return s.mapListRows(ctx, rows, total)
}

func (s *TopicService) GetResourceList(
	ctx context.Context,
	req *dto.ListTopicsRequest,
	isNSFW, authenticated bool,
) ([]dto.TopicCard, int64, *errors.AppError) {
	rows, total, err := s.listRepo.FindResourceList(
		req.Page, req.Limit,
		req.SortField, req.SortOrder, req.Category,
		isNSFW, authenticated,
	)
	if err != nil {
		return nil, 0, errors.ErrInternal("获取资源话题列表失败")
	}

	return s.mapListRows(ctx, rows, total)
}

func (s *TopicService) mapListRows(ctx context.Context, rows []repository.TopicCardRow, total int64) ([]dto.TopicCard, int64, *errors.AppError) {
	topicIDs := make([]int, len(rows))
	for i, r := range rows {
		topicIDs[i] = r.ID
	}

	sectionMap, _ := s.taxonomyRepo.FindSectionNamesByTopicIDs(topicIDs)
	miniApps := s.topicRepo.FindTopicMiniApps(topicIDs)

	uids := userclient.CollectIDs(rows, func(r repository.TopicCardRow) int { return r.UserID })
	userMap := s.userClient.Hydrate(ctx, uids)
	for i := range rows {
		u := userMap[rows[i].UserID]
		rows[i].UserName = u.Name
		rows[i].UserAvatar = u.Avatar
	}

	cards := make([]dto.TopicCard, 0, len(rows))
	for i, r := range rows {
		if u, ok := userMap[r.UserID]; ok && !userclient.IsRenderable(u) {
			continue
		}
		cards = append(cards, toTopicCard(r, sectionMap[r.ID], miniApps[r.ID]))
		_ = i
	}
	return cards, total, nil
}

func (s *TopicService) GetDetail(
	ctx context.Context,
	topicID int,
	userInfo *middleware.UserInfo,
) (*dto.TopicDetail, *errors.AppError) {
	topic, err := s.topicRepo.FindByID(topicID)
	if err != nil {
		return nil, errors.ErrNotFound("未找到该话题")
	}
	grants, appErr := requireTopicRead(s.topicRepo, topic, userInfo)
	if appErr != nil {
		return nil, appErr
	}

	g, _ := errgroup.WithContext(ctx)

	var author *repository.TopicAuthorUser
	var authorBanned bool
	var sections []string
	var miniApps []string
	var isLiked, isDisliked, isFavorited, isUpvoted bool

	g.Go(func() error {
		u, _, e := s.userClient.User(ctx, topic.UserID)
		if e != nil {
			return e
		}
		if !userclient.IsRenderable(u) {
			authorBanned = true
			return nil
		}
		moe := 0
		if state, _ := s.stateRepo.FindByID(topic.UserID); state != nil {
			moe = state.Moemoepoint
		}
		author = &repository.TopicAuthorUser{
			ID: u.ID, Name: u.Name, Avatar: u.Avatar, Moemoepoint: moe,
		}
		return nil
	})
	g.Go(func() error {
		var e error
		sections, e = s.taxonomyRepo.FindSectionNamesByTopicID(topicID)
		return e
	})
	g.Go(func() error {
		miniApps = s.topicRepo.FindTopicMiniApps([]int{topicID})[topicID]
		return nil
	})

	if userInfo != nil {
		userID := userInfo.ID
		g.Go(func() error {
			isLiked, _ = s.topicRepo.HasUserLiked(userID, topicID)
			return nil
		})
		g.Go(func() error {
			isDisliked, _ = s.topicRepo.HasUserDisliked(userID, topicID)
			return nil
		})
		g.Go(func() error {
			isFavorited, _ = s.topicRepo.HasUserFavorited(userID, topicID)
			return nil
		})
		g.Go(func() error {
			isUpvoted, _ = s.topicRepo.HasUserUpvoted(userID, topicID)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, errors.ErrInternal("获取话题详情失败")
	}
	if authorBanned {
		return nil, errors.ErrNotFound("未找到该话题")
	}

	go func() { _ = s.topicRepo.IncrementView(topicID) }()

	if sections == nil {
		sections = []string{}
	}
	covers := []string(topic.CoverImages)
	if covers == nil {
		covers = []string{}
	}

	topicMentionNames := map[int]string{}
	if ids := markdown.ExtractMentionIDs(topic.Content); len(ids) > 0 {
		for id, u := range s.userClient.Hydrate(ctx, ids) {
			topicMentionNames[id] = u.Name
		}
	}

	detail := &dto.TopicDetail{
		AccessScope:    topic.AccessScope,
		AccessGrants:   topicDetailGrants(topic, userInfo, grants),
		ID:             topic.ID,
		Title:          topic.Title,
		Content:        topic.Content,
		ContentHtml:    markdown.ResolveMentionNames(markdown.Render(topic.Content), topicMentionNames),
		View:           topic.View,
		Status:         topic.Status,
		HiddenBy:       topic.HiddenBy,
		IsNSFW:         topic.IsNSFW,
		Category:       topic.Category,
		Sections:       sections,
		CoverImages:    covers,
		CoverImageMeta: markdown.ResolveContentImageMeta(covers),
		User: dto.KunUserWithMoemoepoint{
			ID:          author.ID,
			Name:        author.Name,
			Avatar:      author.Avatar,
			Moemoepoint: author.Moemoepoint,
		},
		LikeCount:        topic.LikeCount,
		IsLiked:          isLiked,
		DislikeCount:     topic.DislikeCount,
		IsDisliked:       isDisliked,
		FavoriteCount:    topic.FavoriteCount,
		IsFavorited:      isFavorited,
		UpvoteCount:      topic.UpvoteCount,
		IsUpvoted:        isUpvoted,
		ReplyCount:       topic.ReplyCount,
		MiniApps:         miniApps,
		StatusUpdateTime: topic.StatusUpdateTime,
		UpvoteTime:       topic.UpvoteTime,
		Edited:           topic.Edited,
		Created:          topic.CreatedAt,
	}

	viewerID := 0
	if userInfo != nil {
		viewerID = userInfo.ID
	}
	rrows, _ := s.topicRepo.GetTopicReactions(topicID)
	mineKeys, _ := s.topicRepo.GetUserTopicReactions(topicID, viewerID)
	detail.Reactions = buildReactionSummaries(
		rrows, mineKeys, s.userClient.Hydrate(ctx, reactionReactorIDs(rrows)))

	if topic.BestAnswerID != nil {
		reply, replyErr := s.topicRepo.FindReplyByID(*topic.BestAnswerID)
		if replyErr == nil && reply != nil {
			ru, _, _ := s.userClient.User(ctx, reply.UserID)
			if userclient.IsRenderable(ru) {
				baMentionNames := map[int]string{}
				if ids := markdown.ExtractMentionIDs(reply.Content); len(ids) > 0 {
					for id, u := range s.userClient.Hydrate(ctx, ids) {
						baMentionNames[id] = u.Name
					}
				}
				detail.BestAnswer = &dto.TopicBestAnswer{
					ID:              reply.ID,
					Floor:           reply.Floor,
					User:            dto.KunUser{ID: ru.ID, Name: ru.Name, Avatar: ru.Avatar},
					ContentMarkdown: reply.Content,
					ContentHtml:     markdown.ResolveMentionNames(markdown.Render(reply.Content), baMentionNames),
					Created:         reply.CreatedAt,
				}
			}
		}
	}

	return detail, nil
}
