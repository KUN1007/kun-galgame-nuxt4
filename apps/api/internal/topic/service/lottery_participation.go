package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"kun-galgame-api/internal/topic/dto"
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/pkg/errors"

	"gorm.io/gorm"
)

func (s *LotteryService) Enter(ctx context.Context, userID int, lotteryID int) *errors.AppError {
	lottery, err := s.lotteryRepo.FindByID(lotteryID)
	if err != nil {
		return errors.ErrNotFound("未找到该抽奖")
	}
	if appErr := s.entryBlocker(ctx, lottery, userID); appErr != nil {
		return appErr
	}

	txErr := s.lotteryRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.lotteryRepo.CreateEntry(tx, &topicModel.TopicLotteryEntry{
			LotteryID: lotteryID, UserID: userID,
		}); err != nil {
			return err
		}
		return s.lotteryRepo.SyncEntryCount(tx, lotteryID)
	})
	if txErr != nil {
		return errors.ErrBadRequest("参与失败, 您可能已经参与过了")
	}
	return nil
}

func (s *LotteryService) Cancel(userID int, canModerate bool, lotteryID int) *errors.AppError {
	lottery, err := s.lotteryRepo.FindByID(lotteryID)
	if err != nil {
		return errors.ErrNotFound("未找到该抽奖")
	}
	if lottery.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限取消此抽奖")
	}
	if lottery.Status != topicModel.LotteryStatusOpen {
		return errors.ErrBadRequest("该抽奖已经结束")
	}
	if err := s.lotteryRepo.UpdateFields(s.lotteryRepo.DB(), lotteryID, map[string]any{
		"status": topicModel.LotteryStatusCancelled, "updated": time.Now(),
	}); err != nil {
		return errors.ErrInternal("取消抽奖失败")
	}
	return nil
}

func (s *LotteryService) Withdraw(userID, lotteryID int) *errors.AppError {
	lottery, err := s.lotteryRepo.FindByID(lotteryID)
	if err != nil {
		return errors.ErrNotFound("未找到该抽奖")
	}
	if lottery.Status != topicModel.LotteryStatusOpen {
		return errors.ErrBadRequest("抽奖已经结束, 无法退出")
	}
	txErr := s.lotteryRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.lotteryRepo.DeleteEntry(tx, lotteryID, userID); err != nil {
			return err
		}
		return s.lotteryRepo.SyncEntryCount(tx, lotteryID)
	})
	if txErr != nil {
		return errors.ErrInternal("退出抽奖失败")
	}
	return nil
}

// entryBlocker returns the reason this user may not enter, or nil. It is the
// single source for both the button state and the POST guard, so the two cannot
// drift into a button that is enabled for a request the server rejects.
func (s *LotteryService) entryBlocker(ctx context.Context, lottery *topicModel.TopicLottery, userID int) *errors.AppError {
	if userID == 0 {
		return errors.ErrForbidden("请先登录")
	}
	if lottery.Status != topicModel.LotteryStatusOpen {
		return errors.ErrBadRequest("抽奖已经结束")
	}
	if lottery.EntryMode == topicModel.LotteryEntryFloor {
		return errors.ErrBadRequest("楼层抽奖无需报名, 直接回帖即可")
	}
	if lottery.Deadline != nil && time.Now().After(*lottery.Deadline) {
		return errors.ErrBadRequest("抽奖已过截止时间")
	}
	if lottery.UserID == userID {
		return errors.ErrBadRequest("不能参加自己发起的抽奖")
	}
	if lottery.EntryMode == topicModel.LotteryEntryReply {
		replied, _ := s.lotteryRepo.HasRepliedTo(lottery.TopicID, userID)
		if !replied {
			return errors.ErrBadRequest("本抽奖要求先在该话题下回复")
		}
	}
	if lottery.MinMoemoepoint > 0 {
		state, err := s.stateRepo.FindByID(userID)
		if err != nil || state.Moemoepoint < lottery.MinMoemoepoint {
			return errors.ErrForbidden(fmt.Sprintf("本抽奖要求至少 %d 萌萌点", lottery.MinMoemoepoint))
		}
	}
	if lottery.MinAccountAgeDays > 0 {
		u, _, _ := s.userClient.User(ctx, userID)
		days, ok := accountAgeDays(u)
		if !ok || days < lottery.MinAccountAgeDays {
			return errors.ErrForbidden(fmt.Sprintf("本抽奖要求注册满 %d 天", lottery.MinAccountAgeDays))
		}
	}
	return nil
}

