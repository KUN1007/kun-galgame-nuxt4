package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"kun-galgame-api/internal/constants"
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/topic/dto"
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/internal/topic/repository"
	"kun-galgame-api/internal/trust/gate"
	userRepo "kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/perm"
	"kun-galgame-api/pkg/userclient"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PollService struct {
	pollRepo   *repository.PollRepository
	topicRepo  *repository.TopicRepository
	stateRepo  *userRepo.StateRepository
	userClient *userclient.Client
	rdb        *redis.Client
	check      *gate.CheckService
	scan       *gate.ScanService
}

func NewPollService(
	pollRepo *repository.PollRepository,
	topicRepo *repository.TopicRepository,
	stateRepo *userRepo.StateRepository,
	userClient *userclient.Client,
	rdb *redis.Client,
	check *gate.CheckService,
	scan *gate.ScanService,
) *PollService {
	return &PollService{
		pollRepo: pollRepo, topicRepo: topicRepo,
		stateRepo: stateRepo, userClient: userClient, rdb: rdb,
		check: check, scan: scan,
	}
}

func pollModerationText(title, description string, optionTexts []string) string {
	parts := make([]string, 0, 2+len(optionTexts))
	parts = append(parts, title, description)
	parts = append(parts, optionTexts...)
	return gate.ComposeText(parts...)
}

func (s *PollService) CreatePoll(
	ctx context.Context,
	userID int,
	canModerate bool,
	req *dto.CreatePollRequest,
) *errors.AppError {
	topic, err := s.topicRepo.FindByID(req.TopicID)
	if err != nil {
		return errors.ErrNotFound("未找到该话题")
	}
	if topic.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限创建投票")
	}

	count, _ := s.pollRepo.CountByTopicID(req.TopicID)
	if count >= constants.MaxPollsPerTopic {
		return errors.ErrBadRequest("该话题的投票数已达上限")
	}

	var deadline *time.Time
	if req.Deadline != nil {
		if t, err := time.Parse(time.RFC3339, *req.Deadline); err == nil {
			deadline = &t
		}
	}

	optionTexts := make([]string, len(req.Options))
	for i, opt := range req.Options {
		optionTexts[i] = opt.Text
	}
	moderationText := pollModerationText(req.Title, req.Description, optionTexts)
	authorID := int64(userID)
	decision, matched := s.check.Decision(ctx, moderationText, &authorID)
	if decision == gate.DecisionDeny {
		return errContentBlocked()
	}

	var newPollID int
	txErr := s.pollRepo.DB().Transaction(func(tx *gorm.DB) error {
		poll := &topicModel.TopicPoll{
			Title:            req.Title,
			Description:      req.Description,
			Type:             req.Type,
			MinChoice:        req.MinChoice,
			MaxChoice:        req.MaxChoice,
			Deadline:         deadline,
			ResultVisibility: req.ResultVisibility,
			IsAnonymous:      req.IsAnonymous,
			CanChangeVote:    req.CanChangeVote,
			TopicID:          req.TopicID,
			UserID:           userID,
		}
		if err := s.pollRepo.CreatePoll(tx, poll); err != nil {
			return err
		}
		newPollID = poll.ID

		for _, opt := range req.Options {
			if err := s.pollRepo.CreatePollOption(tx, &topicModel.TopicPollOption{
				Text:   opt.Text,
				PollID: poll.ID,
			}); err != nil {
				return err
			}
		}

		return s.pollRepo.TouchTopicStatusUpdateTime(tx, req.TopicID, time.Now())
	})

	if txErr != nil {
		return errors.ErrInternal("创建投票失败")
	}

	if decision == gate.DecisionHold {
		slog.Info("trust check hold", "subject_kind", gate.SubjectKindTopicPoll, "subject_id", newPollID, "author_id", userID, "matched", matched)
	}
	s.scan.ScanBg(gate.SubjectKindTopicPoll, strconv.Itoa(newPollID), moderationText, int64(userID))
	return nil
}

