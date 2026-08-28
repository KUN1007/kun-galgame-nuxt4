package catalogclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteMyClaim_TravelsAsTheUserWithAPrecondition(t *testing.T) {
	var method, path, auth, ifMatch string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		auth = r.Header.Get("Authorization")
		ifMatch = r.Header.Get("If-Match")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	if err := New(Config{BaseURL: srv.URL}).DeleteMyClaim(context.Background(), "user-jwt", 4649); err != nil {
		t.Fatalf("DeleteMyClaim: %v — 204 carries no body and must not read as a decode failure", err)
	}
	if method != http.MethodDelete || path != "/v2/me/claims/4649" {
		t.Fatalf("hit %s %s, want DELETE /v2/me/claims/4649", method, path)
	}
	if auth != "Bearer user-jwt" {
		t.Errorf("auth = %q, want the user's bearer and no Basic credential", auth)
	}
	if ifMatch != "*" {
		t.Errorf("If-Match = %q; the operation lists 412 and 428 among its responses", ifMatch)
	}
}

func TestDeleteMyClaim_SurfacesTheUpstreamRefusal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"code":"INVALID_STATE_TRANSITION","detail":"claim is live"}`))
	}))
	t.Cleanup(srv.Close)

	err := New(Config{BaseURL: srv.URL}).DeleteMyClaim(context.Background(), "user-jwt", 4649)
	var apiErr *UserAPIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusConflict {
		t.Fatalf("err = %v, want the 409 catalog raises for anything but a draft", err)
	}
}

func TestDeleteMyClaim_NeedsATokenAndABaseURL(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	t.Cleanup(srv.Close)

	if err := New(Config{BaseURL: srv.URL}).DeleteMyClaim(context.Background(), "", 1); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
	if called {
		t.Error("a tokenless user-plane call must not reach the catalog")
	}
	if err := New(Config{}).DeleteMyClaim(context.Background(), "user-jwt", 1); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("err = %v, want ErrNotConfigured", err)
	}
}
