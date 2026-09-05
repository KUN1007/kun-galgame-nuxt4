package service

import (
	"context"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"kun-galgame-api/internal/constants"
	"kun-galgame-api/internal/infrastructure/markdown"
	msgService "kun-galgame-api/internal/message/service"
	"kun-galgame-api/internal/moemoepoint"
	"kun-galgame-api/internal/topic/dto"
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/internal/topic/repository"
	"kun-galgame-api/internal/trust/gate"
	userRepo "kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type TopicWriteService struct {
	topicRepo    *repository.TopicRepository
	taxonomyRepo *repository.TopicTaxonomyRepository
	replyRepo    *repository.ReplyRepository
	stateRepo    *userRepo.StateRepository
	rdb          *redis.Client
	notifier     msgService.Notifier
	check        *gate.CheckService
	scan         *gate.ScanService
	helpers      InteractionHelpers
}

func NewTopicWriteService(
	topicRepo *repository.TopicRepository,
	taxonomyRepo *repository.TopicTaxonomyRepository,
	replyRepo *repository.ReplyRepository,
	stateRepo *userRepo.StateRepository,
	rdb *redis.Client,
	notifier msgService.Notifier,
	check *gate.CheckService,
	scan *gate.ScanService,
) *TopicWriteService {
	return &TopicWriteService{
		topicRepo:    topicRepo,
		taxonomyRepo: taxonomyRepo,
		replyRepo:    replyRepo,
		stateRepo:    stateRepo,
		rdb:          rdb,
		notifier:     notifier,
		check:        check,
		scan:         scan,
	}
}

func topicModerationText(title, content string) string {
	return title + "\n\n" + content
}

func errContentBlocked() *errors.AppError {
	return gate.ErrContentBlocked()
}

func anyConsumeSection(sections []string) bool {
	for _, sec := range sections {
		if constants.TopicSectionConsume[sec] {
			return true
		}
	}
	return false
}

func topicSectionFootprint(consume bool) int {
	if consume {
		return -constants.CostConsumeSection
	}
	return constants.RewardCreateTopic
}

var coverImageTokenRe = regexp.MustCompile(`^/image/[0-9a-f]{64}$`)

func normalizeCoverImages(in []string) (topicModel.ImageTokens, *errors.AppError) {
	if len(in) == 0 {
		return nil, nil
	}
	if len(in) > 9 {
		return nil, errors.ErrBadRequest("封面图最多 9 张")
	}
	seen := make(map[string]struct{}, len(in))
	out := make(topicModel.ImageTokens, 0, len(in))
	for _, tk := range in {
		if !coverImageTokenRe.MatchString(tk) {
			return nil, errors.ErrBadRequest("封面图格式不正确")
		}
		if _, dup := seen[tk]; dup {
			continue
		}
		seen[tk] = struct{}{}
		out = append(out, tk)
	}
	return out, nil
}

