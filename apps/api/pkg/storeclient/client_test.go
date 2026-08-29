package storeclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(Config{BaseURL: srv.URL, APIKey: "nmk_live_test"})
}

func TestPurchaseLinks_NoCampaign(t *testing.T) {
	var gotPath, gotAuth string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth = r.URL.Path, r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"object":"purchase_links","product_id":"RJ297925",
			"purchase_url":"https://s.example/s/abc","coupon_url":null,"campaign":null}`))
	})

	links, err := c.PurchaseLinks(context.Background(), "RJ297925")
	if err != nil {
		t.Fatalf("PurchaseLinks: %v", err)
	}
	if gotPath != "/v2/store/purchase-links/RJ297925" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer nmk_live_test" {
		t.Errorf("auth = %q", gotAuth)
	}
	if links.PurchaseURL != "https://s.example/s/abc" {
		t.Errorf("purchase_url = %q", links.PurchaseURL)
	}
	if links.CouponURL != "" || links.Campaign != nil {
		t.Errorf("no campaign should leave both empty, got %q / %+v", links.CouponURL, links.Campaign)
	}
}

func TestPurchaseLinks_Campaign(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"object":"purchase_links","product_id":"VJ013550",
			"purchase_url":"https://s.example/s/abc","coupon_url":"https://s.example/s/def",
			"campaign":{"object":"campaign","id":"7","name":"夏日特惠"}}`))
	})

	links, err := c.PurchaseLinks(context.Background(), "VJ013550")
	if err != nil {
		t.Fatalf("PurchaseLinks: %v", err)
	}
	if links.CouponURL != "https://s.example/s/def" {
		t.Errorf("coupon_url = %q", links.CouponURL)
	}
	// id is a string in v2, and a client that re-typed it as an integer would
	// fail the whole record rather than just that field.
	if links.Campaign == nil || links.Campaign.ID != "7" || links.Campaign.Name != "夏日特惠" {
		t.Errorf("campaign = %+v", links.Campaign)
	}
}

func TestPurchaseLinks_ProblemCodes(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		// 403 is also what a missing scope answers, so only the code separates a
		// refusal that retrying can fix from one it never will.
		{"quota", http.StatusForbidden, `{"code":"STORE_QUOTA_EXCEEDED"}`, ErrQuotaExceeded},
		{"shortener down", http.StatusBadGateway, `{"code":"STORE_LINK_UNAVAILABLE"}`, ErrUnavailable},
		{"face off", http.StatusServiceUnavailable, `{"code":"SERVICE_UNAVAILABLE"}`, ErrUnavailable},
		{"bad scope", http.StatusForbidden, `{"code":"INSUFFICIENT_SCOPE"}`, ErrUnauthorized},
		{"no key", http.StatusUnauthorized, `{"code":"MISSING_CREDENTIAL"}`, ErrUnauthorized},
		{"rejected id", http.StatusUnprocessableEntity, `{"code":"VALIDATION_FAILED"}`, ErrInvalidProduct},
		{"rate limited", http.StatusTooManyRequests, `{"code":"RATE_LIMITED"}`, ErrUpstream},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			_, err := c.PurchaseLinks(context.Background(), "RJ297925")
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestPurchaseLinks_RejectsMalformedProductLocally(t *testing.T) {
	called := false
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { called = true })
	if _, err := c.PurchaseLinks(context.Background(), "RJ12345"); !errors.Is(err, ErrInvalidProduct) {
		t.Errorf("err = %v, want ErrInvalidProduct", err)
	}
	if called {
		t.Error("a workno the face's own pattern rejects must not cost a round trip")
	}
}

func TestNotConfigured(t *testing.T) {
	if _, err := New(Config{BaseURL: "https://example.test"}).PurchaseLinks(context.Background(), "RJ297925"); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("a base URL with no key must be ErrNotConfigured, got %v", err)
	}
	var nilClient *Client
	if nilClient.Configured() {
		t.Error("nil client must report unconfigured")
	}
}

func TestOriginTrimsVersionSuffix(t *testing.T) {
	for _, in := range []string{"http://catalog:9281", "http://catalog:9281/", "http://catalog:9281/v2"} {
		if got := origin(in); got != "http://catalog:9281" {
			t.Errorf("origin(%q) = %q", in, got)
		}
	}
}
