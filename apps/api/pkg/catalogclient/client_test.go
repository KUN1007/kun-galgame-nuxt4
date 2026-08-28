package catalogclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfiguredPlanes(t *testing.T) {
	// The two planes have different credentials. The user plane carries the
	// reader's own token and needs nothing but a host; the application plane
	// needs the nmk_ key. Gating a user-plane call on the app key (or the other
	// way round) silently disables a working feature.
	if New(Config{}).Configured() {
		t.Error("no base URL is not configured")
	}
	if !New(Config{BaseURL: "http://x"}).Configured() {
		t.Error("the user plane needs only a base URL")
	}
	if New(Config{BaseURL: "http://x"}).AppConfigured() {
		t.Error("the application plane must not report configured without a key")
	}
	if !New(Config{BaseURL: "http://x", AppKey: "nmk_k"}).AppConfigured() {
		t.Error("base URL + application key is configured")
	}
}

func TestAppPlaneStripsAVersionSuffixFromTheBase(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path
		if auth := r.Header.Get("Authorization"); auth != "Bearer nmk_k" {
			t.Errorf("Authorization = %q, want the application key as a Bearer", auth)
		}
		_, _ = w.Write([]byte(`{"object":"list","items":[]}`))
	}))
	defer srv.Close()

	// KUN_CATALOG_API_BASE is a host with no path, but a stale value carrying a
	// version suffix would otherwise build /v2/v2/... and 404 the whole plane.
	for _, suffix := range []string{"", "/v2", "/v1", "/api/v1"} {
		c := New(Config{BaseURL: srv.URL + suffix, AppKey: "nmk_k"})
		if _, err := c.ClaimEventHead(context.Background(), "kungal"); err != nil {
			t.Fatalf("base%q: %v", suffix, err)
		}
		if got != "/v2/catalog/claim-events" {
			t.Errorf("base%q reached %q", suffix, got)
		}
	}
}

func TestAppPlaneErrorMapping(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   error
	}{
		{http.StatusUnauthorized, `{"code":"UNAUTHORIZED"}`, ErrUnauthorized},
		{http.StatusNotFound, `{"code":"NOT_FOUND"}`, ErrNotFound},
		{http.StatusInternalServerError, `{"code":"INTERNAL"}`, ErrUpstream},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		}))
		c := New(Config{BaseURL: srv.URL, AppKey: "nmk_k"})
		_, err := c.ClaimEventHead(context.Background(), "kungal")
		if !errors.Is(err, tc.want) {
			t.Errorf("status %d: want %v, got %v", tc.status, tc.want, err)
		}
		srv.Close()
	}

	// A scope refusal must stay distinguishable: it names an operator grant the
	// deployment is missing, not an outage and not the reader's session.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"status":403,"code":"SCOPE_REQUIRED",` +
			`"detail":"this operation additionally requires the claim_events:read scope."}`))
	}))
	defer srv.Close()
	var apiErr *UserAPIError
	_, err := New(Config{BaseURL: srv.URL, AppKey: "nmk_k"}).ClaimEventHead(context.Background(), "kungal")
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("err = %v, want a 403 the caller can report", err)
	}
	if !errors.Is(err, err) || apiErr.Message == "" {
		t.Error("the 403 lost the detail naming the scope")
	}
}

func TestNotConfigured(t *testing.T) {
	if _, err := New(Config{}).ClaimEventHead(context.Background(), "kungal"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}