func (s *TopicWriteService) Create(
	ctx context.Context,
	userID int,
	req *dto.CreateTopicRequest,
) (int, *errors.AppError) {
	hasConsumeSection := anyConsumeSection(req.Sections)

	coverInput := req.CoverImages
	if len(coverInput) == 0 {
		coverInput = markdown.ExtractContentImages(req.Content, 9)
	}
	covers, coverErr := normalizeCoverImages(coverInput)
	if coverErr != nil {
		return 0, coverErr
	}

	moderationText := topicModerationText(req.Title, req.Content)
	authorID := int64(userID)
	decision, matched := s.check.Decision(ctx, moderationText, &authorID)
	if decision == gate.DecisionDeny {
		return 0, errContentBlocked()
	}

	var newTopicID int

	err := s.topicRepo.DB().Transaction(func(tx *gorm.DB) error {
		state, err := s.stateRepo.LockForUpdate(tx, userID)
		if err != nil {
			return err
		}

		todayCount, err := s.topicRepo.CountTodayTopicsByUser(tx, userID)
		if err != nil {
			return err
		}
		dailyLimit := int64(state.Moemoepoint/10 + 1)
		if todayCount >= dailyLimit {
			return gorm.ErrInvalidData
		}

		if hasConsumeSection && state.Moemoepoint < constants.CostConsumeSection {
			return gorm.ErrInvalidData
		}

		topic := &topicModel.Topic{
			Title:       req.Title,
			Content:     req.Content,
			Category:    req.Category,
			IsNSFW:      req.IsNSFW,
			UserID:      userID,
			CoverImages: covers,
		}
		if err := s.topicRepo.CreateTopic(tx, topic); err != nil {
			return err
		}
		newTopicID = topic.ID

		sections, err := s.taxonomyRepo.FindSectionsByNamesTx(tx, req.Sections)
		if err != nil {
			return err
		}
		for _, sec := range sections {
			if err := s.taxonomyRepo.CreateSectionRelation(tx, topic.ID, sec.ID); err != nil {
				return err
			}
		}

		pointsDelta := topicSectionFootprint(hasConsumeSection)
		mpReason := moemoepoint.ReasonContentApproved
		if hasConsumeSection {
			mpReason = moemoepoint.ReasonContentRemoved
		}
		s.helpers.AdjustMoemoepoint(tx, userID, pointsDelta, mpReason, moemoepoint.Ref("topic", topic.ID))
		return s.helpers.NotifyMentions(tx, userID, topic.ID, 0, 0, req.Content)
	})

	if err != nil {
		if err == gorm.ErrInvalidData {
			if hasConsumeSection {
				return 0, errors.ErrBadRequest("您的萌萌点不足, 无法发布此类型话题")
			}
			return 0, errors.ErrBadRequest("您今日发布的话题已达上限")
		}
		return 0, errors.ErrInternal("创建话题失败")
	}

	if decision == gate.DecisionHold {
		slog.Info("trust check hold", "subject_kind", gate.SubjectKindTopic, "subject_id", newTopicID, "author_id", userID, "matched", matched)
	}
	s.scan.ScanBg(gate.SubjectKindTopic, strconv.Itoa(newTopicID), moderationText, int64(userID))

	return newTopicID, nil
}

func (s *TopicWriteService) Update(
	ctx context.Context,
	userID int,
	canModerate bool,
	topicID int,
	req *dto.UpdateTopicRequest,
) *errors.AppError {
	topic, err := s.topicRepo.FindByID(topicID)
	if err != nil {
		return errors.ErrNotFound("未找到该话题")
	}
	if topic.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限编辑此话题")
	}

	oldSections, err := s.taxonomyRepo.FindSectionNamesByTopicID(topicID)
	if err != nil {
		return errors.ErrInternal("更新话题失败")
	}
	oldConsume := anyConsumeSection(oldSections)
	newConsume := anyConsumeSection(req.Sections)

	covers, coverErr := normalizeCoverImages(req.CoverImages)
	if coverErr != nil {
		return coverErr
	}

	moderationText := topicModerationText(req.Title, req.Content)
	authorID := int64(topic.UserID)
	decision, matched := s.check.Decision(ctx, moderationText, &authorID)
	if decision == gate.DecisionDeny {
		return errContentBlocked()
	}

	now := time.Now()
	txErr := s.topicRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.topicRepo.UpdateTopicFields(tx, topicID, map[string]any{
			"title":              req.Title,
			"content":            req.Content,
			"category":           req.Category,
			"is_nsfw":            req.IsNSFW,
			"cover_images":       covers,
			"edited":             &now,
			"status_update_time": now,
		}); err != nil {
			return err
		}

		sections, err := s.taxonomyRepo.FindSectionsByNamesTx(tx, req.Sections)
		if err != nil {
			return err
		}
		sectionIDs := make([]int, len(sections))
		for i, sec := range sections {
			sectionIDs[i] = sec.ID
		}
		if err := s.taxonomyRepo.ReplaceSectionRelations(tx, topicID, sectionIDs); err != nil {
			return err
		}

		if delta := topicSectionFootprint(newConsume) - topicSectionFootprint(oldConsume); delta != 0 {
			mpReason := moemoepoint.ReasonContentApproved
			if delta < 0 {
				mpReason = moemoepoint.ReasonContentRemoved
			}
			s.helpers.AdjustMoemoepoint(tx, topic.UserID, delta, mpReason, moemoepoint.Ref("topic", topicID))
		}

		return s.helpers.NotifyMentions(tx, userID, topicID, 0, 0, req.Content)
	})

	if txErr != nil {
		return errors.ErrInternal("更新话题失败")
	}

	if decision == gate.DecisionHold {
		slog.Info("trust check hold", "subject_kind", gate.SubjectKindTopic, "subject_id", topicID, "author_id", topic.UserID, "matched", matched)
	}
	s.scan.ScanBg(gate.SubjectKindTopic, strconv.Itoa(topicID), moderationText, int64(topic.UserID))

	return nil
}

