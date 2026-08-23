package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
)

func TestCatalogLibraryRequest_DefaultBrowseUsesTheCatalog(t *testing.T) {
	if !catalogLibraryRequest(&dto.GalgameListRequest{Page: 1, Limit: 24}) {
		t.Fatal("default /galgame must be the catalog library")
	}
	if catalogLibraryRequest(&dto.GalgameListRequest{Indexed: true}) {
		t.Fatal("indexed=1 is the sitemap, not the library")
	}
	if catalogLibraryRequest(&dto.GalgameListRequest{Type: "game"}) {
		t.Fatal("a resource-type filter is the local resource list")
	}
	if catalogLibraryRequest(&dto.GalgameListRequest{SortField: "view"}) {
		t.Fatal("view sort is local")
	}
	if !catalogLibraryRequest(&dto.GalgameListRequest{SortField: "popularity"}) {
		t.Fatal("popularity is the catalog library sort")
	}
}

type libraryRecorder struct {
	mu    sync.Mutex
	path  string
	query map[string][]string
}

func (r *libraryRecorder) get(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.query == nil {
		return ""
	}
	v := r.query[key]
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func TestCatalogLibrary_DoesNotSendClaimState(t *testing.T) {
	rec := &libraryRecorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec.mu.Lock()
		rec.path = req.URL.Path
		rec.query = req.URL.Query()
		rec.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"items":[],"total":0}}`))
	}))
	t.Cleanup(srv.Close)

	svc := NewGalgameService(nil, nil, nil, nil, nil, nil, nil,
		client.New(srv.URL, "nm_test_key", ""), nil, nil, "", "")
	page, appErr := svc.GetList(context.Background(), &dto.GalgameListRequest{
		Page: 1, Limit: 24, SortField: "popularity", SortOrder: "desc",
	}, true)
	if appErr != nil {
		t.Fatalf("GetList: %v", appErr)
	}
	if page.Total != 0 {
		t.Errorf("total = %d, want 0 from empty upstream", page.Total)
	}
	if rec.path != "/v1/catalog/works/search" {
		t.Errorf("path = %q, want /v1/catalog/works/search", rec.path)
	}
	if got := rec.get("claim_state"); got != "" {
		t.Errorf("claim_state = %q, want it absent", got)
	}
	if got := rec.get("sort"); got != "popularity" {
		t.Errorf("sort = %q, want popularity", got)
	}
	if got := rec.get("nsfw"); got != "1" {
		t.Errorf("nsfw = %q, want 1", got)
	}
	if got := rec.get("content_limit"); got != "sfw" {
		t.Errorf("content_limit = %q, want sfw", got)
	}
}
