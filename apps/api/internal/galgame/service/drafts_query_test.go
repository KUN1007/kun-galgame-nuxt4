package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"kun-galgame-api/internal/galgame/client"
)

type draftsRecorder struct {
	mu    sync.Mutex
	path  string
	query url.Values
}

func (r *draftsRecorder) service(t *testing.T) *DraftsService {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.mu.Lock()
		r.path = req.URL.Path
		r.query = req.URL.Query()
		r.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"items":[],"total":0}}`))
	}))
	t.Cleanup(srv.Close)
	return NewDraftsService(client.New(srv.URL, "nm_test_key", ""), &GalgameEnricher{})
}

func (r *draftsRecorder) get(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.query.Get(key)
}

func TestDrafts_AsksForUnclaimedWorksOnly(t *testing.T) {
	rec := &draftsRecorder{}
	svc := rec.service(t)

	if _, appErr := svc.GetDrafts(context.Background(), 2, 24, DraftFilters{}); appErr != nil {
		t.Fatalf("GetDrafts: %v", appErr)
	}
	if rec.path != "/v2/catalog/works" {
		t.Errorf("path = %q, want /v2/catalog/works", rec.path)
	}
	if got := rec.get("claimed"); got != "false" {
		t.Errorf("claimed = %q, want false — anything else lists games kungal already has", got)
	}
	if got := rec.get("page"); got != "" {
		t.Errorf("page = %q, want it rewritten to a cursor", got)
	}
	if got := rec.get("cursor"); !strings.HasPrefix(got, "cur_") {
		t.Errorf("cursor = %q, want cur_… for page 2", got)
	}
	if got := rec.get("limit"); got != "24" {
		t.Errorf("limit = %q, want 24", got)
	}
	if got := rec.get("nsfw"); got != "true" {
		t.Errorf("nsfw = %q, want true — the age gate is never a population cut", got)
	}
	if got := rec.get("content_limit"); got != "" {
		t.Errorf("content_limit = %q, want it absent on the unclaimed-works funnel", got)
	}
}

func TestDrafts_EntityScopeUsesCatalogIDs(t *testing.T) {
	for name, tc := range map[string]struct {
		filters DraftFilters
		param   string
		want    string
	}{
		"label":  {DraftFilters{LabelID: 129}, "company_id", "129"},
		"tag":    {DraftFilters{TagID: 55}, "tag_id", "55"},
		"engine": {DraftFilters{EngineID: 7}, "engine_id", "7"},
	} {
		t.Run(name, func(t *testing.T) {
			rec := &draftsRecorder{}
			svc := rec.service(t)
			if _, appErr := svc.GetDrafts(context.Background(), 1, 24, tc.filters); appErr != nil {
				t.Fatalf("GetDrafts: %v", appErr)
			}
			if got := rec.get(tc.param); got != tc.want {
				t.Errorf("%s = %q, want %q", tc.param, got, tc.want)
			}
			if got := rec.get("nsfw"); got != "true" {
				t.Errorf("nsfw = %q, want true on every lane", got)
			}
		})
	}
}