// ClaimCode is the ONLY path that returns a plaintext activation code. It is a
// POST for that reason: nothing about it may be cacheable, prefetchable, or
// reachable from a page load whose payload Nuxt inlines into the SSR HTML.
func (s *LotteryService) ClaimCode(userID, lotteryID int) (string, *errors.AppError) {
	lottery, err := s.lotteryRepo.FindByID(lotteryID)
	if err != nil {
		return "", errors.ErrNotFound("未找到该抽奖")
	}
	entry, err := s.lotteryRepo.FindEntry(lotteryID, userID)
	if err != nil || entry.PrizeID == 0 {
		return "", errors.ErrForbidden("您不是该抽奖的中奖者")
	}
	if entry.CodeID == 0 {
		return "", errors.ErrBadRequest("该奖项不是兑换码, 请等待发起人联系您")
	}
	if lottery.Status != topicModel.LotteryStatusDrawn {
		return "", errors.ErrBadRequest("抽奖尚未开奖")
	}
	code, err := s.lotteryRepo.FindCodeByID(entry.CodeID)
	if err != nil {
		return "", errors.ErrInternal("兑换码丢失, 请联系管理员")
	}
	if code.ClaimedBy != userID {
		return "", errors.ErrForbidden("该兑换码不属于您")
	}
	plain, err := s.box.Open(code.Secret)
	if err != nil {
		slog.Error("兑换码解密失败", "lottery_id", lotteryID, "code_id", code.ID, "error", err)
		return "", errors.ErrInternal("兑换码解密失败, 请联系管理员")
	}
	if entry.Fulfillment != topicModel.LotteryFulfillReceived {
		_ = s.lotteryRepo.UpdateEntryFields(s.lotteryRepo.DB(), entry.ID, map[string]any{
			"fulfillment": topicModel.LotteryFulfillReceived, "updated": time.Now(),
		})
	}
	return plain, nil
}

func (s *LotteryService) SetFulfillment(
	userID int,
	canModerate bool,
	req *dto.LotteryFulfillRequest,
) *errors.AppError {
	lottery, err := s.lotteryRepo.FindByID(req.LotteryID)
	if err != nil {
		return errors.ErrNotFound("未找到该抽奖")
	}
	entries, err := s.lotteryRepo.FindWinners(req.LotteryID)
	if err != nil {
		return errors.ErrInternal("获取中奖记录失败")
	}
	var target *topicModel.TopicLotteryEntry
	for i := range entries {
		if entries[i].ID == req.EntryID {
			target = &entries[i]
			break
		}
	}
	if target == nil {
		return errors.ErrNotFound("未找到该中奖记录")
	}

	isAuthor := lottery.UserID == userID
	isWinner := target.UserID == userID
	// The winner may only confirm receipt or give the prize up; every other
	// transition is the author's to make.
	if !isAuthor && !canModerate {
		if !isWinner {
			return errors.ErrForbidden("您没有权限修改此记录")
		}
		if req.Fulfillment != topicModel.LotteryFulfillReceived &&
			req.Fulfillment != topicModel.LotteryFulfillForfeited {
			return errors.ErrForbidden("中奖者只能确认收货或放弃奖品")
		}
	}

	if err := s.lotteryRepo.UpdateEntryFields(s.lotteryRepo.DB(), target.ID, map[string]any{
		"fulfillment": req.Fulfillment, "updated": time.Now(),
	}); err != nil {
		return errors.ErrInternal("更新履约状态失败")
	}
	return nil
}
