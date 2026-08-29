package storelink

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"kun-galgame-api/pkg/storeclient"

	"gorm.io/gorm/clause"
)

const (
	queueSize = 1024
	// The default developer-key tier is 60 requests a minute. Pacing at 40 keeps
	// a whole minute of headroom for the campaign probe and for a burst the
	// limiter counted in the previous window.
	mintInterval = 1500 * time.Millisecond
	mintTimeout  = 10 * time.Second
)

// Start loads the cache and runs the minter until the returned func is called.
// It is nil-safe and a no-op when the store face is unconfigured — the affiliate
// template alone is a complete, working feature.
func (r *Resolver) Start() func() {
	if !r.Configured() {
		return func() {}
	}
	r.load()
	slog.Info("dlsite 短链已启用", "cached_links", r.Count())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Without this the coupon reverts to the static landing page for a whole
		// cron period after every deploy: the campaign is only ever learned from a
		// purchase-links answer, and a site whose links are all minted has none.
		r.RefreshCampaign()
		r.run(ctx)
	}()
	return func() {
		cancel()
		<-done
	}
}

// RefreshCampaign re-reads the running campaign. Every mint already restates it,
// but a site whose links are all minted mints nothing for days, so the cron
// carries the coupon on its own.
func (r *Resolver) RefreshCampaign() {
	if !r.Configured() || r.stopped() {
		return
	}
	probe := r.probeProduct()
	if probe == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), mintTimeout)
	defer cancel()
	links, err := r.client.PurchaseLinks(ctx, probe)
	if err != nil {
		slog.Warn("dlsite 优惠券活动刷新失败, 沿用上一次的结果", "error", err)
		return
	}
	r.setCampaign(links.Campaign, links.CouponURL)
}

func (r *Resolver) run(ctx context.Context) {
	timer := time.NewTimer(mintInterval)
	defer timer.Stop()
	if !timer.Stop() {
		<-timer.C
	}
	for {
		var workno string
		select {
		case <-ctx.Done():
			return
		case workno = <-r.queue:
		}

		r.mint(ctx, workno)

		timer.Reset(mintInterval)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}

func (r *Resolver) enqueue(workno string) {
	if !r.Configured() || !storeclient.ValidProductID(workno) {
		return
	}
	r.pendMu.Lock()
	_, isBad := r.bad[workno]
	_, isQueued := r.queued[workno]
	if r.halted || isBad || isQueued {
		r.pendMu.Unlock()
		return
	}
	r.queued[workno] = struct{}{}
	r.pendMu.Unlock()

	select {
	case r.queue <- workno:
	default:
		// Queue full. Dropping is safe: the next reader of this galgame enqueues
		// it again, and a full queue means the minter is already saturated.
		r.release(workno)
	}
}

func (r *Resolver) release(workno string) {
	r.pendMu.Lock()
	delete(r.queued, workno)
	r.pendMu.Unlock()
}

func (r *Resolver) stopped() bool {
	r.pendMu.Lock()
	defer r.pendMu.Unlock()
	return r.halted
}

func (r *Resolver) mint(ctx context.Context, workno string) {
	defer r.release(workno)
	if r.stopped() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, mintTimeout)
	defer cancel()
	links, err := r.client.PurchaseLinks(ctx, workno)
	if err != nil {
		r.mintFailed(workno, err)
		return
	}

	row := Link{ProductID: links.ProductID, ShortURL: links.PurchaseURL}
	// DoNothing, not DoUpdates: infra pins one alias per (client, product)
	// forever and refuses to re-mint, so a row that already exists is the alias
	// the shortener is counting. Overwriting it would move attribution to a link
	// nobody clicked.
	if err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "product_id"}}, DoNothing: true,
	}).Create(&row).Error; err != nil {
		slog.Error("写入 dlsite 短链失败", "workno", workno, "error", err)
		return
	}
	r.remember(row.ProductID, row.ShortURL)
	r.setCampaign(links.Campaign, links.CouponURL)
}

func (r *Resolver) mintFailed(workno string, err error) {
	switch {
	case errors.Is(err, storeclient.ErrQuotaExceeded), errors.Is(err, storeclient.ErrUnauthorized):
		r.halt()
		slog.Error("dlsite 短链铸造已停止, 新作品回落联盟直链(丢失归因); 提高 infra 的 per-client 配额或修 store:read scope 后重启生效",
			"workno", workno, "minted", r.Count(), "error", err)
	case errors.Is(err, storeclient.ErrInvalidProduct):
		r.markBad(workno)
		slog.Warn("store 面拒绝该 workno, 不再重试", "workno", workno, "error", err)
	default:
		slog.Warn("dlsite 短链铸造失败, 下一次浏览会重试", "workno", workno, "error", err)
	}
}

func (r *Resolver) halt() {
	r.pendMu.Lock()
	r.halted = true
	r.pendMu.Unlock()
}

func (r *Resolver) markBad(workno string) {
	r.pendMu.Lock()
	r.bad[workno] = struct{}{}
	r.pendMu.Unlock()
}