var reactionKeys = map[string]bool{
	"like": true, "dislike": true,
	"heart": true, "fire": true, "party": true, "love": true,
	"clap": true, "thinking": true, "mindblown": true, "scream": true,
	"cry": true, "pray": true, "eyes": true, "hundred": true,
	"partyface": true, "starstruck": true,
	"angry": true, "anxious": true, "banana": true, "eyebrow": true,
	"voltage": true, "hotdog": true, "hot": true, "sob": true,
	"moai": true, "newmoon": true, "police": true, "pouting": true,
	"salute": true, "shrimp": true, "halo": true, "sunglasses": true,
	"whale": true,
}

func (s *TopicWriteService) ToggleLike(ctx context.Context, userID, topicID int) *errors.AppError {
	return s.ToggleReaction(ctx, userID, topicID, "like")
}

func (s *TopicWriteService) ToggleDislike(ctx context.Context, userID, topicID int) *errors.AppError {
	return s.ToggleReaction(ctx, userID, topicID, "dislike")
}

func (s *TopicWriteService) ToggleReaction(ctx context.Context, userID, topicID int, reaction string) *errors.AppError {
	if !reactionKeys[reaction] {
		return errors.ErrBadRequest("无效的 reaction")
	}
	err := s.topicRepo.DB().Transaction(func(tx *gorm.DB) error {
		topic, err := s.topicRepo.FindByIDTx(tx, topicID)
		if err != nil {
			return err
		}
		if topic.Status == 1 {
			return gorm.ErrRecordNotFound
		}
		if reaction == "like" && topic.UserID == userID {
			return gorm.ErrInvalidData
		}

		has, err := s.topicRepo.HasReaction(tx, topicID, userID, reaction)
		if err != nil {
			return err
		}
		if has {
			if err := s.topicRepo.RemoveReaction(tx, topicID, userID, reaction); err != nil {
				return err
			}
			switch reaction {
			case "like":
				if err := s.topicRepo.AdjustLikeCount(tx, topicID, -1); err != nil {
					return err
				}
				s.helpers.AdjustMoemoepoint(tx, topic.UserID, -1,
					moemoepoint.ReasonLiked, moemoepoint.Ref("topic", topicID))
			case "dislike":
				if err := s.topicRepo.AdjustDislikeCount(tx, topicID, -1); err != nil {
					return err
				}
			}
			return nil
		}

		switch reaction {
		case "like":
			if err := s.clearTopicReaction(tx, topicID, userID, "dislike", topic.UserID); err != nil {
				return err
			}
			if err := s.topicRepo.AddReaction(tx, topicID, userID, "like"); err != nil {
				return err
			}
			if err := s.topicRepo.AdjustLikeCount(tx, topicID, 1); err != nil {
				return err
			}
			s.helpers.AdjustMoemoepoint(tx, topic.UserID, 1,
				moemoepoint.ReasonLiked, moemoepoint.Ref("topic", topicID))
			if err := s.helpers.CreateTopicMessageWithContent(tx, userID, topic.UserID, "liked",
				truncate(topic.Title, constants.TextPreviewLength), topicID, 0, 0); err != nil {
				return err
			}
		case "dislike":
			if err := s.clearTopicReaction(tx, topicID, userID, "like", topic.UserID); err != nil {
				return err
			}
			if err := s.topicRepo.AddReaction(tx, topicID, userID, "dislike"); err != nil {
				return err
			}
			if err := s.topicRepo.AdjustDislikeCount(tx, topicID, 1); err != nil {
				return err
			}
		default:
			if err := s.topicRepo.AddReaction(tx, topicID, userID, reaction); err != nil {
				return err
			}
		}
		return nil
	})

	switch err {
	case nil:
		return nil
	case gorm.ErrRecordNotFound:
		return errors.ErrNotFound("未找到该话题")
	case gorm.ErrInvalidData:
		return errors.ErrBadRequest("您不能给自己点赞")
	default:
		return errors.ErrInternal("操作失败")
	}
}

