package service

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"kun-galgame-api/internal/constants"
	msgService "kun-galgame-api/internal/message/service"
	"kun-galgame-api/internal/topic/dto"
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/internal/topic/repository"
	"kun-galgame-api/internal/trust/gate"
	userRepo "kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/imageclient"
	"kun-galgame-api/pkg/secretbox"
	"kun-galgame-api/pkg/userclient"

	"gorm.io/gorm"
)

type LotteryService struct {
	lotteryRepo *repository.LotteryRepository
	topicRepo   *repository.TopicRepository
	stateRepo   *userRepo.StateRepository
	userClient  *userclient.Client
	notifier    msgService.Notifier
	box         *secretbox.Box
	cdnBase     string
	check       *gate.CheckService
	scan        *gate.ScanService
	imageMeta   func(hashes []string) map[string]imageclient.ImageMeta
}

// SetImageMetaResolver wires in the image service's machine grade. It stays nil
// when the image client has no credentials, and then nothing is machine marked
// and only what the author marked is withheld.
func (s *LotteryService) SetImageMetaResolver(
	resolve func(hashes []string) map[string]imageclient.ImageMeta,
) {
	s.imageMeta = resolve
}

func NewLotteryService(
	lotteryRepo *repository.LotteryRepository,
	topicRepo *repository.TopicRepository,
	stateRepo *userRepo.StateRepository,
	userClient *userclient.Client,
	notifier msgService.Notifier,
	box *secretbox.Box,
	cdnBase string,
	check *gate.CheckService,
	scan *gate.ScanService,
) *LotteryService {
	return &LotteryService{
		lotteryRepo: lotteryRepo, topicRepo: topicRepo, stateRepo: stateRepo,
		userClient: userClient, notifier: notifier, box: box, cdnBase: cdnBase,
		check: check, scan: scan,
	}
}

func lotteryModerationText(title, description string, prizes []dto.LotteryPrizeInput) string {
	parts := make([]string, 0, 2+len(prizes)*2)
	parts = append(parts, title, description)
	for _, p := range prizes {
		parts = append(parts, p.Name, p.Description)
	}
	return gate.ComposeText(parts...)
}

// eligibleToCreate is the anti-scam bar, not a permission. A lottery advertises
// free goods to every reader of a topic, which is exactly what a throwaway
// account is for, so an author needs either an account old enough to be
// inconvenient to farm or enough moemoepoint to have actually contributed.
func (s *LotteryService) eligibleToCreate(ctx context.Context, userID int) *errors.AppError {
	state, err := s.stateRepo.FindByID(userID)
	if err == nil && state.Moemoepoint >= constants.LotteryMinMoemoepoint {
		return nil
	}
	u, _, _ := s.userClient.User(ctx, userID)
	if days, ok := accountAgeDays(u); ok && days >= constants.LotteryMinAccountAgeDays {
		return nil
	}
	return errors.ErrForbidden(fmt.Sprintf(
		"发起抽奖需要注册满 %d 天, 或拥有至少 %d 萌萌点",
		constants.LotteryMinAccountAgeDays, constants.LotteryMinMoemoepoint))
}

// OAuth returns created_at as a string and it is not a column this database can
// filter on, so every account-age decision happens in Go after a hydrate.
func accountAgeDays(u userclient.User) (int, bool) {
	if u.CreatedAt == "" {
		return 0, false
	}
	t, err := time.Parse(time.RFC3339, u.CreatedAt)
	if err != nil {
		return 0, false
	}
	return int(time.Since(t).Hours() / 24), true
}

