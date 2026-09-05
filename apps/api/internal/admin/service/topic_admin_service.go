package service

import (
	"context"
	"errors"
	"kun-galgame-api/internal/admin/dto"
	"kun-galgame-api/internal/admin/repository"
	apperrors "kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"
	"log/slog"

	"gorm.io/gorm"
)

type TopicAdminService struct {
	repo       *repository.TopicAdminRepository
	userClient *userclient.Client
}

func NewTopicAdminService(repo *repository.TopicAdminRepository, uc *userclient.Client) *TopicAdminService {
	return &TopicAdminService{repo: repo, userClient: uc}
}
func (s *TopicAdminService) ListHidden(ctx context.Context, page, limit int, hiddenBy, keywords string) (dto.HiddenTopicList, *apperrors.AppError) {
	rows, total, err := s.repo.ListHidden(page, limit, hiddenBy, keywords)
	if err != nil {
		return dto.HiddenTopicList{}, apperrors.ErrInternal("获取隐藏话题失败")
	}
	ids := make([]int, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.UserID)
	}
	users := s.userClient.Hydrate(ctx, ids)
	out := make([]dto.HiddenTopic, 0, len(rows))
	for _, r := range rows {
		u := users[r.UserID]
		out = append(out, dto.HiddenTopic{ID: r.ID, Title: r.Title, HiddenBy: r.HiddenBy, ReplyCount: r.ReplyCount, StatusUpdateTime: r.StatusUpdateTime, Created: r.Created, User: dto.TopicAdminUser{ID: u.ID, Name: u.Name, Avatar: u.Avatar}})
	}
	return dto.HiddenTopicList{Topics: out, Total: total}, nil
}
func (s *TopicAdminService) PurgeStats(ctx context.Context, id int) (dto.TopicPurgeStats, *apperrors.AppError) {
	st, err := s.repo.PurgeStats(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return st, apperrors.ErrNotFound("未找到该话题")
		}
		return st, apperrors.ErrInternal("获取话题清除统计失败")
	}
	u := s.userClient.Hydrate(ctx, []int{st.User.ID})[st.User.ID]
	st.User.Name = u.Name
	st.User.Avatar = u.Avatar
	return st, nil
}
func (s *TopicAdminService) Delete(ctx context.Context, operatorID, id int) (dto.TopicPurgeStats, *apperrors.AppError) {
	st, err := s.repo.Delete(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return st, apperrors.ErrNotFound("未找到该话题")
		}
		return st, apperrors.ErrInternal("删除话题失败")
	}
	u := s.userClient.Hydrate(ctx, []int{st.User.ID})[st.User.ID]
	st.User.Name = u.Name
	st.User.Avatar = u.Avatar
	slog.Info("admin topic purged", "operator_id", operatorID, "topic_id", id, "title", st.Title, "author_id", st.User.ID, "replies", st.Replies, "comments", st.Comments, "polls", st.Polls, "lotteries", st.Lotteries, "drawn_lotteries", st.DrawnLotteries, "favorites", st.Favorites)
	return st, nil
}