func (s *TopicWriteService) clearTopicReaction(tx *gorm.DB, topicID, userID int, reaction string, ownerID int) error {
	has, err := s.topicRepo.HasReaction(tx, topicID, userID, reaction)
	if err != nil || !has {
		return err
	}
	if err := s.topicRepo.RemoveReaction(tx, topicID, userID, reaction); err != nil {
		return err
	}
	switch reaction {
	case "like":
		if err := s.topicRepo.AdjustLikeCount(tx, topicID, -1); err != nil {
			return err
		}
		s.helpers.AdjustMoemoepoint(tx, ownerID, -1,
			moemoepoint.ReasonLiked, moemoepoint.Ref("topic", topicID))
	case "dislike":
		if err := s.topicRepo.AdjustDislikeCount(tx, topicID, -1); err != nil {
			return err
		}
	}
	return nil
}

func (s *TopicWriteService) Upvote(ctx context.Context, userID, topicID int, description string) *errors.AppError {
	description = truncate(description, 30)
	err := s.topicRepo.DB().Transaction(func(tx *gorm.DB) error {
		topic, err := s.topicRepo.FindByIDTx(tx, topicID)
		if err != nil {
			return err
		}
		if topic.Status == 1 {
			return gorm.ErrRecordNotFound
		}
		if topic.UserID == userID {
			return gorm.ErrInvalidData
		}

		state, err := s.stateRepo.LockForUpdate(tx, userID)
		if err != nil {
			return err
		}
		if state.Moemoepoint < constants.CostUpvoteSender {
			return gorm.ErrCheckConstraintViolated
		}

		now := time.Now()

		if err := s.topicRepo.CreateTopicUpvote(tx, userID, topicID, description); err != nil {
			return err
		}
		if err := s.topicRepo.ApplyUpvoteCountAndTime(tx, topicID, now); err != nil {
			return err
		}

		s.helpers.AdjustMoemoepoint(tx, userID, -constants.CostUpvoteSender,
			moemoepoint.ReasonContentRemoved, moemoepoint.Ref("topic_upvote", topicID))
		s.helpers.AdjustMoemoepoint(tx, topic.UserID, constants.RewardUpvoteOwner,
			moemoepoint.ReasonContentApproved, moemoepoint.Ref("topic_upvote", topicID))
		return s.helpers.CreateTopicMessageWithContent(tx, userID, topic.UserID, "upvoted",
			truncate(topic.Title, constants.TextPreviewLength), topicID, 0, 0)
	})

	if err == gorm.ErrRecordNotFound {
		return errors.ErrNotFound("未找到该话题")
	}
	if err == gorm.ErrInvalidData {
		return errors.ErrBadRequest("您不能推自己的话题")
	}
	if err == gorm.ErrCheckConstraintViolated {
		return errors.ErrBadRequest("萌萌点不足, 推话题需要 10 萌萌点")
	}
	if err != nil {
		return errors.ErrInternal("操作失败")
	}
	return nil
}

