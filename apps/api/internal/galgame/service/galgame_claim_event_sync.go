package service

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"kun-galgame-api/internal/constants"
	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/repository"
	"kun-galgame-api/internal/moemoepoint"
	"kun-galgame-api/pkg/catalogclient"
	"kun-galgame-api/pkg/userclient"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Ingests the registry's claim-transition feed. It is NOT a translation of the
// retired wiki message feed, and three differences bite:
//
//   - The wiki delivered typed messages; the registry delivers TRANSITIONS.
//     One destination state is reachable by more than one route (`live` comes
//     both from a reviewer approving and from the owning site publishing), so
//     the local side effect keys off the DESTINATION state, never a type word.
//   - The wiki named the beneficiary of the award; the registry names the
//     ACTOR, who on an approval is the reviewer. See submitterOf.
//   - The id spaces are disjoint and both are small integers, so the cursor key
//     AND the moemoepoint idempotency key live in fresh namespaces. Reusing
//     either silently mistakes a wiki message for a claim event.
//
// At-least-once with idempotent effects: the cursor advances past an event only
// once its effects land, a transient failure holds it, and a permanent
// rejection is logged and skipped so one poison row cannot wedge the feed.
type GalgameClaimEventSync struct {
	catalog     *catalogclient.Client
	galgameRepo *repository.GalgameRepository
	rdb         *redis.Client
	batch       int
}

func NewGalgameClaimEventSync(
	catalog *catalogclient.Client,
	galgameRepo *repository.GalgameRepository,
	rdb *redis.Client,
) *GalgameClaimEventSync {
	return &GalgameClaimEventSync{
		catalog: catalog, galgameRepo: galgameRepo, rdb: rdb, batch: claimFeedBatch,
	}
}

const (
	claimCursorKey = "catalog:claim:cron:since"

	claimSubmitterKeyPrefix = "catalog:claim:submitter:"

	claimSubmitterTTL = 90 * 24 * time.Hour

	claimFeedBatch    = 1000
	claimMaxPagesRun  = 50
	claimFeedPageWait = 5 * time.Minute
)

func (s *GalgameClaimEventSync) Run() {
	if s.catalog == nil || !s.catalog.Configured() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), claimFeedPageWait)
	defer cancel()

	since, seeded, err := s.readCursor(ctx)
	if err != nil {
		slog.Warn("读取 claim 事件游标失败 (本轮跳过)", "error", err)
		return
	}
	if !seeded {
		head, err := s.feedHead(ctx)
		if err != nil {
			slog.Warn("catalog claim feed 定位队首失败, 下一轮重试", "error", err)
			return
		}
		s.writeCursor(ctx, head)
		slog.Info("catalog claim 同步已初始化游标 (不回填历史)", "head", head)
		return
	}

	startedFrom, maxSeen := since, since
	for pages := 1; pages <= claimMaxPagesRun; pages++ {
		page, err := s.catalog.ClaimEventsSince(ctx, maxSeen, s.batch, claimSite)
		if err != nil {
			slog.Warn("catalog claim feed 拉取失败", "error", err, "since", maxSeen)
			break
		}
		if len(page.Items) == 0 {
			break
		}
		holding := false
		for i := range page.Items {
			if s.apply(ctx, &page.Items[i]) {
				holding = true
				break
			}
			if page.Items[i].ID > maxSeen {
				maxSeen = page.Items[i].ID
			}
		}
		if holding || len(page.Items) < s.batch {
			break
		}
	}

	if maxSeen > startedFrom {
		s.writeCursor(ctx, maxSeen)
		slog.Info("catalog claim 同步完成", "from", startedFrom, "to", maxSeen)
	}
}

func (s *GalgameClaimEventSync) feedHead(ctx context.Context) (int64, error) {
	var head int64
	for pages := 1; pages <= claimMaxPagesRun; pages++ {
		page, err := s.catalog.ClaimEventsSince(ctx, head, s.batch, claimSite)
		if err != nil {
			return 0, err
		}
		if len(page.Items) == 0 {
			break
		}
		head = page.Items[len(page.Items)-1].ID
		if len(page.Items) < s.batch {
			break
		}
	}
	return head, nil
}

const claimSite = client.ClaimSiteKungal

type claimEffect int

const (
	claimEffectNone claimEffect = iota
	claimEffectSeedStub
	claimEffectUnpublish
	claimEffectRememberSubmitter
	claimEffectUnknownState
)

func effectOf(ev *catalogclient.ClaimEventFeedItem) claimEffect {
	if ev.ProductWorkID == nil || *ev.ProductWorkID <= 0 {
		return claimEffectNone
	}
	switch ev.ToState {
	case catalogclient.ClaimStateLive:
		return claimEffectSeedStub
	// A ban unpublishes; it must never delete the local row. galgame_comment,
	// galgame_like, galgame_favorite and galgame_resource are ON DELETE CASCADE
	// off galgame (galgame_rating is RESTRICT), so dropping the stub took every
	// comment and resource down with it — or silently failed on a rated entry —
	// and a later unban rebuilt the row owned by whoever lifted the ban.
	case catalogclient.ClaimStateHidden:
		return claimEffectUnpublish
	case catalogclient.ClaimStatePending:
		return claimEffectRememberSubmitter
	case catalogclient.ClaimStateDraft, catalogclient.ClaimStateDeclined:
		return claimEffectUnpublish
	default:
		return claimEffectUnknownState
	}
}

