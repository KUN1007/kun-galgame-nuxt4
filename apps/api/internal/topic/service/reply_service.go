package service

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"kun-galgame-api/internal/constants"
	"kun-galgame-api/internal/infrastructure/markdown"
	msgService "kun-galgame-api/internal/message/service"
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/moemoepoint"
	"kun-galgame-api/internal/topic/dto"
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/internal/topic/repository"
	"kun-galgame-api/internal/trust/gate"
	userRepo "kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ReplyService struct {
	replyRepo   *repository.ReplyRepository
	commentRepo *repository.CommentRepository
	topicRepo   *repository.TopicRepository
	stateRepo   *userRepo.StateRepository
	userClient  *userclient.Client
	rdb         *redis.Client
	check       *gate.CheckService
	scan        *gate.ScanService
	helpers     InteractionHelpers
}

func NewReplyService(
	replyRepo *repository.ReplyRepository,
	commentRepo *repository.CommentRepository,
	topicRepo *repository.TopicRepository,
	stateRepo *userRepo.StateRepository,
	userClient *userclient.Client,
	rdb *redis.Client,
	check *gate.CheckService,
	scan *gate.ScanService,
) *ReplyService {
	return &ReplyService{
		replyRepo:   replyRepo,
		commentRepo: commentRepo,
		topicRepo:   topicRepo,
		stateRepo:   stateRepo,
		userClient:  userClient,
		rdb:         rdb,
		check:       check,
		scan:        scan,
	}
}

func (s *ReplyService) LocateReply(topicID, floor, commentID, limit int, userInfo *middleware.UserInfo) (*dto.ReplyLocateResponse, *errors.AppError) {
	topic, err := s.topicRepo.FindByID(topicID)
	if err != nil {
		return nil, errors.ErrNotFound("未找到该话题")
	}
	if _, appErr := requireTopicRead(s.topicRepo, topic, userInfo); appErr != nil {
		return nil, appErr
	}
	replyID := 0
	if commentID > 0 {
		f, rid, ok, err := s.replyRepo.FindReplyFloorByCommentID(topicID, commentID)
		if err != nil {
			return nil, errors.ErrInternal("定位评论失败")
		}
		if !ok {
			return nil, errors.ErrNotFound("评论不存在或已删除")
		}
		floor, replyID = f, rid
	}
	if floor <= 0 {
		return nil, errors.ErrBadRequest("缺少 reply 或 comment 参数")
	}
	page, err := s.replyRepo.LocateReplyPageByFloor(topicID, floor, limit)
	if err != nil {
		return nil, errors.ErrInternal("定位回复失败")
	}
	return &dto.ReplyLocateResponse{
		Page:      page,
		Floor:     floor,
		ReplyID:   replyID,
		CommentID: commentID,
	}, nil
}

func (s *ReplyService) GetReplies(
	ctx context.Context,
	req *dto.ListRepliesRequest,
	userInfo *middleware.UserInfo,
) ([]dto.TopicReplyResponse, *errors.AppError) {
	topic, err := s.topicRepo.FindByID(req.TopicID)
	if err != nil {
		return []dto.TopicReplyResponse{}, nil
	}
	if _, appErr := requireTopicRead(s.topicRepo, topic, userInfo); appErr != nil {
		return nil, appErr
	}

	var specialIDs []int
	if topic.PinnedReplyID != nil {
		specialIDs = append(specialIDs, *topic.PinnedReplyID)
	}
	if topic.BestAnswerID != nil && (topic.PinnedReplyID == nil || *topic.BestAnswerID != *topic.PinnedReplyID) {
		specialIDs = append(specialIDs, *topic.BestAnswerID)
	}

	var result []dto.TopicReplyResponse

	if req.Page == 1 && len(specialIDs) > 0 {
		if specialRows, err := s.replyRepo.FindRepliesByIDs(specialIDs); err == nil {
			result = append(result, s.buildReplyResponses(ctx, specialRows, topic, userInfo)...)
		}
	}

	regularRows, err := s.replyRepo.FindRepliesPaginated(
		req.TopicID, specialIDs,
		req.Page, req.Limit, req.SortOrder,
	)
	if err != nil {
		return nil, errors.ErrInternal("获取回复列表失败")
	}

	result = append(result, s.buildReplyResponses(ctx, regularRows, topic, userInfo)...)

	if result == nil {
		result = []dto.TopicReplyResponse{}
	}
	return result, nil
}

