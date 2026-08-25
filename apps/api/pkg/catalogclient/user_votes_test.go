package catalogclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVoteCover_ForwardsBearerAndParsesEnvelope(t *testing.T) {
	var gotAuth, gotPath, gotMethod string
	var gotBodyLen int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath, gotMethod, gotBodyLen = r.Header.Get("Authorization"), r.URL.Path, r.Method, r.ContentLength
		_, _ = w.Write([]byte(`{"code":0,"message":"成功","data":{"cover_id":88,"vote_count":13,"voted":true}}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"})
	res, err := c.VoteCover(context.Background(), "user-jwt", 1000, 88)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Fatalf("method = %s, want PUT", gotMethod)
	}
	if gotPath != "/v2/me/cover-votes/88" {
		t.Fatalf("bad path %q", gotPath)
	}
	if gotAuth != "Bearer user-jwt" {
		t.Fatalf("auth = %q, want the user's bearer", gotAuth)
	}
	if gotBodyLen == 0 {
		t.Fatalf("v2 vote must send {\"vote\":\"up\"}")
	}
	if res.CoverID != 88 || res.VoteCount != 13 || !res.Voted {
		t.Fatalf("envelope not parsed: %+v", res)
	}
}

func TestUnvoteCover_UsesDelete(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"code":0,"message":"成功","data":{"cover_id":88,"vote_count":12,"voted":false}}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"})
	res, err := c.UnvoteCover(context.Background(), "user-jwt", 1000, 88)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/v2/me/cover-votes/88" {
		t.Fatalf("unvote hit %s %s", gotMethod, gotPath)
	}
	if res.Voted || res.VoteCount != 12 {
		t.Fatalf("withdrawal result wrong: %+v", res)
	}
}

func TestCoverVoteErrorMapping(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"dead token", http.StatusUnauthorized, `{"code":205,"message":"invalid token"}`, ErrUnauthorized},
		{"missing scope", http.StatusForbidden, `{"code":233,"message":"missing required scope: catalog:edit"}`, ErrInsufficientScope},
		{"unknown cover", http.StatusNotFound, `{"code":233,"message":"cover not on work"}`, ErrNotFound},
		{"upstream down", http.StatusBadGateway, `nonsense`, ErrUpstream},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			c := New(Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"})
			_, err := c.VoteCover(context.Background(), "user-jwt", 1000, 88)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestCoverVoteForbiddenWithoutScopeIsGeneric(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":233,"message":"client is not bound to a catalog site"}`))
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"})
	_, err := c.VoteCover(context.Background(), "user-jwt", 1000, 88)
	if errors.Is(err, ErrInsufficientScope) {
		t.Fatal("a site-binding 403 must not read as a scope problem")
	}
	var apiErr *UserAPIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("err = %v, want a *UserAPIError 403", err)
	}
}

func TestCoverVoteWithoutTokenNeverCalls(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"})
	if _, err := c.VoteCover(context.Background(), "", 1000, 88); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
	if called {
		t.Fatal("a tokenless vote must not reach the catalog")
	}
}

func TestCoverVoteUnconfigured(t *testing.T) {
	if _, err := New(Config{}).VoteCover(context.Background(), "user-jwt", 1, 2); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err = %v, want ErrNotConfigured", err)
	}
}

func TestWorkCoverVotes(t *testing.T) {
	var gotAuth, gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath, gotQuery = r.Header.Get("Authorization"), r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(`{"code":0,"message":"成功","data":{"work":{"id":1000},"covers":[` +
			`{"id":88,"image_hash":"aa","vote_count":3,"voted":true},` +
			`{"id":89,"image_hash":"bb","vote_count":0}]}}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"})
	tallies, err := c.WorkCoverVotes(context.Background(), 1000, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/catalog/works/1000" || gotQuery != "uid=7" {
		t.Fatalf("read hit %s?%s", gotPath, gotQuery)
	}
	if len(gotAuth) < 6 || gotAuth[:6] != "Basic " {
		t.Fatalf("the tally read is S2S; auth = %q", gotAuth)
	}
	if len(tallies) != 2 || tallies[0].ID != 88 || tallies[0].ImageHash != "aa" ||
		tallies[0].VoteCount != 3 || !tallies[0].Voted || tallies[1].Voted {
		t.Fatalf("tallies decoded wrong: %+v", tallies)
	}
}

func TestWorkCoverVotesAnonymousSendsNoUID(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"code":0,"message":"成功","data":{"covers":[]}}`))
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"})
	if _, err := c.WorkCoverVotes(context.Background(), 1000, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "" {
		t.Fatalf("anonymous read sent %q", gotQuery)
	}
}