func (s *LotteryService) CreateLottery(
	ctx context.Context,
	userID int,
	canModerate bool,
	req *dto.CreateLotteryRequest,
) *errors.AppError {
	topic, err := s.topicRepo.FindByID(req.TopicID)
	if err != nil {
		return errors.ErrNotFound("未找到该话题")
	}
	if topic.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限在该话题创建抽奖")
	}
	if !canModerate {
		if appErr := s.eligibleToCreate(ctx, userID); appErr != nil {
			return appErr
		}
	}

	count, _ := s.lotteryRepo.CountByTopicID(req.TopicID)
	if count >= constants.MaxLotteriesPerTopic {
		return errors.ErrBadRequest("该话题的抽奖数已达上限")
	}

	deadline := parseLotteryDeadline(req.Deadline)
	if appErr := s.validateShape(req.EntryMode, req.FloorRule, req.DrawMode, req.DrawThreshold, deadline, req.Prizes); appErr != nil {
		return appErr
	}

	moderationText := lotteryModerationText(req.Title, req.Description, req.Prizes)
	authorID := int64(userID)
	decision, matched := s.check.Decision(ctx, moderationText, &authorID)
	if decision == gate.DecisionDeny {
		return errContentBlocked()
	}

	seed, seedHash := newDrawSeed()
	if req.EntryMode == topicModel.LotteryEntryFloor {
		// A floor lottery is decided by the floor numbers themselves, which every
		// reader can already see. Publishing a commitment for a draw that has no
		// randomness would claim a fairness proof that is not being made.
		seed, seedHash = "", ""
	}

	var newID int
	txErr := s.lotteryRepo.DB().Transaction(func(tx *gorm.DB) error {
		lottery := &topicModel.TopicLottery{
			TopicID:           req.TopicID,
			UserID:            userID,
			Title:             req.Title,
			Description:       req.Description,
			EntryMode:         req.EntryMode,
			FloorRule:         req.FloorRule,
			DrawMode:          req.DrawMode,
			DrawThreshold:     req.DrawThreshold,
			Deadline:          deadline,
			MinAccountAgeDays: req.MinAccountAgeDays,
			MinMoemoepoint:    req.MinMoemoepoint,
			ShowEntrants:      req.ShowEntrants,
			Status:            topicModel.LotteryStatusOpen,
			SeedHash:          seedHash,
			Seed:              seed,
		}
		if err := s.lotteryRepo.Create(tx, lottery); err != nil {
			return err
		}
		newID = lottery.ID
		if err := s.writePrizes(tx, lottery.ID, req.Prizes); err != nil {
			return err
		}
		return s.touchTopic(tx, req.TopicID)
	})
	if txErr != nil {
		if appErr, ok := txErr.(*errors.AppError); ok {
			return appErr
		}
		return errors.ErrInternal("创建抽奖失败")
	}

	if decision == gate.DecisionHold {
		slog.Info("trust check hold", "subject_kind", gate.SubjectKindTopicLottery,
			"subject_id", newID, "author_id", userID, "matched", matched)
	}
	s.scan.ScanBg(gate.SubjectKindTopicLottery, strconv.Itoa(newID), moderationText, int64(userID))
	return nil
}

