package service

import (
	"context"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"kun-galgame-api/internal/galgame/client"
	galgameService "kun-galgame-api/internal/galgame/service"
	msgService "kun-galgame-api/internal/message/service"
	"kun-galgame-api/internal/moemoepoint"
	"kun-galgame-api/internal/user/dto"
	"kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/communityclient"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/role"
	"kun-galgame-api/pkg/userclient"

	"github.com/redis/go-redis/v9"
)

type UserService struct {
	stateRepo     *repository.StateRepository
	userStatsRepo *repository.UserStatsRepository
	rdb           *redis.Client
	galgameClient *client.GalgameClient
	galgameStats  *galgameService.GalgameUserStatsService
	userClient    *userclient.Client
	community     *communityclient.Client
	commentCache  *visiblePostsCache
}

func NewUserService(
	stateRepo *repository.StateRepository,
	userStatsRepo *repository.UserStatsRepository,
	rdb *redis.Client,
	galgameClient *client.GalgameClient,
	galgameStats *galgameService.GalgameUserStatsService,
	userClient *userclient.Client,
	community *communityclient.Client,
) *UserService {
	return &UserService{
		stateRepo:     stateRepo,
		userStatsRepo: userStatsRepo,
		rdb:           rdb,
		galgameClient: galgameClient,
		galgameStats:  galgameStats,
		userClient:    userClient,
		community:     community,
		commentCache:  newVisiblePostsCache(),
	}
}

func (s *UserService) GetUserProfile(ctx context.Context, userID int) (*dto.UserProfileDetail, *errors.AppError) {
	u, ok, err := s.userClient.User(ctx, userID)
	if err != nil {
		return nil, errors.ErrInternal("查询用户信息失败")
	}
	if !ok {
		return nil, errors.ErrNotFound("未找到该用户")
	}
	if u.Status != 0 {
		return &dto.UserProfileDetail{ID: u.ID, Name: u.Name, Status: u.Status}, nil
	}

	stats, err := s.userStatsRepo.GetUserStats(userID)
	if err != nil {
		return nil, errors.ErrInternal("获取用户统计失败")
	}
	state, _ := s.stateRepo.FindByID(userID)
	moe := 0
	if state != nil {
		moe = state.Moemoepoint
	}

	profile := &dto.UserProfileDetail{
		ID:          u.ID,
		Name:        u.Name,
		Avatar:      u.Avatar,
		Roles:       u.Roles,
		Status:      u.Status,
		Moemoepoint: moe,
		Bio:         u.Bio,
	}
	if t, perr := time.Parse(time.RFC3339, u.CreatedAt); perr == nil {
		profile.CreatedAt = t
	} else if state != nil {
		profile.CreatedAt = state.CreatedAt
	}

	profile.Topic = stats.Topic
	profile.TopicPoll = stats.TopicPoll
	profile.TopicLottery = stats.TopicLottery
	profile.ReplyCreated = stats.ReplyCreated
	profile.CommentCreated = stats.CommentCreated
	profile.GalgameComment = s.communityVisiblePosts(ctx, userID)
	profile.GalgameRating = stats.GalgameRating
	profile.GalgameResource = stats.GalgameResource
	profile.GalgameToolset = stats.GalgameToolset
	profile.GalgameToolsetResource = stats.GalgameToolsetResource
	profile.Upvote = stats.Upvote
	profile.Like = stats.Like
	profile.Dislike = stats.Dislike
	profile.DailyTopicCount = stats.DailyTopicCount

	galgameStats := s.galgameStats.Stats(ctx, int64(userID))
	profile.Galgame = galgameStats.Published
	profile.DailyGalgameCount = int64(galgameStats.PublishedToday)
	profile.ContributeGalgame = int64(galgameStats.Contributed)

	return profile, nil
}

func (s *UserService) CheckIn(ctx context.Context, userID int) (int, *errors.AppError) {
	applied, err := s.stateRepo.CheckIn(userID)
	if err != nil {
		return 0, errors.ErrInternal("签到失败")
	}
	if !applied {
		return 0, errors.ErrBadRequest("您今天已经签到过了")
	}

	points := rand.IntN(8)
	moemoepoint.Award(userID, points, moemoepoint.ReasonDailyCheckin, "",
		moemoepoint.Key("checkin", strconv.Itoa(userID), time.Now().Format("2006-01-02")))
	return points, nil
}