func (s *TopicWriteService) ToggleFavorite(ctx context.Context, userID, topicID int) *errors.AppError {
	err := s.topicRepo.DB().Transaction(func(tx *gorm.DB) error {
		topic, err := s.topicRepo.FindByIDTx(tx, topicID)
		if err != nil {
			return err
		}
		if topic.Status == 1 {
			return gorm.ErrRecordNotFound
		}

		existing, findErr := s.topicRepo.FindTopicFavorite(tx, userID, topicID)

		if findErr == gorm.ErrRecordNotFound {
			if err := s.topicRepo.CreateTopicFavorite(tx, userID, topicID); err != nil {
				return err
			}
			if err := s.topicRepo.AdjustFavoriteCount(tx, topicID, 1); err != nil {
				return err
			}
			if userID != topic.UserID {
				s.helpers.AdjustMoemoepoint(tx, topic.UserID, 1,
					moemoepoint.ReasonLiked, moemoepoint.Ref("topic", topicID))
				if err := s.helpers.CreateTopicMessageWithContent(tx, userID, topic.UserID, "favorite",
					truncate(topic.Title, constants.TextPreviewLength), topicID, 0, 0); err != nil {
					return err
				}
			}
		} else if findErr == nil {
			if err := s.topicRepo.DeleteTopicFavorite(tx, existing); err != nil {
				return err
			}
			if err := s.topicRepo.AdjustFavoriteCount(tx, topicID, -1); err != nil {
				return err
			}
			if userID != topic.UserID {
				s.helpers.AdjustMoemoepoint(tx, topic.UserID, -1,
					moemoepoint.ReasonLiked, moemoepoint.Ref("topic", topicID))
			}
		} else {
			return findErr
		}
		return nil
	})

	if err == gorm.ErrRecordNotFound {
		return errors.ErrNotFound("未找到该话题")
	}
	if err != nil {
		return errors.ErrInternal("操作失败")
	}
	return nil
}

func (s *TopicWriteService) ToggleHide(ctx context.Context, userID int, canModerate bool, topicID int) *errors.AppError {
	topic, err := s.topicRepo.FindByID(topicID)
	if err != nil {
		return errors.ErrNotFound("未找到该话题")
	}
	newStatus, hiddenBy, decisionErr := hideDecision(topic, userID, canModerate)
	if decisionErr != nil {
		return decisionErr
	}
	if err := s.topicRepo.UpdateFields(topicID, map[string]any{"status": newStatus, "hidden_by": hiddenBy}); err != nil {
		return errors.ErrInternal("操作失败")
	}
	return nil
}

func (s *TopicWriteService) SetBestAnswer(ctx context.Context, userID int, canModerate bool, topicID, replyID int) *errors.AppError {
	topic, err := s.topicRepo.FindByID(topicID)
	if err != nil {
		return errors.ErrNotFound("未找到该话题")
	}
	if topic.UserID != userID && !canModerate {
		return errors.ErrForbidden("只有话题作者或管理员可以设置最佳回答")
	}

	var reply topicModel.TopicReply
	if err := s.topicRepo.DB().First(&reply, replyID).Error; err != nil {
		return errors.ErrNotFound("未找到该回复")
	}
	if reply.TopicID != topicID {
		return errors.ErrBadRequest("该回复不属于此话题")
	}

	isCurrentBest := topic.BestAnswerID != nil && *topic.BestAnswerID == replyID
	delta := 7
	if isCurrentBest {
		delta = -7
	}

	txErr := s.topicRepo.DB().Transaction(func(tx *gorm.DB) error {
		if isCurrentBest {
			if err := tx.Model(&topicModel.Topic{}).Where("id = ?", topicID).
				Update("best_answer_id", nil).Error; err != nil {
				return err
			}
		} else {
			now := time.Now()
			if err := tx.Model(&topicModel.Topic{}).Where("id = ?", topicID).
				Updates(map[string]any{
					"best_answer_id": &replyID,
					"status_update_time": gorm.Expr(
						"CASE WHEN created > ? THEN ? ELSE status_update_time END",
						topicModel.BumpCutoff(now), now),
				}).Error; err != nil {
				return err
			}
		}
		bestReason := moemoepoint.ReasonContentApproved
		if delta < 0 {
			bestReason = moemoepoint.ReasonContentRemoved
		}
		s.helpers.AdjustMoemoepoint(tx, reply.UserID, delta, bestReason, moemoepoint.Ref("topic_reply", reply.ID))
		if !isCurrentBest {
			return s.notifier.Emit(tx, msgService.Spec{
				SenderID:   userID,
				ReceiverID: reply.UserID,
				Kind:       msgService.NotifySolution,
				Content:    replyPlainPreview(reply),
				TopicID:    topicID,
				ReplyFloor: reply.Floor,
			})
		}
		return nil
	})
	if txErr != nil {
		return errors.ErrInternal("设置最佳回答失败")
	}
	return nil
}
