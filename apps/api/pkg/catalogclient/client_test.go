package catalogclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestConfigured(t *testing.T) {
	if New(Config{BaseURL: "http://x", ClientID: "", ClientSecret: ""}).Configured() {
		t.Fatal("expected not configured without credentials")
	}
	if New(Config{BaseURL: "", ClientID: "a", ClientSecret: "b"}).Configured() {
		t.Fatal("expected not configured without base URL")
	}
	if !New(Config{BaseURL: "http://x", ClientID: "a", ClientSecret: "b"}).Configured() {
		t.Fatal("expected configured with base URL + credentials")
	}
}

func TestNotConfigured(t *testing.T) {
	if _, err := New(Config{}).getData(context.Background(), "/api/v1/catalog/ping", nil); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func TestGetData_ForwardsBasicAndPassesThrough(t *testing.T) {
	var gotAuth, gotPath, gotQuery string
	body := `{"work_id":3853,"groups":[{"role_id":1,"role_key":"scenario","role_name":"剧本","credits":[{"credit_name_id":42,"name":"丸戸史明"}]}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"code":0,"message":"成功","data":` + body + `}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"})
	data, err := c.getData(context.Background(), "/api/v2/catalog/works/3853/credits",
		url.Values{"limit": {"10"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v2/catalog/works/3853/credits" {
		t.Fatalf("bad path %q", gotPath)
	}
	if gotQuery != "limit=10" {
		t.Fatalf("bad query %q", gotQuery)
	}
	if len(gotAuth) < 6 || gotAuth[:6] != "Basic " {
		t.Fatalf("expected Basic auth, got %q", gotAuth)
	}
	var got any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("returned data is not valid json: %v", err)
	}
	if string(data) != body {
		t.Fatalf("data not passed through verbatim:\n got %s\nwant %s", data, body)
	}
}

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   error
	}{
		{http.StatusUnauthorized, `{"code":10001,"message":"bad creds"}`, ErrUnauthorized},
		{http.StatusNotFound, `{"code":4,"message":"not found"}`, ErrNotFound},
		{http.StatusInternalServerError, `{"code":1,"message":"boom"}`, ErrUpstream},
		{http.StatusOK, `{"code":1,"message":"logical failure"}`, ErrUpstream},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		}))
		c := New(Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"})
		_, err := c.getData(context.Background(), "/api/v2/catalog/works/1/credits", nil)
		if !errors.Is(err, tc.want) {
			t.Fatalf("status %d: want %v, got %v", tc.status, tc.want, err)
		}
		srv.Close()
	}
}

func TestUpstreamUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	deadURL := srv.URL
	srv.Close()
	c := New(Config{BaseURL: deadURL, ClientID: "cid", ClientSecret: "sec"})
	if _, err := c.getData(context.Background(), "/api/v2/catalog/works/1/credits", nil); !errors.Is(err, ErrUpstream) {
		t.Fatalf("want ErrUpstream, got %v", err)
	}
}
