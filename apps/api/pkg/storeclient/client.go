// Package storeclient reads the NextMoe store face at /v2/store: the DLsite
// purchase short link this site sends readers to, and the coupon link of
// whatever campaign is running.
//
// It holds its own credential rather than reusing the catalog one. The face is
// gated on the scope store:read and the v2 limiter buckets per key, so minting
// a few thousand links cannot spend the catalogue's minute budget. The links
// themselves are keyed to the OAuth client, not to the key — a second key under
// the same client shares one link namespace and one quota, which is what keeps
// per-site attribution intact.
package storeclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	ErrNotConfigured  = errors.New("storeclient: not configured (empty base URL or API key)")
	ErrUnauthorized   = errors.New("storeclient: unauthorized (key missing the store:read scope?)")
	ErrQuotaExceeded  = errors.New("storeclient: purchase-link quota exhausted for this application")
	ErrInvalidProduct = errors.New("storeclient: product id rejected by the store face")
	ErrUnavailable    = errors.New("storeclient: no link was issued")
	ErrUpstream       = errors.New("storeclient: store service error")
)

// ValidProductID mirrors the store face's own pattern. pkg/dlsite's validator is
// looser — it takes any digit count — and a workno it accepts but the face
// rejects buys a 422 round trip that can never succeed.
var validProductID = regexp.MustCompile(`^(RJ|VJ)[0-9]{6,8}$`)

func ValidProductID(s string) bool { return validProductID.MatchString(s) }

type Config struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(cfg Config) *Client {
	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    origin(cfg.BaseURL),
		apiKey:     cfg.APIKey,
		httpClient: hc,
	}
}

func (c *Client) Configured() bool {
	return c != nil && c.baseURL != "" && c.apiKey != ""
}

type Campaign struct {
	ID   string
	Name string
}

type PurchaseLinks struct {
	ProductID   string
	PurchaseURL string
	// CouponURL and Campaign are set together, and only while a campaign is
	// running. Both empty is the normal steady state, not an error.
	CouponURL string
	Campaign  *Campaign
}

func (c *Client) PurchaseLinks(ctx context.Context, productID string) (*PurchaseLinks, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	if !ValidProductID(productID) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidProduct, productID)
	}

	raw, err := c.get(ctx, "/v2/store/purchase-links/"+productID)
	if err != nil {
		return nil, err
	}

	var body struct {
		ProductID   string  `json:"product_id"`
		PurchaseURL string  `json:"purchase_url"`
		CouponURL   *string `json:"coupon_url"`
		Campaign    *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"campaign"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("%w: malformed purchase-links payload", ErrUpstream)
	}
	if body.PurchaseURL == "" {
		return nil, fmt.Errorf("%w: empty purchase_url", ErrUpstream)
	}

	out := &PurchaseLinks{ProductID: body.ProductID, PurchaseURL: body.PurchaseURL}
	if body.CouponURL != nil {
		out.CouponURL = *body.CouponURL
	}
	if body.Campaign != nil {
		out.Campaign = &Campaign{ID: body.Campaign.ID, Name: body.Campaign.Name}
	}
	return out, nil
}

func (c *Client) get(ctx context.Context, path string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode == http.StatusOK {
		return raw, nil
	}
	return nil, problemErr(resp.StatusCode, raw)
}

// The face answers problem+json, and its `code` is the only thing that
// separates a permanent refusal from a retryable one: STORE_QUOTA_EXCEEDED is a
// 403 like a missing scope, but re-sending it never turns into a link.
func problemErr(status int, raw []byte) error {
	var p struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}
	_ = json.Unmarshal(raw, &p)

	switch p.Code {
	case "STORE_QUOTA_EXCEEDED":
		return fmt.Errorf("%w: %s", ErrQuotaExceeded, p.Detail)
	case "STORE_LINK_UNAVAILABLE", "SERVICE_UNAVAILABLE":
		return fmt.Errorf("%w: %s", ErrUnavailable, p.Detail)
	case "VALIDATION_FAILED":
		return fmt.Errorf("%w: %s", ErrInvalidProduct, p.Detail)
	}

	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%w (status %d)", ErrUnauthorized, status)
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return fmt.Errorf("%w (status %d)", ErrInvalidProduct, status)
	case http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("%w (status %d)", ErrUnavailable, status)
	default:
		return fmt.Errorf("%w (status %d)", ErrUpstream, status)
	}
}

func origin(raw string) string {
	u := strings.TrimRight(strings.TrimSpace(raw), "/")
	for _, suf := range []string{"/api/v1", "/v2", "/v1"} {
		if strings.HasSuffix(u, suf) {
			return strings.TrimSuffix(u, suf)
		}
	}
	return u
}