func (s *PollService) GetPollsByTopic(
	ctx context.Context,
	topicID int,
	userInfo *middleware.UserInfo,
) ([]dto.TopicPollResponse, *errors.AppError) {
	polls, err := s.pollRepo.FindByTopicID(topicID)
	if err != nil {
		return nil, errors.ErrInternal("获取投票失败")
	}

	userID := 0
	canModerate := false
	if userInfo != nil {
		userID = userInfo.ID
		canModerate = perm.CanUser(userInfo.ID, userInfo.Roles, perm.PollViewRestricted)
	}

	responses := make([]dto.TopicPollResponse, 0, len(polls))
	for _, poll := range polls {
		resp, ok := s.buildPollResponse(ctx, &poll, userID, canModerate)
		if !ok {
			continue
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func (s *PollService) Vote(
	ctx context.Context,
	user *middleware.UserInfo,
	req *dto.VoteRequest,
) *errors.AppError {
	userID := user.ID
	poll, err := s.pollRepo.FindByID(req.PollID)
	if err != nil {
		return errors.ErrNotFound("未找到该投票")
	}
	topicRepo := s.topicRepo
	topic, topicErr := topicRepo.FindByID(poll.TopicID)
	if topicErr != nil {
		return errors.ErrNotFound("未找到该话题")
	}

	if _, appErr := requireTopicRead(topicRepo, topic, user); appErr != nil {
		return appErr
	}

	if poll.Status == "closed" {
		return errors.ErrBadRequest("投票已被关闭")
	}
	if poll.Deadline != nil && time.Now().After(*poll.Deadline) {
		return errors.ErrBadRequest("投票已过截止日期")
	}

	if poll.Type == "single" && len(req.OptionIDArray) != 1 {
		return errors.ErrBadRequest("单选投票只能选择一个选项")
	}
	if poll.Type == "multiple" {
		if len(req.OptionIDArray) < poll.MinChoice {
			return errors.ErrBadRequest("至少选择 " + fmt.Sprintf("%d", poll.MinChoice) + " 个选项")
		}
		if len(req.OptionIDArray) > poll.MaxChoice {
			return errors.ErrBadRequest("最多选择 " + fmt.Sprintf("%d", poll.MaxChoice) + " 个选项")
		}
	}

	hasVoted, _ := s.pollRepo.HasUserVoted(req.PollID, userID)
	if hasVoted && !poll.CanChangeVote {
		return errors.ErrBadRequest("该投票不允许修改投票结果")
	}

	txErr := s.pollRepo.DB().Transaction(func(tx *gorm.DB) error {
		if hasVoted {
			oldOptionIDs, _ := s.pollRepo.FindUserVoteOptionIDs(req.PollID, userID)
			if err := s.pollRepo.DeleteUserVotes(tx, req.PollID, userID); err != nil {
				return err
			}
			for _, oid := range oldOptionIDs {
				if err := s.pollRepo.AdjustOptionVoteCount(tx, oid, -1); err != nil {
					return err
				}
			}
		}

		for _, optionID := range req.OptionIDArray {
			if err := s.pollRepo.CreateVote(tx, req.PollID, optionID, userID); err != nil {
				return err
			}
			if err := s.pollRepo.AdjustOptionVoteCount(tx, optionID, 1); err != nil {
				return err
			}
		}
		return nil
	})

	if txErr != nil {
		return errors.ErrInternal("投票失败")
	}
	return nil
}

func (s *PollService) UpdatePoll(
	ctx context.Context,
	userID int,
	canModerate bool,
	req *dto.UpdatePollRequest,
) *errors.AppError {
	poll, err := s.pollRepo.FindByID(req.PollID)
	if err != nil {
		return errors.ErrNotFound("投票不存在")
	}

	topic, err := s.topicRepo.FindByID(poll.TopicID)
	if err != nil {
		return errors.ErrNotFound("未找到该话题")
	}
	if topic.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限修改此投票")
	}

	var deadline *time.Time
	if req.Deadline != nil {
		if t, err := time.Parse(time.RFC3339, *req.Deadline); err == nil {
			deadline = &t
		}
	}
	scalarFields := map[string]any{
		"title":             req.Title,
		"description":       req.Description,
		"type":              req.Type,
		"min_choice":        req.MinChoice,
		"max_choice":        req.MaxChoice,
		"deadline":          deadline,
		"result_visibility": req.ResultVisibility,
		"is_anonymous":      req.IsAnonymous,
		"can_change_vote":   req.CanChangeVote,
	}

	optionTexts := make([]string, 0, len(req.Options.Add)+len(req.Options.Update))
	for _, opt := range req.Options.Add {
		optionTexts = append(optionTexts, opt.Text)
	}
	for _, opt := range req.Options.Update {
		optionTexts = append(optionTexts, opt.Text)
	}
	moderationText := pollModerationText(req.Title, req.Description, optionTexts)
	authorID := int64(poll.UserID)
	decision, matched := s.check.Decision(ctx, moderationText, &authorID)
	if decision == gate.DecisionDeny {
		return errContentBlocked()
	}

	totalOptionOps := len(req.Options.Add) + len(req.Options.Update) + len(req.Options.Delete)

	if totalOptionOps == 0 {
		if err := s.pollRepo.UpdatePollFields(s.pollRepo.DB(), req.PollID, scalarFields); err != nil {
			return errors.ErrInternal("更新投票失败")
		}
		s.afterPollModeration(decision, matched, req.PollID, poll.UserID, moderationText)
		return nil
	}

	if !poll.CanChangeVote {
		return errors.ErrBadRequest("本投票结果不可修改")
	}

	touchedIDs := collectOptionIDs(req.Options)
	current, err := s.pollRepo.FindOptionsByIDs(touchedIDs)
	if err != nil {
		return errors.ErrInternal("获取选项失败")
	}
	currentByID := make(map[int]topicModel.TopicPollOption, len(current))
	for _, o := range current {
		currentByID[o.ID] = o
	}

	for _, u := range req.Options.Update {
		opt, ok := currentByID[u.OptionID]
		if !ok {
			return errors.ErrBadRequest(fmt.Sprintf("要更新的选项 ID %d 不存在", u.OptionID))
		}
		if opt.VoteCount > 0 && opt.Text != u.Text {
			return errors.ErrBadRequest(fmt.Sprintf("选项 %q 已有投票，无法修改其内容", opt.Text))
		}
	}
	for _, id := range req.Options.Delete {
		opt, ok := currentByID[id]
		if !ok {
			return errors.ErrBadRequest(fmt.Sprintf("要删除的选项 ID %d 不存在", id))
		}
		if opt.VoteCount > 0 {
			return errors.ErrBadRequest(fmt.Sprintf("选项 %q 已有投票，无法删除", opt.Text))
		}
	}

	txErr := s.pollRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.pollRepo.UpdatePollFields(tx, req.PollID, scalarFields); err != nil {
			return err
		}
		for _, opt := range req.Options.Add {
			if err := s.pollRepo.CreatePollOption(tx, &topicModel.TopicPollOption{
				Text: opt.Text, PollID: req.PollID,
			}); err != nil {
				return err
			}
		}
		for _, u := range req.Options.Update {
			if err := s.pollRepo.UpdateOptionText(tx, u.OptionID, u.Text); err != nil {
				return err
			}
		}
		return s.pollRepo.DeleteOptionsByIDs(tx, req.Options.Delete)
	})
	if txErr != nil {
		return errors.ErrInternal("更新投票失败")
	}
	s.afterPollModeration(decision, matched, req.PollID, poll.UserID, moderationText)
	return nil
}

