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
	"kun-galgame-api/internal/search/dto"
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
	return NewSearchService(nil, client.New(srv.URL, "nm_test_key", ""), &galgameService.GalgameEnricher{}, nil, nil, nil, nil)
}

func (r *searchRecorder) get(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.query.Get(key)
}

func TestSearchGalgames_AsksTheCatalogWithoutAClaimGate(t *testing.T) {
	rec := &searchRecorder{}
	svc := rec.service(t)

	if _, appErr := svc.SearchGalgames(context.Background(), "恋爱", 1, 24, true, dto.GalgameFilter{}); appErr != nil {
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

	if _, appErr := svc.SearchGalgames(context.Background(), "恋爱", 1, 24, false, dto.GalgameFilter{}); appErr != nil {
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

func TestSearchGalgames_FilterRidesTheSameSearchRequest(t *testing.T) {
	rec := &searchRecorder{}
	svc := rec.service(t)

	_, appErr := svc.SearchGalgames(context.Background(), "少女", 1, 24, true, dto.GalgameFilter{
		CompanyID:    "993",
		TagIDs:       "41,638",
		ReleasedFrom: "2015",
		ReleasedTo:   "2018",
		Sort:         "released_desc",
	})
	if appErr != nil {
		t.Fatalf("SearchGalgames: %v", appErr)
	}
	if rec.path != "/v2/catalog/works" {
		t.Errorf("path = %q, want the one search request, not a second pass", rec.path)
	}
	for _, want := range []struct{ key, value string }{
		{"q", "少女"},
		{"company_id", "993"},
		{"tag_id", "41,638"},
		{"released_after", "2015-01-01"},
		{"released_before", "2018-12-31"},
		{"sort", "released_desc"},
	} {
		if got := rec.get(want.key); got != want.value {
			t.Errorf("%s = %q, want %q", want.key, got, want.value)
		}
	}
}

// A tag id list is user input on the way to a catalog query string. Anything
// that is not a positive integer is dropped rather than forwarded.
func TestSearchGalgames_DropsJunkTagIDs(t *testing.T) {
	rec := &searchRecorder{}
	svc := rec.service(t)

	_, appErr := svc.SearchGalgames(context.Background(), "少女", 1, 24, true, dto.GalgameFilter{
		TagIDs:    "41,,abc,-3,0,638",
		CompanyID: "0",
	})
	if appErr != nil {
		t.Fatalf("SearchGalgames: %v", appErr)
	}
	if got := rec.get("tag_id"); got != "41,638" {
		t.Errorf("tag_id = %q, want 41,638", got)
	}
	if got := rec.get("company_id"); got != "" {
		t.Errorf("company_id = %q, want it absent — 0 is 不限, not a company", got)
	}
}

func TestSearchGalgames_CapsTagIDsAtCatalogsLimit(t *testing.T) {
	rec := &searchRecorder{}
	svc := rec.service(t)

	_, appErr := svc.SearchGalgames(context.Background(), "少女", 1, 24, true, dto.GalgameFilter{
		TagIDs: "1,2,3,4,5,6,7,8,9,10,11,12",
	})
	if appErr != nil {
		t.Fatalf("SearchGalgames: %v", appErr)
	}
	if got := rec.get("tag_id"); got != "1,2,3,4,5,6,7,8,9,10" {
		t.Errorf("tag_id = %q, want the first ten — catalog answers 400 past that", got)
	}
}

func TestSearchGalgames_RejectsAnUnparseableYear(t *testing.T) {
	rec := &searchRecorder{}
	svc := rec.service(t)

	_, appErr := svc.SearchGalgames(context.Background(), "少女", 1, 24, true, dto.GalgameFilter{
		ReleasedFrom: "20x5",
	})
	if appErr == nil {
		t.Fatal("SearchGalgames accepted 20x5 as a year")
	}
}
