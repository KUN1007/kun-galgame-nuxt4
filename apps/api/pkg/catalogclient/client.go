package catalogclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	AppKey       string
	HTTPClient   *http.Client
}

type Client struct {
	basicAuth  string
	appKey     string
	baseURL    string
	httpClient *http.Client
}

func New(cfg Config) *Client {
	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	var ba string
	if cfg.ClientID != "" && cfg.ClientSecret != "" {
		ba = "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.ClientID+":"+cfg.ClientSecret))
	}
	return &Client{
		basicAuth:  ba,
		appKey:     strings.TrimSpace(cfg.AppKey),
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		httpClient: hc,
	}
}

func (c *Client) Configured() bool { return c.baseURL != "" && c.basicAuth != "" }

func (c *Client) AppConfigured() bool { return c.baseURL != "" && c.appKey != "" }

var (
	ErrNotConfigured = errors.New("catalogclient: not configured (empty base URL or credentials)")
	ErrUnauthorized  = errors.New("catalogclient: unauthorized (check client_id/secret)")
	ErrNotFound      = errors.New("catalogclient: entity not found")
	ErrUpstream      = errors.New("catalogclient: catalog service error")
)

func (c *Client) getData(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	reqURL := c.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.basicAuth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("%w (status %d)", ErrUpstream, resp.StatusCode)
	}

	var env struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("%w: malformed envelope", ErrUpstream)
	}
	if env.Code != 0 {
		return nil, fmt.Errorf("%w (code %d): %s", ErrUpstream, env.Code, env.Message)
	}
	return env.Data, nil
}
