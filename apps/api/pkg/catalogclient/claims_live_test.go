package catalogclient

import (
	"context"
	"os"
	"testing"
	"time"
)

// A live smoke against a running catalog. The v1 feed was replaced wholesale on
// 2026-08-28 and the last cutover taught that a green stub suite proves the
// fixtures agree with the code, not that the wire does.
//
//	SMOKE_CLAIM_EVENTS=1 SMOKE_CATALOG_BASE=… SMOKE_CATALOG_KEY=… go test -run TestLiveClaim
func liveClaimClient(t *testing.T) *Client {
	t.Helper()
	if os.Getenv("SMOKE_CLAIM_EVENTS") == "" {
		t.Skip("set SMOKE_CLAIM_EVENTS=1 to run against a live catalog")
	}
	base, key := os.Getenv("SMOKE_CATALOG_BASE"), os.Getenv("SMOKE_CATALOG_KEY")
	if base == "" || key == "" {
		t.Fatal("SMOKE_CATALOG_BASE and SMOKE_CATALOG_KEY are required")
	}
	return New(Config{BaseURL: base, AppKey: key})
}

const liveSite = "kungal"

func TestLiveClaimEventHead(t *testing.T) {
	c := liveClaimClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	head, err := c.ClaimEventHead(ctx, liveSite)
	if err != nil {
		t.Fatalf("ClaimEventHead: %v", err)
	}
	if head <= 0 {
		t.Fatalf("head = %d; the feed answered but named no newest event", head)
	}
	t.Logf("head = %d", head)

	// The head must be past every page the ascending walk can reach, or seeding
	// puts the watermark inside history and the cron replays it.
	page, err := c.ClaimEventsSince(ctx, head, 100, liveSite)
	if err != nil {
		t.Fatalf("walk from head: %v", err)
	}
	if len(page.Items) != 0 || page.HasMore {
		t.Errorf("walking from the head returned %d items HasMore=%v, want 0/false",
			len(page.Items), page.HasMore)
	}
}

func TestLiveClaimEventCursorIsExclusive(t *testing.T) {
	c := liveClaimClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	first, err := c.ClaimEventsSince(ctx, 0, 3, liveSite)
	if err != nil {
		t.Fatalf("ClaimEventsSince(0): %v", err)
	}
	if len(first.Items) < 3 {
		t.Fatalf("feed returned %d rows for limit=3", len(first.Items))
	}
	for i := 1; i < len(first.Items); i++ {
		if first.Items[i].ID <= first.Items[i-1].ID {
			t.Fatalf("ids are not ascending: %d then %d — sort=recorded_asc did not apply",
				first.Items[i-1].ID, first.Items[i].ID)
		}
	}

	// The forum stores its own event id, not the opaque next_cursor, and re-derives
	// the cursor each page. If cur_<id> were inclusive the last applied event would
	// come back every run and re-drive EnsureLocalStub and UnpublishLocal.
	mid := first.Items[1].ID
	after, err := c.ClaimEventsSince(ctx, mid, 3, liveSite)
	if err != nil {
		t.Fatalf("ClaimEventsSince(%d): %v", mid, err)
	}
	if len(after.Items) == 0 {
		t.Fatal("no rows after the second event")
	}
	if after.Items[0].ID != first.Items[2].ID {
		t.Errorf("cursor at %d returned %d first, want %d — the cursor is inclusive",
			mid, after.Items[0].ID, first.Items[2].ID)
	}
}

func TestLiveClaimEventFieldsSurviveTheWire(t *testing.T) {
	c := liveClaimClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	head, err := c.ClaimEventHead(ctx, liveSite)
	if err != nil {
		t.Fatalf("ClaimEventHead: %v", err)
	}
	from := head - 2000
	if from < 0 {
		from = 0
	}

	var (
		rows, withActor, withProduct, withFrom, pages int
		states                                        = map[string]int{}
		prev                                          int64
	)
	for at := from; pages < 30; pages++ {
		page, err := c.ClaimEventsSince(ctx, at, 100, liveSite)
		if err != nil {
			t.Fatalf("page at %d: %v", at, err)
		}
		for i := range page.Items {
			ev := &page.Items[i]
			rows++
			// Every id is a decimal string upstream. A silent decode failure here
			// reads as "no events", which is exactly how this feed fails quietly.
			if ev.ID <= prev {
				t.Fatalf("id %d did not advance past %d", ev.ID, prev)
			}
			prev = ev.ID
			if ev.WorkID <= 0 {
				t.Errorf("event %d has work_id %d — the string id did not decode", ev.ID, ev.WorkID)
			}
			if ev.Site != liveSite {
				t.Errorf("event %d carries site %q despite site=%s", ev.ID, ev.Site, liveSite)
			}
			if ev.CreatedAt.IsZero() {
				t.Errorf("event %d has no created_at", ev.ID)
			}
			if ev.ToState == "" {
				t.Errorf("event %d has no to_state — the whole effect switch keys off it", ev.ID)
			}
			states[ev.ToState]++
			if ev.ActorUID > 0 {
				withActor++
			}
			if ev.ProductWorkID != nil && *ev.ProductWorkID > 0 {
				withProduct++
			}
			if ev.FromState != nil {
				withFrom++
			}
			at = ev.ID
		}
		if !page.HasMore || len(page.Items) == 0 {
			break
		}
	}

	t.Logf("read %d events over %d pages; to_state=%v; actor_uid set on %d, product_work_id on %d, from_state on %d",
		rows, pages+1, states, withActor, withProduct, withFrom)
	if rows == 0 {
		t.Fatal("the feed returned nothing over the last 2000 ids")
	}
	// These three are what the sync actually consumes. A feed that decoded but
	// left them empty would run silently and award nobody.
	if withActor == 0 {
		t.Error("not one event carried actor_uid — the reward has no recipient")
	}
	if withProduct == 0 {
		t.Error("not one event carried product_work_id — every effect is skipped as unanchored")
	}
	if withFrom == 0 {
		t.Error("not one event carried from_state — no transition can be classed as an approval")
	}
	if len(states) < 2 {
		t.Errorf("to_state took only %v across %d events; the transition field looks constant", states, rows)
	}
}