func (s *UserService) GetMoemoepointLog(
	ctx context.Context,
	userID, limit, beforeID int,
	reason string,
) (userclient.MoemoepointLogPage, *errors.AppError) {
	page, err := s.userClient.MoemoepointLog(ctx, userID, limit, beforeID, reason)
	if err != nil {
		return userclient.MoemoepointLogPage{}, errors.ErrInternal("获取萌萌点明细失败")
	}
	return page, nil
}

func (s *UserService) GetUserStatus(ctx context.Context, userID int) (*dto.UserStatusResponse, *errors.AppError) {
	moe := 0
	isCheckIn := false
	var uploadBytes int64
	var mutedTypes []string
	if state, err := s.stateRepo.FindByID(userID); err == nil && state != nil {
		moe = state.Moemoepoint
		isCheckIn = state.DailyCheckIn == 1
		uploadBytes = state.DailyToolsetUploadBytes
		mutedTypes = state.MutedNotificationTypes
	}

	localMuted, chatMuted := msgService.SplitMuted(mutedTypes)

	unreadMessage, _ := s.userStatsRepo.CountUnreadMessages(userID, localMuted)
	unreadSystem, _ := s.userStatsRepo.CountUnreadSystemMessages(userID)
	var unreadChat int64
	if !chatMuted {
		unreadChat, _ = s.userStatsRepo.CountUnreadChatMessages(userID)
	}

	isCreator := false
	if u, ok, uErr := s.userClient.User(ctx, userID); ok && uErr == nil {
		isCreator = role.IsCreator(u.Roles)
	}

	return &dto.UserStatusResponse{
		Moemoepoints:            moe,
		IsCheckIn:               isCheckIn,
		HasNewMessage:           (unreadMessage + unreadSystem + unreadChat) > 0,
		DailyToolsetUploadBytes: uploadBytes,
		IsCreator:               isCreator,
	}, nil
}

func (s *UserService) GetNotificationPreferences(userID int) (*dto.NotificationPreferenceResponse, *errors.AppError) {
	muted := []string{}
	if state, err := s.stateRepo.FindByID(userID); err == nil && state != nil && len(state.MutedNotificationTypes) > 0 {
		muted = msgService.SanitizeMutedKeys(state.MutedNotificationTypes)
	}
	return &dto.NotificationPreferenceResponse{MutedTypes: muted}, nil
}

func (s *UserService) UpdateNotificationPreferences(userID int, keys []string) (*dto.NotificationPreferenceResponse, *errors.AppError) {
	clean := msgService.SanitizeMutedKeys(keys)
	if err := s.stateRepo.Ensure(userID); err != nil {
		return nil, errors.ErrInternal("保存通知偏好失败")
	}
	if err := s.stateRepo.UpdateMutedTypes(userID, clean); err != nil {
		return nil, errors.ErrInternal("保存通知偏好失败")
	}
	return &dto.NotificationPreferenceResponse{MutedTypes: clean}, nil
}

func (s *UserService) GetFloatingCard(ctx context.Context, userID int) (*dto.FloatingCardResponse, *errors.AppError) {
	u, ok, err := s.userClient.User(ctx, userID)
	if err != nil {
		return nil, errors.ErrInternal("查询用户信息失败")
	}
	if !ok || u.Status != 0 {
		return nil, errors.ErrNotFound("未找到该用户")
	}

	state, _ := s.stateRepo.FindByID(userID)
	moe := 0
	if state != nil {
		moe = state.Moemoepoint
	}

	stats := s.userStatsRepo.FindFloatingStats(userID)
	commentCount := stats.TopicCommentCount + s.communityVisiblePosts(ctx, userID)
	return &dto.FloatingCardResponse{
		ID:                   u.ID,
		Name:                 u.Name,
		Avatar:               u.Avatar,
		Moemoepoint:          moe,
		TopicCount:           stats.TopicCount,
		TopicReplyCount:      stats.TopicReplyCount,
		TopicCommentCount:    commentCount,
		GalgameResourceCount: stats.ResourceCount,
	}, nil
}

func (s *UserService) SearchMentionUsers(ctx context.Context, q string, limit int) ([]dto.MentionUser, *errors.AppError) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []dto.MentionUser{}, nil
	}
	if limit <= 0 || limit > 20 {
		limit = 8
	}
	users, err := s.userClient.SearchUsers(ctx, q, limit)
	if err != nil {
		return nil, errors.ErrInternal("搜索用户失败")
	}
	out := make([]dto.MentionUser, 0, len(users))
	for _, u := range users {
		if u.Status != 0 {
			continue
		}
		out = append(out, dto.MentionUser{ID: u.ID, Name: u.Name, Avatar: u.Avatar})
	}
	return out, nil
}
