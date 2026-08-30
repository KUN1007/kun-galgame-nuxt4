package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	msgService "kun-galgame-api/internal/message/service"
	"kun-galgame-api/internal/moemoepoint"
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"

	"gorm.io/gorm"
)

const drawSweepBatch = 50

// newDrawSeed returns the secret and the commitment published with it. Anyone
// can recompute the winners from the seed once it is revealed, so the author
// cannot pick a favourable ordering after seeing who entered. This defends
// against the topic author, not against the site operator, which is the honest
// boundary of a commit-reveal scheme run on the operator's own server.
func newDrawSeed() (seed, seedHash string) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		// crypto/rand failing is not a condition to paper over with time.Now().
		panic("抽奖随机数生成失败: " + err.Error())
	}
	seed = hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(seed))
	return seed, hex.EncodeToString(sum[:])
}

func rankKey(seed string, lotteryID, userID int) string {
	mac := hmac.New(sha256.New, []byte(seed))
	fmt.Fprintf(mac, "%d:%d", lotteryID, userID)
	return hex.EncodeToString(mac.Sum(nil))
}

// parseFloorRule turns the author's rule into the ordered floors that win.
// Two forms only: an explicit list ("8,18,28") and "every:N" (每 N 楼). Both are
// verifiable by a reader counting floors, which is the entire point of a floor
// lottery.
func parseFloorRule(rule string, slots int) ([]int, *errors.AppError) {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return nil, errors.ErrBadRequest("楼层抽奖需要填写楼层规则")
	}
	if after, ok := strings.CutPrefix(rule, "every:"); ok {
		step, err := strconv.Atoi(strings.TrimSpace(after))
		if err != nil || step <= 0 {
			return nil, errors.ErrBadRequest("楼层间隔必须是正整数, 例如 every:10")
		}
		floors := make([]int, 0, slots)
		for i := 1; i <= slots; i++ {
			floors = append(floors, step*i)
		}
		return floors, nil
	}

	seen := map[int]bool{}
	floors := make([]int, 0, slots)
	for _, part := range strings.Split(rule, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n <= 0 {
			return nil, errors.ErrBadRequest("楼层号必须是正整数, 用英文逗号分隔, 例如 8,18,28")
		}
		if seen[n] {
			continue
		}
		seen[n] = true
		floors = append(floors, n)
	}
	if len(floors) == 0 {
		return nil, errors.ErrBadRequest("楼层抽奖需要至少一个楼层号")
	}
	if len(floors) != slots {
		return nil, errors.ErrBadRequest(fmt.Sprintf(
			"楼层数 %d 与总名额 %d 不一致, 请让每个名额对应一个楼层", len(floors), slots))
	}
	sort.Ints(floors)
	return floors, nil
}

type LotteryDrawer struct {
	svc     *LotteryService
	running sync.Mutex
}

func NewLotteryDrawer(svc *LotteryService) *LotteryDrawer {
	return &LotteryDrawer{svc: svc}
}

// Run is the every-minute sweep. topic_poll's notification_sent column sat dead
// from the day it was declared because the job that was supposed to write it
// was never built, so the lottery's own deadline handling lives here where it
// can be seen to run, and the poll half is finally wired up alongside it.
func (d *LotteryDrawer) Run() {
	if !d.running.TryLock() {
		return
	}
	defer d.running.Unlock()

	ctx := context.Background()
	drawn := d.svc.sweepDueLotteries(ctx)
	closed := d.svc.sweepClosedPolls()
	if drawn > 0 || closed > 0 {
		slog.Info("话题小程序到点处理完成", "lotteries_drawn", drawn, "polls_closed", closed)
	}
}