func (s *ReplyService) GetReplyDetail(
	ctx context.Context,
	replyID int,
	userInfo *middleware.UserInfo,
) (*dto.TopicReplyResponse, *errors.AppError) {
	rows, err := s.replyRepo.FindRepliesByIDs([]int{replyID})
	if err != nil || len(rows) == 0 {
		return nil, errors.ErrNotFound("未找到该回复")
	}

	topic, topicErr := s.topicRepo.FindByID(rows[0].TopicID)
	if topicErr != nil {
		return nil, errors.ErrNotFound("未找到该回复")
	}
	if _, appErr := requireTopicRead(s.topicRepo, topic, userInfo); appErr != nil {
		return nil, appErr
	}
	responses := s.buildReplyResponses(ctx, rows, topic, userInfo)
	if len(responses) == 0 {
		return nil, errors.ErrNotFound("未找到该回复")
	}
	return &responses[0], nil
}

func (s *ReplyService) CreateReply(
	ctx context.Context,
	user *middleware.UserInfo,
	req *dto.CreateReplyRequest,
) (*dto.TopicReplyResponse, *errors.AppError) {
	userID := user.ID
	topic, err := s.topicRepo.FindByID(req.TopicID)
	if err != nil {
		return nil, errors.ErrNotFound("未找到该话题")
	}

	if _, appErr := requireTopicRead(s.topicRepo, topic, user); appErr != nil {
		return nil, appErr
	}

	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.ErrBadRequest("回复内容不能为空")
	}

	authorID := int64(userID)
	decision, matched := s.check.Decision(ctx, req.Content, &authorID)
	if decision == gate.DecisionDeny {
		return nil, errContentBlocked()
	}

	var newReply *topicModel.TopicReply

	txErr := s.replyRepo.DB().Transaction(func(tx *gorm.DB) error {
		maxFloor, err := s.replyRepo.GetMaxFloor(tx, req.TopicID)
		if err != nil {
			return err
		}

		newReply = &topicModel.TopicReply{
			UserID:  userID,
			TopicID: req.TopicID,
			Floor:   maxFloor + 1,
			Content: req.Content,
		}
		if err := s.replyRepo.CreateReply(tx, newReply); err != nil {
			return err
		}

		if err := s.topicRepo.TouchStatusUpdateTime(tx, req.TopicID, time.Now()); err != nil {
			return err
		}

		if err := recomputeTopicCounts(tx, req.TopicID); err != nil {
			return err
		}

		preview := truncate(strings.TrimSpace(req.Content), constants.TextPreviewLength)

		if topic.UserID != userID {
			s.helpers.AdjustMoemoepoint(tx, topic.UserID, constants.RewardReply,
				moemoepoint.ReasonContentApproved, moemoepoint.Ref("topic", req.TopicID))
			if err := s.helpers.CreateReplyMessage(tx, userID, topic.UserID, "replied", preview, req.TopicID, newReply.Floor, 0); err != nil {
				return err
			}
		}

		return s.helpers.NotifyMentions(tx, userID, req.TopicID, newReply.Floor, 0, req.Content)
	})

	if txErr != nil {
		return nil, errors.ErrInternal("创建回复失败")
	}

	if decision == gate.DecisionHold {
		slog.Info("trust check hold", "subject_kind", gate.SubjectKindReply, "subject_id", newReply.ID, "author_id", userID, "matched", matched)
	}
	s.scan.ScanBg(gate.SubjectKindReply, strconv.Itoa(newReply.ID), req.Content, int64(userID))

	rows, _ := s.replyRepo.FindRepliesByIDs([]int{newReply.ID})
	if len(rows) == 0 {
		return nil, errors.ErrInternal("创建回复失败")
	}
	responses := s.buildReplyResponses(ctx, rows, topic, nil)
	return &responses[0], nil
}