func isApproval(ev *catalogclient.ClaimEventFeedItem) bool {
	return ev.ToState == catalogclient.ClaimStateLive &&
		ev.FromState != nil && *ev.FromState == catalogclient.ClaimStatePending
}

// isDirectBirth is the trusted-submit path: the registry mints the claim
// straight to live, so the birth event has no from_state. It is not an
// approval (a reviewer never touched it), but it still earns the create award.
func isDirectBirth(ev *catalogclient.ClaimEventFeedItem) bool {
	return ev.ToState == catalogclient.ClaimStateLive && ev.FromState == nil
}

func (s *GalgameClaimEventSync) claimantOf(ctx context.Context, ev *catalogclient.ClaimEventFeedItem, gid int) int {
	if isApproval(ev) {
		return s.submitterOf(ctx, ev.WorkID, gid)
	}
	// Lifting a ban restores what somebody else established, so the admin who
	// lifted it is not the author. Everything else that reaches live — a claimed
	// draft being published — really is the actor's own entry.
	if ev.FromState != nil && *ev.FromState == catalogclient.ClaimStateHidden {
		return 0
	}
	return int(ev.ActorUID)
}

func (s *GalgameClaimEventSync) apply(ctx context.Context, ev *catalogclient.ClaimEventFeedItem) (retry bool) {
	switch effectOf(ev) {
	case claimEffectSeedStub:
		gid := int(*ev.ProductWorkID)
		creator := s.claimantOf(ctx, ev, gid)
		if err := s.galgameRepo.DB().Transaction(func(tx *gorm.DB) error {
			if err := s.galgameRepo.PublishLocal(tx, gid); err != nil {
				return err
			}
			if creator > 0 {
				return s.galgameRepo.SetCreatorIfUnset(tx, gid, creator)
			}
			return nil
		}); err != nil {
			slog.Warn("claim live: 发布本地行失败, 将重试", "event", ev.ID, "gid", gid, "error", err)
			return true
		}
		if isApproval(ev) || isDirectBirth(ev) {
			return s.awardLive(ctx, ev, gid)
		}
	case claimEffectUnpublish:
		if err := s.galgameRepo.UnpublishLocal(int(*ev.ProductWorkID)); err != nil {
			slog.Warn("claim unpublish: 下架本地行失败, 将重试",
				"event", ev.ID, "gid", *ev.ProductWorkID, "error", err)
			return true
		}
	case claimEffectRememberSubmitter:
		s.rememberSubmitter(ctx, ev)
	case claimEffectUnknownState:
		slog.Warn("收到未识别的 claim 目标状态, 跳过", "state", ev.ToState, "event", ev.ID)
	}
	return false
}

func (s *GalgameClaimEventSync) awardLive(ctx context.Context, ev *catalogclient.ClaimEventFeedItem, gid int) (retry bool) {
	uid := s.claimantOf(ctx, ev, gid)
	if uid <= 0 {
		slog.Error("claim live: 无法归属投稿人, 跳过发奖",
			"event", ev.ID, "work", ev.WorkID, "gid", gid)
		return false
	}
	if err := moemoepoint.AwardSync(uid, constants.RewardCreateGalgame,
		moemoepoint.ReasonContentApproved, moemoepoint.Ref("galgame", gid),
		moemoepoint.Key("claim_approved", strconv.FormatInt(ev.ID, 10))); err != nil {
		if moemoepoint.IsPermanentAwardError(err) {
			slog.Error("claim live: 发奖被 OAuth 永久拒绝, 跳过该事件",
				"event", ev.ID, "target", uid, "error", err)
			return false
		}
		slog.Warn("claim live: 发奖瞬时失败, 将重试", "event", ev.ID, "target", uid, "error", err)
		return true
	}
	return false
}

func (s *GalgameClaimEventSync) rememberSubmitter(ctx context.Context, ev *catalogclient.ClaimEventFeedItem) {
	if ev.ActorUID <= 0 {
		return
	}
	key := claimSubmitterKeyPrefix + strconv.FormatInt(ev.WorkID, 10)
	if err := s.rdb.Set(ctx, key, ev.ActorUID, claimSubmitterTTL).Err(); err != nil {
		slog.Warn("记录投稿人失败", "work", ev.WorkID, "error", err)
	}
}

// Redis only remembers the submitter for 90 days and loses it on a flush, so
// Submit also stamps the creator on the local row the moment a submission lands.
// That row is the durable answer; the cache just saves a query.
func (s *GalgameClaimEventSync) submitterOf(ctx context.Context, workID int64, gid int) int {
	if v, err := s.rdb.Get(ctx, claimSubmitterKeyPrefix+strconv.FormatInt(workID, 10)).Int(); err == nil {
		return v
	}
	if gid <= 0 {
		return 0
	}
	return userclient.DerefID(s.galgameRepo.FindLocal(gid).CreatorUserID)
}

func (s *GalgameClaimEventSync) readCursor(ctx context.Context) (since int64, seeded bool, err error) {
	v, err := s.rdb.Get(ctx, claimCursorKey).Int64()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return v, true, nil
}

func (s *GalgameClaimEventSync) writeCursor(ctx context.Context, id int64) {
	if err := s.rdb.Set(ctx, claimCursorKey, id, 0).Err(); err != nil {
		slog.Warn("写入 claim 事件游标失败", "id", id, "error", err)
	}
}
