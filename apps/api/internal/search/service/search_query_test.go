package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"kun-galgame-api/internal/galgame/client"
	galgameService "kun-galgame-api/internal/galgame/service"
)

type searchRecorder struct {
	mu    sync.Mutex
	path  string
	query url.Values
}

func (r *searchRecorder) service(t *testing.T) *SearchService {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.mu.Lock()
		r.path = req.URL.Path
		r.query = req.URL.Query()
		r.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"total":0}`))
	}))
	t.Cleanup(srv.Close)
	return NewSearchService(nil, client.New(srv.URL, "nm_test_key", ""), &galgameService.GalgameEnricher{}, nil)
}

func (r *searchRecorder) get(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.query.Get(key)
}

func TestSearchGalgames_AsksTheCatalogWithoutAClaimGate(t *testing.T) {
	rec := &searchRecorder{}
	svc := rec.service(t)

	if _, appErr := svc.SearchGalgames(context.Background(), "恋爱", 1, 24, true); appErr != nil {
		t.Fatalf("SearchGalgames: %v", appErr)
	}
	if rec.path != "/v2/catalog/works" {
		t.Errorf("path = %q, want /v2/catalog/works", rec.path)
	}
	if got := rec.get("claim_state"); got != "" {
		t.Errorf("claim_state = %q, want it absent — catalog is the existence layer", got)
	}
	if got := rec.get("q"); got != "恋爱" {
		t.Errorf("q = %q, want the raw keywords", got)
	}
	if got := rec.get("sort"); got != "relevance" {
		t.Errorf("sort = %q, want relevance", got)
	}
	if got := rec.get("nsfw"); got != "true" {
		t.Errorf("nsfw = %q, want true — the age gate is never a population cut", got)
	}
	if got := rec.get("content_limit"); got != "sfw" {
		t.Errorf("content_limit = %q, want sfw for an SFW caller", got)
	}
}

func TestSearchGalgames_NSFWCallerStillHasNoClaimGate(t *testing.T) {
	rec := &searchRecorder{}
	svc := rec.service(t)

	if _, appErr := svc.SearchGalgames(context.Background(), "恋爱", 1, 24, false); appErr != nil {
		t.Fatalf("SearchGalgames: %v", appErr)
	}
	if got := rec.get("nsfw"); got != "true" {
		t.Errorf("nsfw = %q, want true for an NSFW-opted caller", got)
	}
	if got := rec.get("content_limit"); got != "" {
		t.Errorf("content_limit = %q, want it absent — an NSFW caller opts out of the editorial gate", got)
	}
	if got := rec.get("claim_state"); got != "" {
		t.Errorf("claim_state = %q, want it absent for an NSFW caller too", got)
	}
}