func (s *LotteryService) validateShape(
	entryMode, floorRule, drawMode string,
	threshold int,
	deadline *time.Time,
	prizes []dto.LotteryPrizeInput,
) *errors.AppError {
	if len(prizes) == 0 {
		return errors.ErrBadRequest("至少需要一个奖项")
	}
	if len(prizes) > constants.MaxPrizesPerLottery {
		return errors.ErrBadRequest("奖项数量超出上限")
	}

	total := 0
	pointTotal := 0
	for _, p := range prizes {
		total += p.Slots
		if p.Slots > constants.MaxSlotsPerPrize {
			return errors.ErrBadRequest("单个奖项的名额超出上限")
		}
		if len(p.ImageHashes) > constants.MaxImagesPerPrize {
			return errors.ErrBadRequest(fmt.Sprintf(
				"奖项 %q 最多 %d 张图片", p.Name, constants.MaxImagesPerPrize))
		}
		for _, hash := range p.NSFWHashes {
			if !slices.Contains(p.ImageHashes, hash) {
				return errors.ErrBadRequest(fmt.Sprintf(
					"奖项 %q 标记的成人内容图片不在该奖项的图片里", p.Name))
			}
		}
		switch p.Delivery {
		case topicModel.LotteryDeliveryCode:
			if !s.box.Enabled() {
				return errors.ErrBadRequest("本站尚未配置兑换码托管密钥, 暂时无法使用「系统托管兑换码」交付方式")
			}
			if len(p.Codes) != p.Slots {
				return errors.ErrBadRequest(fmt.Sprintf(
					"奖项 %q 有 %d 个名额, 需要正好 %d 个兑换码, 当前 %d 个",
					p.Name, p.Slots, p.Slots, len(p.Codes)))
			}
		case topicModel.LotteryDeliveryPoint:
			if p.PointAmount <= 0 {
				return errors.ErrBadRequest(fmt.Sprintf("奖项 %q 的萌萌点数量必须大于 0", p.Name))
			}
			if isPointPool(p.PointMode) && p.PointAmount < p.Slots {
				return errors.ErrBadRequest(fmt.Sprintf(
					"奖项 %q 的奖池至少需要 %d 萌萌点, 每个名额至少分到 1 点", p.Name, p.Slots))
			}
			pointTotal += prizePointTotal(p.PointMode, p.PointAmount, p.Slots)
		}
	}
	if total > constants.MaxSlotsPerPrize {
		return errors.ErrBadRequest("总名额超出上限")
	}
	if pointTotal > constants.MaxLotteryPointTotal {
		return errors.ErrBadRequest(fmt.Sprintf(
			"本次抽奖共需发放 %d 萌萌点, 超过单次上限 %d",
			pointTotal, constants.MaxLotteryPointTotal))
	}

	switch drawMode {
	case topicModel.LotteryDrawDeadline:
		if deadline == nil {
			return errors.ErrBadRequest("到点开奖需要设置截止时间")
		}
		if deadline.Before(time.Now()) {
			return errors.ErrBadRequest("截止时间必须晚于当前时间")
		}
	case topicModel.LotteryDrawThreshold:
		if threshold < total {
			return errors.ErrBadRequest("满员开奖的人数必须不少于总名额")
		}
	}

	if entryMode == topicModel.LotteryEntryFloor {
		if drawMode == topicModel.LotteryDrawThreshold {
			return errors.ErrBadRequest("楼层抽奖不能使用满员开奖")
		}
		if _, appErr := parseFloorRule(floorRule, total); appErr != nil {
			return appErr
		}
	}
	return nil
}

// isPointPool reports whether point_amount is the whole prize's budget rather
// than one winner's share.
func isPointPool(mode string) bool {
	return mode == topicModel.LotteryPointSplit || mode == topicModel.LotteryPointRandom
}

func prizePointTotal(mode string, amount, slots int) int {
	if isPointPool(mode) {
		return amount
	}
	return amount * slots
}

