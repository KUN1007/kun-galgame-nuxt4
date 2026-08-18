package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"kun-galgame-api/pkg/catalogclient"

	"github.com/redis/go-redis/v9"
)

func ptr[T any](v T) *T { return &v }

func event(id int64, from *string, to string, productWorkID *int64) *catalogclient.ClaimEventFeedItem {
	return &catalogclient.ClaimEventFeedItem{
		ID: id, WorkID: 900 + id, FromState: from, ToState: to,
		ProductWorkID: productWorkID, Site: "kungal",
	}
}

func TestEffectOfTransition(t *testing.T) {
	gid := ptr(int64(4321))
	cases := []struct {
		name string
		ev   *catalogclient.ClaimEventFeedItem
		want claimEffect
	}{
		{"birth into live seeds the stub", event(1, nil, catalogclient.ClaimStateLive, gid), claimEffectSeedStub},
		{"approval seeds the stub", event(2, ptr(catalogclient.ClaimStatePending), catalogclient.ClaimStateLive, gid), claimEffectSeedStub},
		{"ban unpublishes, never deletes (the children cascade)", event(3, ptr(catalogclient.ClaimStateLive), catalogclient.ClaimStateHidden, gid), claimEffectUnpublish},
		{"submit remembers the submitter", event(4, ptr(catalogclient.ClaimStateDraft), catalogclient.ClaimStatePending, gid), claimEffectRememberSubmitter},
		{"withdrawal unpublishes, never deletes", event(5, ptr(catalogclient.ClaimStateLive), catalogclient.ClaimStateDraft, gid), claimEffectUnpublish},
		{"decline unpublishes", event(6, ptr(catalogclient.ClaimStatePending), catalogclient.ClaimStateDeclined, gid), claimEffectUnpublish},
		{"no product anchor is inert", event(7, nil, catalogclient.ClaimStateLive, nil), claimEffectNone},
		{"zero product anchor is inert", event(8, nil, catalogclient.ClaimStateLive, ptr(int64(0))), claimEffectNone},
		{"an unknown state is reported, not guessed", event(9, nil, "archived", gid), claimEffectUnknownState},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := effectOf(tc.ev); got != tc.want {
				t.Errorf("effectOf = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLiveAwardCondition(t *testing.T) {
	gid := ptr(int64(11))
	awards := func(from *string) bool {
		ev := event(1, from, catalogclient.ClaimStateLive, gid)
		return isApproval(ev) || isDirectBirth(ev)
	}
	if !awards(ptr(catalogclient.ClaimStatePending)) {
		t.Error("pending → live is the approval route and must award")
	}
	if !awards(nil) {
		t.Error("a trusted submit born live must award")
	}
	if awards(ptr(catalogclient.ClaimStateDraft)) {
		t.Error("draft → live is the owner publishing; the request path already awarded it")
	}
	if awards(ptr(catalogclient.ClaimStateHidden)) {
		t.Error("unban must not award — lifting a ban does not make the admin an author")
	}
	declined := event(2, ptr(catalogclient.ClaimStatePending), catalogclient.ClaimStateDeclined, gid)
	if isApproval(declined) || isDirectBirth(declined) {
		t.Error("a decline is not a live transition")
	}
}

func TestClaimantAttribution(t *testing.T) {
	gid := ptr(int64(11))
	s := NewGalgameClaimEventSync(nil, nil, redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1}))

	publish := event(1, ptr(catalogclient.ClaimStateDraft), catalogclient.ClaimStateLive, gid)
	publish.ActorUID = 61516
	if got := s.claimantOf(t.Context(), publish, 11); got != 61516 {
		t.Errorf("owner publish: claimant = %d, want the event's actor 61516", got)
	}

	born := event(2, nil, catalogclient.ClaimStateLive, gid)
	born.ActorUID = 7
	if got := s.claimantOf(t.Context(), born, 11); got != 7 {
		t.Errorf("born live: claimant = %d, want the event's actor 7", got)
	}

	// gid 0 keeps the local-row fallback out of it: with neither a memo nor a
	// row, an approval must still refuse to name the reviewer.
	approval := event(3, ptr(catalogclient.ClaimStatePending), catalogclient.ClaimStateLive, gid)
	approval.ActorUID = 2
	if got := s.claimantOf(t.Context(), approval, 0); got != 0 {
		t.Errorf("approval without memo: claimant = %d, want 0 (never the reviewer)", got)
	}

	unban := event(4, ptr(catalogclient.ClaimStateHidden), catalogclient.ClaimStateLive, gid)
	unban.ActorUID = 2
	if got := s.claimantOf(t.Context(), unban, 11); got != 0 {
		t.Errorf("unban: claimant = %d, want 0 — lifting a ban does not make the admin the author", got)
	}
}

type claimFeedStub struct {
	mu    sync.Mutex
	total int64
	page  int
	sites []string
}

func (f *claimFeedStub) client(t *testing.T) *catalogclient.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)
		f.mu.Lock()
		f.sites = append(f.sites, r.URL.Query().Get("site"))
		f.mu.Unlock()

		items := []string{}
		for id := since + 1; id <= f.total && len(items) < f.page; id++ {
			items = append(items, fmt.Sprintf(
				`{"id":%d,"work_id":%d,"from_state":null,"to_state":"live","actor_uid":7,`+
					`"reason":null,"site":"kungal","product_work_id":%d,`+
					`"created_at":"2026-07-30T10:00:00Z"}`, id, 5000+id, 100+id))
		}
		next := since
		if n := len(items); n > 0 {
			next = since + int64(n)
		}
		_, _ = fmt.Fprintf(w, `{"code":0,"message":"成功","data":{"items":[%s],"next_since":%d}}`,
			strings.Join(items, ","), next)
	}))
	t.Cleanup(srv.Close)
	return catalogclient.New(catalogclient.Config{
		BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec",
	})
}

func TestFeedHeadWalksToTheLastEvent(t *testing.T) {
	stub := &claimFeedStub{total: 250, page: 100}
	s := NewGalgameClaimEventSync(stub.client(t), nil, nil)
	s.batch = stub.page

	head, err := s.feedHead(t.Context())
	if err != nil {
		t.Fatalf("feedHead: %v", err)
	}
	if head != 250 {
		t.Errorf("head = %d, want 250 (the last id, not the first page's)", head)
	}
	stub.mu.Lock()
	defer stub.mu.Unlock()
	for _, site := range stub.sites {
		if site != claimSite {
			t.Fatalf("feed asked with site=%q, want %q — an unfiltered feed would "+
				"hand kungal another product's claims to seed stubs for", site, claimSite)
		}
	}
}

func TestClaimFeedHeadOnEmptyFeed(t *testing.T) {
	stub := &claimFeedStub{total: 0, page: 100}
	s := NewGalgameClaimEventSync(stub.client(t), nil, nil)
	s.batch = stub.page

	head, err := s.feedHead(t.Context())
	if err != nil {
		t.Fatalf("feedHead: %v", err)
	}
	if head != 0 {
		t.Errorf("head = %d, want 0", head)
	}
}

func TestClaimNamespacesAreDistinctFromTheWikiFeed(t *testing.T) {
	if strings.Contains(claimCursorKey, "wiki:") {
		t.Errorf("cursor key %q reuses the wiki namespace", claimCursorKey)
	}
	if claimCursorKey == "wiki:msg:cron:since" {
		t.Errorf("cursor key %q is the retired wiki message cursor", claimCursorKey)
	}
}