func (s *LotteryService) sweepDueLotteries(ctx context.Context) int {
	due, err := s.lotteryRepo.ClaimDue(time.Now(), drawSweepBatch)
	if err != nil {
		slog.Warn("抽奖到点扫描失败", "error", err)
		return 0
	}
	drawn := 0
	for i := range due {
		if appErr := s.draw(ctx, &due[i]); appErr != nil {
			slog.Warn("自动开奖失败, 已退回待开奖队列",
				"lottery_id", due[i].ID, "error", appErr.Message)
			if err := s.lotteryRepo.ReleaseDrawing(due[i].ID); err != nil {
				slog.Error("退回待开奖队列失败, 该抽奖会卡在 drawing",
					"lottery_id", due[i].ID, "error", err)
			}
			continue
		}
		drawn++
	}
	return drawn
}

// DrawNow is the author's manual button. It takes the same status flip the
// sweep takes, so the two cannot both draw the same lottery.
func (s *LotteryService) DrawNow(ctx context.Context, userID int, canModerate bool, lotteryID int) *errors.AppError {
	lottery, err := s.lotteryRepo.FindByID(lotteryID)
	if err != nil {
		return errors.ErrNotFound("未找到该抽奖")
	}
	if lottery.UserID != userID && !canModerate {
		return errors.ErrForbidden("您没有权限为此抽奖开奖")
	}
	if lottery.Status != topicModel.LotteryStatusOpen {
		return errors.ErrBadRequest("该抽奖已经开过奖了")
	}

	res := s.lotteryRepo.DB().Exec(
		`UPDATE topic_lottery SET status = 'drawing' WHERE id = ? AND status = 'open'`, lotteryID)
	if res.Error != nil {
		return errors.ErrInternal("开奖失败")
	}
	if res.RowsAffected == 0 {
		return errors.ErrBadRequest("该抽奖正在开奖中")
	}

	lottery.Status = topicModel.LotteryStatusDrawing
	if appErr := s.draw(ctx, lottery); appErr != nil {
		if err := s.lotteryRepo.ReleaseDrawing(lotteryID); err != nil {
			slog.Error("退回待开奖队列失败", "lottery_id", lotteryID, "error", err)
		}
		return appErr
	}
	return nil
}

// draw fixes the winner list once and writes it down. It never recomputes:
// re-deriving winners at read time would change the answer every time someone
// new entered or a participant got banned.
func (s *LotteryService) draw(ctx context.Context, lottery *topicModel.TopicLottery) *errors.AppError {
	prizes, err := s.lotteryRepo.FindPrizes(lottery.ID)
	if err != nil {
		return errors.ErrInternal("获取奖项失败")
	}
	slots := make([]topicModel.TopicLotteryPrize, 0, len(prizes))
	for _, p := range prizes {
		for i := 0; i < p.Slots; i++ {
			slots = append(slots, p)
		}
	}

	var winners, pool []topicModel.TopicLotteryEntry
	var appErr *errors.AppError
	if lottery.EntryMode == topicModel.LotteryEntryFloor {
		winners, appErr = s.pickFloorWinners(ctx, lottery, len(slots))
	} else {
		pool, appErr = s.eligiblePool(ctx, lottery)
		if appErr == nil {
			winners = takeLowestRanked(pool, len(slots))
		}
	}
	if appErr != nil {
		return appErr
	}

	now := time.Now()
	txErr := s.lotteryRepo.DB().Transaction(func(tx *gorm.DB) error {
		for i := range winners {
			prize := slots[i]
			fields := map[string]any{
				"prize_id":    prize.ID,
				"rank_key":    winners[i].RankKey,
				"won_at":      now,
				"fulfillment": topicModel.LotteryFulfillPending,
				"updated":     now,
			}
			if prize.Delivery == topicModel.LotteryDeliveryCode {
				code, err := s.lotteryRepo.TakeCode(tx, prize.ID, winners[i].UserID)
				if err != nil {
					// A prize with fewer codes than slots is the author's mistake,
					// not a reason to abort the whole draw: the other winners keep
					// their prizes and this one falls back to manual delivery.
					slog.Warn("兑换码不足, 该名额改为人工发放",
						"lottery_id", lottery.ID, "prize_id", prize.ID, "user_id", winners[i].UserID)
				} else {
					fields["code_id"] = code.ID
				}
			}
			if winners[i].ID == 0 {
				entry := &topicModel.TopicLotteryEntry{
					LotteryID:   lottery.ID,
					UserID:      winners[i].UserID,
					ReplyFloor:  winners[i].ReplyFloor,
					PrizeID:     prize.ID,
					RankKey:     winners[i].RankKey,
					WonAt:       &now,
					Fulfillment: topicModel.LotteryFulfillPending,
				}
				if codeID, ok := fields["code_id"].(int); ok {
					entry.CodeID = codeID
				}
				if err := s.lotteryRepo.CreateEntry(tx, entry); err != nil {
					return err
				}
				winners[i].ID = entry.ID
				continue
			}
			if err := s.lotteryRepo.UpdateEntryFields(tx, winners[i].ID, fields); err != nil {
				return err
			}
		}
		if err := s.lotteryRepo.SyncEntryCount(tx, lottery.ID); err != nil {
			return err
		}
		return s.lotteryRepo.UpdateFields(tx, lottery.ID, map[string]any{
			"status": topicModel.LotteryStatusDrawn, "drawn_at": now, "updated": now,
		})
	})
	if txErr != nil {
		return errors.ErrInternal("开奖失败: " + txErr.Error())
	}

	s.afterDraw(lottery, slots, winners, pool)
	return nil
}