func (s *PollService) afterPollModeration(decision string, matched []string, pollID, authorID int, text string) {
	if decision == gate.DecisionHold {
		slog.Info("trust check hold", "subject_kind", gate.SubjectKindTopicPoll, "subject_id", pollID, "author_id", authorID, "matched", matched)
	}
	s.scan.ScanBg(gate.SubjectKindTopicPoll, strconv.Itoa(pollID), text, int64(authorID))
}

func collectOptionIDs(ops dto.PollOptionsUpdate) []int {
	ids := make([]int, 0, len(ops.Update)+len(ops.Delete))
	for _, u := range ops.Update {
		ids = append(ids, u.OptionID)
	}
	ids = append(ids, ops.Delete...)
	return ids
}

func (s *PollService) DeletePoll(
	ctx context.Context,
	userID int,
	canModerate bool,
	pollID int,
) *errors.AppError {
	poll, err := s.pollRepo.FindByID(pollID)
	if err != nil {
		return errors.ErrNotFound("未找到该投票")
	}
	if poll.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限删除此投票")
	}

	txErr := s.pollRepo.DB().Transaction(func(tx *gorm.DB) error {
		return s.pollRepo.DeletePollCascade(tx, pollID)
	})

	if txErr != nil {
		return errors.ErrInternal("删除投票失败")
	}
	return nil
}

func (s *PollService) GetVoteLog(
	ctx context.Context,
	pollID, page, limit int,
	userInfo *middleware.UserInfo,
) ([]dto.PollVoteLogEntry, int64, *errors.AppError) {
	poll, err := s.pollRepo.FindByID(pollID)
	if err != nil {
		return nil, 0, errors.ErrNotFound("未找到该投票")
	}

	userID := 0
	canModerate := false
	if userInfo != nil {
		userID = userInfo.ID
		canModerate = perm.CanUser(userInfo.ID, userInfo.Roles, perm.PollViewRestricted)
	}

	hasVoted, _ := s.pollRepo.HasUserVoted(pollID, userID)
	if !canViewResults(poll, userID, canModerate, hasVoted) || poll.IsAnonymous {
		return []dto.PollVoteLogEntry{}, 0, nil
	}

	rows, total, err := s.pollRepo.FindVoteLogs(pollID, page, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal("获取投票记录失败")
	}

	uids := userclient.CollectIDs(rows, func(r repository.VoteLogRow) int { return r.UserID })
	userMap := s.userClient.Hydrate(ctx, uids)

	entries := make([]dto.PollVoteLogEntry, 0, len(rows))
	for _, r := range rows {
		u := userMap[r.UserID]
		if !userclient.IsRenderable(u) {
			continue
		}
		entries = append(entries, dto.PollVoteLogEntry{
			ID:      r.ID,
			Created: r.CreatedAt,
			User:    dto.KunUser{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			Option:  r.OptionText,
		})
	}
	return entries, total, nil
}