func (s *ReplyService) UpdateReply(
	ctx context.Context,
	userID int,
	canEditAny bool,
	req *dto.UpdateReplyRequest,
) *errors.AppError {
	reply, err := s.replyRepo.FindByID(req.ReplyID)
	if err != nil {
		return errors.ErrNotFound("未找到该回复")
	}
	if reply.UserID != userID && !canEditAny {
		return errors.ErrForbidden("您没有权限编辑此回复")
	}

	if strings.TrimSpace(req.Content) == "" {
		return errors.ErrBadRequest("回复内容不能为空")
	}

	authorID := int64(reply.UserID)
	decision, matched := s.check.Decision(ctx, req.Content, &authorID)
	if decision == gate.DecisionDeny {
		return errContentBlocked()
	}

	now := time.Now()
	txErr := s.replyRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.replyRepo.UpdateReplyContent(tx, req.ReplyID, map[string]any{
			"content": req.Content,
			"edited":  &now,
		}); err != nil {
			return err
		}

		return s.helpers.NotifyMentions(tx, reply.UserID, reply.TopicID, reply.Floor, 0, req.Content)
	})

	if txErr != nil {
		return errors.ErrInternal("更新回复失败")
	}

	if decision == gate.DecisionHold {
		slog.Info("trust check hold", "subject_kind", gate.SubjectKindReply, "subject_id", req.ReplyID, "author_id", reply.UserID, "matched", matched)
	}
	s.scan.ScanBg(gate.SubjectKindReply, strconv.Itoa(req.ReplyID), req.Content, int64(reply.UserID))

	return nil
}

func (s *ReplyService) DeleteReply(
	ctx context.Context,
	userID int,
	canModerate bool,
	replyID int,
) *errors.AppError {
	reply, err := s.replyRepo.FindByID(replyID)
	if err != nil {
		return errors.ErrNotFound("未找到该回复")
	}
	if reply.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限删除此回复")
	}

	commentCount, likeCount, _ := s.replyRepo.CountReplyRelated(replyID)

	penalty := 3
	if reply.UserID == userID && !canModerate {
		penalty = 3 * int(commentCount+likeCount+1)
	}

	txErr := s.replyRepo.DB().Transaction(func(tx *gorm.DB) error {
		state, err := s.stateRepo.LockForUpdate(tx, reply.UserID)
		if err != nil {
			return err
		}
		if state.Moemoepoint < penalty {
			return gorm.ErrCheckConstraintViolated
		}

		if err := s.replyRepo.DeleteRepliesByIDs(tx, []int{replyID}); err != nil {
			return err
		}

		if err := recomputeTopicCounts(tx, reply.TopicID); err != nil {
			return err
		}

		s.helpers.AdjustMoemoepoint(tx, reply.UserID, -penalty,
			moemoepoint.ReasonContentRemoved, moemoepoint.Ref("topic_reply", replyID))
		return nil
	})

	if txErr == gorm.ErrCheckConstraintViolated {
		return errors.ErrBadRequest("萌萌点不足, 无法删除此回复")
	}
	if txErr != nil {
		return errors.ErrInternal("删除回复失败")
	}
	return nil
}

func (s *ReplyService) ModerationRemove(replyID int) error {
	reply, err := s.replyRepo.FindByID(replyID)
	if err != nil {
		return nil
	}
	return s.replyRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.replyRepo.DeleteRepliesByIDs(tx, []int{replyID}); err != nil {
			return err
		}
		return recomputeTopicCounts(tx, reply.TopicID)
	})
}

func (s *ReplyService) ToggleReplyLike(ctx context.Context, userID, replyID int) *errors.AppError {
	return s.ToggleReplyReaction(ctx, userID, replyID, "like")
}

func (s *ReplyService) ToggleReplyDislike(ctx context.Context, userID, replyID int) *errors.AppError {
	return s.ToggleReplyReaction(ctx, userID, replyID, "dislike")
}