// eligiblePool ranks every eligible entry by HMAC(seed, lottery:user). Ban
// status is re-checked HERE rather than at entry time: an account banned between
// entering and the draw must not win, and the entry may be weeks old.
func (s *LotteryService) eligiblePool(
	ctx context.Context,
	lottery *topicModel.TopicLottery,
) ([]topicModel.TopicLotteryEntry, *errors.AppError) {
	entries, err := s.lotteryRepo.FindEntries(lottery.ID)
	if err != nil {
		return nil, errors.ErrInternal("获取参与记录失败")
	}
	uids := userclient.CollectIDs(entries, func(e topicModel.TopicLotteryEntry) int { return e.UserID })
	userMap := s.userClient.Hydrate(ctx, uids)

	pool := make([]topicModel.TopicLotteryEntry, 0, len(entries))
	for _, e := range entries {
		if !userclient.IsRenderable(userMap[e.UserID]) {
			continue
		}
		e.RankKey = rankKey(lottery.Seed, lottery.ID, e.UserID)
		pool = append(pool, e)
	}
	sort.Slice(pool, func(i, j int) bool { return pool[i].RankKey < pool[j].RankKey })
	return pool, nil
}

func takeLowestRanked(pool []topicModel.TopicLotteryEntry, slots int) []topicModel.TopicLotteryEntry {
	if len(pool) > slots {
		return pool[:slots]
	}
	return pool
}

// pickFloorWinners resolves the floor rule against the replies that exist right
// now. A floor nobody reached, or one whose reply was deleted, simply has no
// winner — inventing a substitute would break the one property a floor lottery
// has, that a reader can verify it by counting.
func (s *LotteryService) pickFloorWinners(
	ctx context.Context,
	lottery *topicModel.TopicLottery,
	slots int,
) ([]topicModel.TopicLotteryEntry, *errors.AppError) {
	floors, appErr := parseFloorRule(lottery.FloorRule, slots)
	if appErr != nil {
		return nil, appErr
	}
	replies, err := s.lotteryRepo.FindRepliesByFloors(lottery.TopicID, floors)
	if err != nil {
		return nil, errors.ErrInternal("获取楼层失败")
	}
	byFloor := make(map[int]int, len(replies))
	for _, r := range replies {
		byFloor[r.Floor] = r.UserID
	}

	uids := make([]int, 0, len(replies))
	for _, r := range replies {
		uids = append(uids, r.UserID)
	}
	userMap := s.userClient.Hydrate(ctx, uids)

	seen := map[int]bool{}
	out := make([]topicModel.TopicLotteryEntry, 0, len(floors))
	for _, floor := range floors {
		uid, ok := byFloor[floor]
		if !ok || uid == lottery.UserID || seen[uid] {
			continue
		}
		if !userclient.IsRenderable(userMap[uid]) {
			continue
		}
		seen[uid] = true
		out = append(out, topicModel.TopicLotteryEntry{
			LotteryID: lottery.ID, UserID: uid, ReplyFloor: floor,
		})
	}
	return out, nil
}

