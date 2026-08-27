package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"kun-galgame-api/internal/galgame/client"
)

type queryRecorder struct {
	mu    sync.Mutex
	path  string
	query url.Values
}

func (r *queryRecorder) service(t *testing.T) *TagService {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.mu.Lock()
		r.path = req.URL.Path
		r.query = req.URL.Query()
		r.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"total":0,"page":1,"limit":20}`))
	}))
	t.Cleanup(srv.Close)
	return NewTagService(client.New(srv.URL, "nm_test_key", ""), &GalgameEnricher{}, nil)
}

func (r *queryRecorder) get(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.query.Get(key)
}

func (r *queryRecorder) all(key string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.query[key]
}

func (r *queryRecorder) has(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.query[key]
	return ok
}

func (r *queryRecorder) urlPath() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.path
}

func TestGetByMultiTag_ForwardsPagination(t *testing.T) {
	rec := &queryRecorder{}
	svc := rec.service(t)

	_, appErr := svc.GetByMultiTag(context.Background(), url.Values{
		"tag_ids": {"638,41"},
		"page":    {"3"},
		"limit":   {"24"},
	}, false)
	if appErr != nil {
		t.Fatalf("GetByMultiTag: %v", appErr)
	}

	if got := rec.urlPath(); got != "/v2/catalog/works" {
		t.Errorf("path = %q, want /v2/catalog/works", got)
	}
	if got := rec.all("tag_id"); len(got) != 1 || got[0] != "638,41" {
		t.Errorf("tag_id = %v, want exactly one param valued \"638,41\"", got)
	}
	// The forum still pages by number; v2 takes an opaque cursor, so the number
	// has to become one on the wire or every page serves page 1.
	if rec.has("page") {
		t.Errorf("page = %q leaked to the wire", rec.get("page"))
	}
	if got := rec.get("cursor"); got == "" {
		t.Error("cursor = empty, want the token page 3 encodes to")
	}
	if got := rec.get("limit"); got != "24" {
		t.Errorf("limit = %q, want %q", got, "24")
	}
	// CatalogCardInclude is written in the forum's v1 vocabulary and remapped at
	// the edge; v2 answers a bare id row for any block it was not asked for.
	if got := rec.get("include"); got != "titles,covers,companies" {
		t.Errorf("include = %q, want titles,covers,companies (cards render names + covers)", got)
	}
	if got := rec.get("nsfw"); got != "true" {
		t.Errorf("nsfw = %q, want true on every lane", got)
	}
	if got := rec.get("content_limit"); got != "" {
		t.Errorf("content_limit = %q, want it absent — an NSFW caller opts out of the editorial gate", got)
	}
}

func TestGetByMultiTag_SFWGateStaysClosed(t *testing.T) {
	rec := &queryRecorder{}
	svc := rec.service(t)

	if _, appErr := svc.GetByMultiTag(context.Background(),
		url.Values{"tag_ids": {"41"}}, true); appErr != nil {
		t.Fatalf("GetByMultiTag: %v", appErr)
	}
	if got := rec.get("nsfw"); got != "true" {
		t.Errorf("nsfw = %q, want true — the age gate is never a population cut", got)
	}
	if got := rec.get("content_limit"); got != "sfw" {
		t.Errorf("content_limit = %q, want sfw for an SFW caller", got)
	}
}

func TestGetByMultiTag_TagFilterShape(t *testing.T) {
	for name, tc := range map[string]struct {
		tagIDs string
		want   []string
	}{
		"one id travels alone":        {"41", []string{"41"}},
		"a multi-select is ONE param": {"638,41,7", []string{"638,41,7"}},
		"over the cap is truncated to ten": {
			"1,2,3,4,5,6,7,8,9,10,11,12",
			[]string{"1,2,3,4,5,6,7,8,9,10"},
		},
		"no selection sends no filter at all": {"", nil},
	} {
		t.Run(name, func(t *testing.T) {
			rec := &queryRecorder{}
			svc := rec.service(t)

			if _, appErr := svc.GetByMultiTag(context.Background(),
				url.Values{"tag_ids": {tc.tagIDs}}, false); appErr != nil {
				t.Fatalf("GetByMultiTag: %v", appErr)
			}

			got := rec.all("tag_id")
			if len(got) != len(tc.want) {
				t.Fatalf("tag_id = %v (%d params), want %v", got, len(got), tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("tag_id[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestGetByMultiTag_MatchMode(t *testing.T) {
	for _, mode := range []string{"contains", "exact", "", "fuzzy"} {
		t.Run("mode="+mode, func(t *testing.T) {
			rec := &queryRecorder{}
			svc := rec.service(t)

			q := url.Values{"tag_ids": {"41"}}
			if mode != "" {
				q.Set("mode", mode)
			}
			if _, appErr := svc.GetByMultiTag(context.Background(), q, false); appErr != nil {
				t.Fatalf("GetByMultiTag: %v", appErr)
			}

			if got := rec.all("tag_id"); len(got) != 1 || got[0] != "41" {
				t.Errorf("tag_id = %v, want [41] regardless of mode", got)
			}
			if rec.has("expand") {
				t.Error("expand leaked upstream; the canonical vocabulary has no DAG to expand into")
			}
			if rec.has("mode") {
				t.Error("mode leaked upstream")
			}
		})
	}
}