func (s *LotteryService) writePrizes(tx *gorm.DB, lotteryID int, prizes []dto.LotteryPrizeInput) error {
	for i, p := range prizes {
		pointMode := p.PointMode
		if pointMode == "" {
			pointMode = topicModel.LotteryPointFixed
		}
		hashes := p.ImageHashes
		if hashes == nil {
			hashes = []string{}
		}
		nsfw := p.NSFWHashes
		if nsfw == nil {
			nsfw = []string{}
		}
		prize := &topicModel.TopicLotteryPrize{
			LotteryID:   lotteryID,
			Name:        p.Name,
			Description: p.Description,
			ImageHashes: hashes,
			NSFWHashes:  nsfw,
			Delivery:    p.Delivery,
			PointMode:   pointMode,
			PointAmount: p.PointAmount,
			Slots:       p.Slots,
			SortOrder:   i,
		}
		if err := s.lotteryRepo.CreatePrize(tx, prize); err != nil {
			return err
		}
		if p.Delivery != topicModel.LotteryDeliveryCode {
			continue
		}
		for _, raw := range p.Codes {
			sealed, err := s.box.Seal(strings.TrimSpace(raw))
			if err != nil {
				return errors.ErrInternal("兑换码加密失败")
			}
			if err := s.lotteryRepo.CreateCode(tx, &topicModel.TopicLotteryCode{
				LotteryID: lotteryID, PrizeID: prize.ID, Secret: sealed,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *LotteryService) touchTopic(tx *gorm.DB, topicID int) error {
	return tx.Model(&topicModel.Topic{}).
		Where("id = ? AND created > ?", topicID, topicModel.BumpCutoff(time.Now())).
		Updates(map[string]any{"status_update_time": time.Now()}).Error
}

func (s *LotteryService) UpdateLottery(
	ctx context.Context,
	userID int,
	canModerate bool,
	req *dto.UpdateLotteryRequest,
) *errors.AppError {
	lottery, err := s.lotteryRepo.FindByID(req.LotteryID)
	if err != nil {
		return errors.ErrNotFound("未找到该抽奖")
	}
	if lottery.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限修改此抽奖")
	}
	if lottery.Status != topicModel.LotteryStatusOpen {
		return errors.ErrBadRequest("抽奖已开奖或已取消, 无法修改")
	}

	deadline := parseLotteryDeadline(req.Deadline)
	rewritePrizes := len(req.Prizes) > 0
	if rewritePrizes {
		if lottery.EntryCount > 0 {
			return errors.ErrBadRequest("已经有人参与, 奖项不可再修改")
		}
		if appErr := s.validateShape(req.EntryMode, req.FloorRule, req.DrawMode, req.DrawThreshold, deadline, req.Prizes); appErr != nil {
			return appErr
		}
	}

	moderationText := lotteryModerationText(req.Title, req.Description, req.Prizes)
	authorID := int64(lottery.UserID)
	decision, matched := s.check.Decision(ctx, moderationText, &authorID)
	if decision == gate.DecisionDeny {
		return errContentBlocked()
	}

	txErr := s.lotteryRepo.DB().Transaction(func(tx *gorm.DB) error {
		fields := map[string]any{
			"title":                req.Title,
			"description":          req.Description,
			"draw_mode":            req.DrawMode,
			"draw_threshold":       req.DrawThreshold,
			"deadline":             deadline,
			"min_account_age_days": req.MinAccountAgeDays,
			"min_moemoepoint":      req.MinMoemoepoint,
			"show_entrants":        req.ShowEntrants,
			"updated":              time.Now(),
		}
		// entry_mode decides who is even in the pool, so it stops being editable
		// the moment the pool is non-empty — otherwise a signup lottery flips to
		// floor mode and silently discards everyone who already entered.
		if lottery.EntryCount == 0 {
			fields["entry_mode"] = req.EntryMode
			fields["floor_rule"] = req.FloorRule
		}
		if err := s.lotteryRepo.UpdateFields(tx, req.LotteryID, fields); err != nil {
			return err
		}
		if !rewritePrizes {
			return nil
		}
		if err := s.lotteryRepo.DeletePrizes(tx, req.LotteryID); err != nil {
			return err
		}
		return s.writePrizes(tx, req.LotteryID, req.Prizes)
	})
	if txErr != nil {
		if appErr, ok := txErr.(*errors.AppError); ok {
			return appErr
		}
		return errors.ErrInternal("更新抽奖失败")
	}

	if decision == gate.DecisionHold {
		slog.Info("trust check hold", "subject_kind", gate.SubjectKindTopicLottery,
			"subject_id", req.LotteryID, "author_id", lottery.UserID, "matched", matched)
	}
	s.scan.ScanBg(gate.SubjectKindTopicLottery, strconv.Itoa(req.LotteryID), moderationText, int64(lottery.UserID))
	return nil
}

func (s *LotteryService) DeleteLottery(userID int, canModerate bool, lotteryID int) *errors.AppError {
	lottery, err := s.lotteryRepo.FindByID(lotteryID)
	if err != nil {
		return errors.ErrNotFound("未找到该抽奖")
	}
	if lottery.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限删除此抽奖")
	}
	if lottery.Status == topicModel.LotteryStatusDrawn && !canModerate {
		return errors.ErrBadRequest("已开奖的抽奖不能删除, 中奖者需要它来领取奖品")
	}

	txErr := s.lotteryRepo.DB().Transaction(func(tx *gorm.DB) error {
		return s.lotteryRepo.Delete(tx, lotteryID)
	})
	if txErr != nil {
		return errors.ErrInternal("删除抽奖失败")
	}
	return nil
}