func (s *LotteryService) afterDraw(
	lottery *topicModel.TopicLottery,
	slots []topicModel.TopicLotteryPrize,
	winners, pool []topicModel.TopicLotteryEntry,
) {
	won := make(map[int]bool, len(winners))
	specs := make([]msgService.Spec, 0, len(pool))
	for i := range winners {
		prize := slots[i]
		won[winners[i].UserID] = true
		specs = append(specs, msgService.Spec{
			SenderID:   lottery.UserID,
			ReceiverID: winners[i].UserID,
			Kind:       msgService.NotifyLotteryWon,
			Content:    fmt.Sprintf("恭喜您在抽奖「%s」中获得 %s", lottery.Title, prize.Name),
			TopicID:    lottery.TopicID,
		})
		if prize.Delivery == topicModel.LotteryDeliveryPoint && prize.PointAmount > 0 {
			// content_approved is the only credit reason infra exposes to a
			// downstream, so a lottery payout shows up in the user's moemoepoint
			// log labelled as approved content. Relabelling needs a new reason on
			// the OAuth side, not a local workaround.
			moemoepoint.Award(winners[i].UserID, prize.PointAmount,
				moemoepoint.ReasonContentApproved,
				fmt.Sprintf("lottery_%d", lottery.ID),
				fmt.Sprintf("kungal:lottery_won:%d_%d", lottery.ID, winners[i].UserID))
		}
	}
	// Someone who entered and lost is still waiting on an answer, and the topic
	// page will not tell them unless they go back and look.
	for _, entry := range pool {
		if won[entry.UserID] {
			continue
		}
		specs = append(specs, msgService.Spec{
			SenderID:   lottery.UserID,
			ReceiverID: entry.UserID,
			Kind:       msgService.NotifyLotteryClosed,
			Content:    fmt.Sprintf("您参与的抽奖「%s」已经开奖, 很遗憾这次没有中奖", lottery.Title),
			TopicID:    lottery.TopicID,
		})
	}
	if err := s.notifier.EmitMany(nil, specs); err != nil {
		slog.Warn("开奖通知发送失败", "lottery_id", lottery.ID, "error", err)
	}
}

// sweepClosedPolls finally writes topic_poll.notification_sent. The column has
// existed since the poll mini-app shipped with no reader and no writer anywhere
// in the tree: "tell the voters the poll closed" was declared in the schema and
// never built, so a deadline passed in silence.
func (s *LotteryService) sweepClosedPolls() int {
	polls, err := s.lotteryRepo.ClaimPollsPastDeadline(time.Now(), drawSweepBatch)
	if err != nil {
		slog.Warn("投票截止扫描失败", "error", err)
		return 0
	}
	for _, poll := range polls {
		voterIDs, err := s.lotteryRepo.FindPollVoterIDs(poll.ID)
		if err != nil {
			slog.Warn("获取投票参与者失败", "poll_id", poll.ID, "error", err)
			continue
		}
		specs := make([]msgService.Spec, 0, len(voterIDs))
		for _, uid := range voterIDs {
			specs = append(specs, msgService.Spec{
				SenderID:   poll.UserID,
				ReceiverID: uid,
				Kind:       msgService.NotifyPollClosed,
				Content:    fmt.Sprintf("您参与的投票「%s」已经截止, 可以查看结果了", poll.Title),
				TopicID:    poll.TopicID,
			})
		}
		if err := s.notifier.EmitMany(nil, specs); err != nil {
			slog.Warn("投票截止通知发送失败", "poll_id", poll.ID, "error", err)
		}
	}
	return len(polls)
}