func (s *ReplyService) ToggleReplyReaction(ctx context.Context, userID, replyID int, reaction string) *errors.AppError {
	if !reactionKeys[reaction] {
		return errors.ErrBadRequest("无效的 reaction")
	}
	err := s.replyRepo.DB().Transaction(func(tx *gorm.DB) error {
		reply, err := s.replyRepo.FindByIDTx(tx, replyID)
		if err != nil {
			return err
		}
		if reaction == "like" && reply.UserID == userID {
			return gorm.ErrInvalidData
		}

		has, err := s.replyRepo.HasReplyReaction(tx, replyID, userID, reaction)
		if err != nil {
			return err
		}
		if has {
			if err := s.replyRepo.RemoveReplyReaction(tx, replyID, userID, reaction); err != nil {
				return err
			}
			switch reaction {
			case "like":
				if err := s.replyRepo.AdjustReplyLikeCount(tx, replyID, -1); err != nil {
					return err
				}
				s.helpers.AdjustMoemoepoint(tx, reply.UserID, -1,
					moemoepoint.ReasonLiked, moemoepoint.Ref("topic_reply", replyID))
			case "dislike":
				if err := s.replyRepo.AdjustReplyDislikeCount(tx, replyID, -1); err != nil {
					return err
				}
			}
			return nil
		}

		switch reaction {
		case "like":
			if err := s.clearReplyReaction(tx, replyID, userID, reply.UserID, "dislike"); err != nil {
				return err
			}
			if err := s.replyRepo.AddReplyReaction(tx, replyID, userID, "like"); err != nil {
				return err
			}
			if err := s.replyRepo.AdjustReplyLikeCount(tx, replyID, 1); err != nil {
				return err
			}
			s.helpers.AdjustMoemoepoint(tx, reply.UserID, 1,
				moemoepoint.ReasonLiked, moemoepoint.Ref("topic_reply", replyID))
			link := msgService.BuildTopicLink(reply.TopicID, reply.Floor, 0)
			if err := createDedupMessage(tx, userID, reply.UserID, "liked",
				truncate(reply.Content, constants.TextPreviewLength), link); err != nil {
				return err
			}
		case "dislike":
			if err := s.clearReplyReaction(tx, replyID, userID, reply.UserID, "like"); err != nil {
				return err
			}
			if err := s.replyRepo.AddReplyReaction(tx, replyID, userID, "dislike"); err != nil {
				return err
			}
			if err := s.replyRepo.AdjustReplyDislikeCount(tx, replyID, 1); err != nil {
				return err
			}
		default:
			if err := s.replyRepo.AddReplyReaction(tx, replyID, userID, reaction); err != nil {
				return err
			}
		}
		return nil
	})

	switch err {
	case nil:
		return nil
	case gorm.ErrInvalidData:
		return errors.ErrBadRequest("您不能给自己的回复点赞")
	default:
		return errors.ErrInternal("操作失败")
	}
}

func (s *ReplyService) GetReplyReactionHistory(ctx context.Context, replyID int) ([]dto.ReactionHistoryItem, *errors.AppError) {
	rows, err := s.replyRepo.GetReplyReactionHistory(replyID, topicReactionHistoryLimit)
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

func (s *ReplyService) clearReplyReaction(tx *gorm.DB, replyID, userID, ownerID int, reaction string) error {
	has, err := s.replyRepo.HasReplyReaction(tx, replyID, userID, reaction)
	if err != nil || !has {
		return err
	}
	if err := s.replyRepo.RemoveReplyReaction(tx, replyID, userID, reaction); err != nil {
		return err
	}
	switch reaction {
	case "like":
		if err := s.replyRepo.AdjustReplyLikeCount(tx, replyID, -1); err != nil {
			return err
		}
		s.helpers.AdjustMoemoepoint(tx, ownerID, -1,
			moemoepoint.ReasonLiked, moemoepoint.Ref("topic_reply", replyID))
	case "dislike":
		if err := s.replyRepo.AdjustReplyDislikeCount(tx, replyID, -1); err != nil {
			return err
		}
	}
	return nil
}

func (s *ReplyService) PinReply(ctx context.Context, userID int, canModerate bool, topicID, replyID int) *errors.AppError {
	topic, err := s.topicRepo.FindByID(topicID)
	if err != nil {
		return errors.ErrNotFound("未找到该话题")
	}
	if topic.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限置顶回复")
	}

	reply, err := s.replyRepo.FindByID(replyID)
	if err != nil {
		return errors.ErrNotFound("未找到该回复")
	}

	isPinning := topic.PinnedReplyID == nil || *topic.PinnedReplyID != replyID
	var newPinned *int
	if isPinning {
		newPinned = &replyID
	}

	txErr := s.replyRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&topicModel.Topic{}).Where("id = ?", topicID).
			Updates(map[string]any{"pinned_reply_id": newPinned}).Error; err != nil {
			return err
		}
		if isPinning && userID != reply.UserID {
			return s.helpers.CreateTopicMessageWithContent(
				tx, userID, reply.UserID, "pin-reply",
				replyPlainPreview(*reply),
				topicID, reply.Floor, 0,
			)
		}
		return nil
	})
	if txErr != nil {
		return errors.ErrInternal("操作失败")
	}
	return nil
}

func replyPlainPreview(reply topicModel.TopicReply) string {
	return markdown.ToPlainText(reply.Content, 500)
}
