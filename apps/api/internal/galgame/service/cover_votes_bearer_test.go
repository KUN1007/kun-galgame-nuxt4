package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/catalogclient"
)

type coverPlaneRecorder struct {
	mu         sync.Mutex
	path       string
	query      string
	auth       string
	staleToken bool
}

func (r *coverPlaneRecorder) server(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.URL.Path == "/v2/catalog/works" && req.URL.Query().Get("refs") != "" {
			_, _ = w.Write([]byte(`{"object":"list","items":[{"id":"1000","claimed_by":{"site":"kungal","work_id":1,"state":"live"},"refs":[{"source":"curated","external_id":"1"}]}]}`))
			return
		}

		r.mu.Lock()
		r.path, r.query, r.auth = req.URL.Path, req.URL.RawQuery, req.Header.Get("Authorization")
		stale := r.staleToken
		r.mu.Unlock()

		if stale && strings.HasPrefix(req.Header.Get("Authorization"), "Bearer ") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"code":"SCOPE_REQUIRED","title":"missing required scope: catalog:edit"}`))
			return
		}
		if strings.HasPrefix(req.URL.Path, "/api/v1/catalog/works/") {
			_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"covers":[` +
				`{"id":88,"image_hash":"abc","vote_count":3,"voted":true}]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"object":"list","items":[` +
			`{"id":"88","hash":"abc","vote_count":3,"voted":true}]}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func (r *coverPlaneRecorder) service(t *testing.T) *GalgameService {
	srv := r.server(t)
	return &GalgameService{
		galgameClient: client.New(srv.URL, "nm_test_key", ""),
		catalog:       catalogclient.New(catalogclient.Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"}),
	}
}

func TestCoverVotesReadAsTheViewerWhenTheyHaveAToken(t *testing.T) {
	rec := &coverPlaneRecorder{}
	covers := []dto.GalgameCover{{ImageHash: "abc"}}

	rec.service(t).hydrateCoverVotes(t.Context(), 1, "user-jwt", covers)

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if rec.path != "/v2/catalog/works/1000/covers" {
		t.Errorf("covers read hit %q, want the user plane's covers face", rec.path)
	}
	if rec.auth != "Bearer user-jwt" {
		t.Errorf("auth = %q, want the viewer's own bearer", rec.auth)
	}
	if strings.Contains(rec.query, "uid=") {
		t.Errorf("the viewer is the token, not a query parameter: %q", rec.query)
	}
	if covers[0].ID != 88 || covers[0].VoteCount != 3 || !covers[0].Voted {
		t.Errorf("hydration lost the tally: %+v", covers[0])
	}
}

func TestCoverVotesFallBackToPublicCountsOnAStaleToken(t *testing.T) {
	rec := &coverPlaneRecorder{staleToken: true}
	covers := []dto.GalgameCover{{ImageHash: "abc"}}

	rec.service(t).hydrateCoverVotes(t.Context(), 1, "pre-scope-jwt", covers)

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if rec.path != "/api/v1/catalog/works/1000" {
		t.Errorf("stale-token covers read landed on %q, want the S2S fallback", rec.path)
	}
	if !strings.HasPrefix(rec.auth, "Basic ") {
		t.Errorf("auth = %q, want the S2S credential after the fallback", rec.auth)
	}
	if covers[0].ID != 88 || covers[0].VoteCount != 3 {
		t.Errorf("fallback lost the tally: %+v", covers[0])
	}
}

func TestCoverVotesStayAnonymousWithoutAToken(t *testing.T) {
	rec := &coverPlaneRecorder{}
	covers := []dto.GalgameCover{{ImageHash: "abc"}}

	rec.service(t).hydrateCoverVotes(t.Context(), 1, "", covers)

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if rec.path != "/api/v1/catalog/works/1000" {
		t.Errorf("anonymous covers read hit %q, want the S2S face", rec.path)
	}
	if !strings.HasPrefix(rec.auth, "Basic ") {
		t.Errorf("auth = %q, want the S2S credential", rec.auth)
	}
	if rec.query != "" {
		t.Errorf("an anonymous read names nobody: %q", rec.query)
	}
	if covers[0].ID != 88 || covers[0].VoteCount != 3 {
		t.Errorf("hydration lost the tally: %+v", covers[0])
	}
}
