package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/pkg/catalogclient"
	"kun-galgame-api/pkg/errors"
)

type claimPlaneRecorder struct {
	mu     sync.Mutex
	path   string
	auth   string
	body   map[string]any
	claimQ url.Values

	status  int
	message string
}

func (r *claimPlaneRecorder) server(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		raw, _ := io.ReadAll(req.Body)

		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(req.URL.Path, "/catalog/lookup/batch"):
			var in struct {
				Items []struct {
					ExternalID string `json:"external_id"`
				} `json:"items"`
			}
			_ = json.Unmarshal(raw, &in)
			out := make([]string, 0, len(in.Items))
			for _, it := range in.Items {
				out = append(out, `{"external_id":"`+it.ExternalID+`","work":null}`)
			}
			_, _ = w.Write([]byte(`{"object":"list","items":[` +
				strings.Join(out, ",") + `],"missing":[]}`))
			return
		case strings.HasSuffix(req.URL.Path, "/catalog/works"):
			_, _ = w.Write([]byte(`{"object":"list","items":[` +
				`{"object":"work","id":"90210","claim":{"site":"kungal","site_work_id":"90210","state":"draft"}}` +
				`],"next_cursor":null}`))
			return
		}

		r.mu.Lock()
		r.path = req.URL.Path
		r.auth = req.Header.Get("Authorization")
		if req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/v2/me/claims") {
			r.claimQ = req.URL.Query()
		} else {
			_ = json.Unmarshal(raw, &r.body)
		}
		status, message := r.status, r.message
		r.mu.Unlock()

		if status != 0 {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"code":233,"message":"` + message + `","data":null}`))
			return
		}
		if req.Method == http.MethodGet && req.URL.Path == "/v2/me/claims" {
			_, _ = w.Write([]byte(`{"object":"list","items":[]}`))
			return
		}
		if req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/v2/me/claims/") {
			w.Header().Set("ETag", `"c90210"`)
			_, _ = w.Write([]byte(`{"object":"claim","id":"90210","state":"draft"}`))
			return
		}
		_, _ = w.Write([]byte(`{` +
			`"work_id":90210,"from_state":"draft","to_state":"pending","event_id":7}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func (r *claimPlaneRecorder) submissionService(t *testing.T) *SubmissionService {
	srv := r.server(t)
	return NewSubmissionService(
		client.New(srv.URL, "nm_test_key", ""),
		catalogclient.New(catalogclient.Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"}),
		nil,
	)
}

func (r *claimPlaneRecorder) reviewService(t *testing.T) *ClaimReviewService {
	srv := r.server(t)
	return NewClaimReviewService(
		client.New(srv.URL, "nm_test_key", ""),
		catalogclient.New(catalogclient.Config{BaseURL: srv.URL, ClientID: "cid", ClientSecret: "sec"}),
	)
}

func TestOwnerActionSpeaksAsTheSubmitter(t *testing.T) {
	rec := &claimPlaneRecorder{}
	svc := rec.submissionService(t)

	if _, appErr := svc.Withdraw(t.Context(), "user-jwt", 90210); appErr != nil {
		t.Fatalf("Withdraw: %v", appErr)
	}

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if rec.path != "/v2/me/claims/90210" {
		t.Errorf("withdraw hit %q, want PATCH /v2/me/claims/{id}", rec.path)
	}
	if rec.auth != "Bearer user-jwt" {
		t.Errorf("auth = %q, want the submitter's bearer alone", rec.auth)
	}
	if _, ok := rec.body["actor"]; ok {
		t.Errorf("an owner action must assert no actor: %v", rec.body)
	}
	if _, ok := rec.body["site"]; ok {
		t.Errorf("an owner action must assert no site: %v", rec.body)
	}
}

func TestReviewVerdictSpeaksAsTheModerator(t *testing.T) {
	rec := &claimPlaneRecorder{}
	svc := rec.reviewService(t)

	if _, appErr := svc.Review(t.Context(), "mod-jwt", 90210,
		catalogclient.ClaimActionDecline, "资料不足"); appErr != nil {
		t.Fatalf("Review: %v", appErr)
	}

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if rec.path != "/v2/moderation/claims/90210/decisions" {
		t.Errorf("decline hit %q, want POST /v2/moderation/claims/{id}/decisions", rec.path)
	}
	if rec.auth != "Bearer mod-jwt" {
		t.Errorf("auth = %q, want the moderator's own bearer", rec.auth)
	}
	if _, ok := rec.body["actor"]; ok {
		t.Errorf("a verdict must assert no actor: %v", rec.body)
	}
	if rec.body["note"] != "资料不足" && rec.body["reason"] != "资料不足" {
		t.Errorf("note/reason = %v, want it recorded on the event", rec.body)
	}
}

func TestListMineIsTheTokensOwnClaims(t *testing.T) {
	rec := &claimPlaneRecorder{}
	svc := rec.submissionService(t)

	page, appErr := svc.ListMine(t.Context(), "user-jwt", url.Values{})
	if appErr != nil {
		t.Fatalf("ListMine: %v", appErr)
	}

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if rec.path != "/v2/me/claims" {
		t.Errorf("ListMine hit %q, want GET /v2/me/claims", rec.path)
	}
	if rec.auth != "Bearer user-jwt" {
		t.Errorf("auth = %q, want the caller's bearer", rec.auth)
	}
	if got := rec.claimQ.Get("site"); got != "" {
		t.Errorf("site = %q, want it absent — the tenant rides the token", got)
	}
	if got := rec.claimQ.Get("claim_state"); got != "pending,declined,draft" {
		t.Errorf("claim_state = %q, want the 我的提交 default", got)
	}
	if got := rec.claimQ.Get("kind"); got != catalogclient.ClaimKindSubmitted {
		t.Errorf("kind = %q, want only the works the caller submitted", got)
	}
	if page.Items == nil || len(page.Items) != 0 {
		t.Errorf("items = %v, want an empty array", page.Items)
	}
}

func TestListAuditIsTheClaimsTheCallerReviewed(t *testing.T) {
	rec := &claimPlaneRecorder{}
	svc := rec.submissionService(t)

	if _, appErr := svc.ListAudit(t.Context(), "mod-jwt", url.Values{}); appErr != nil {
		t.Fatalf("ListAudit: %v", appErr)
	}

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if got := rec.claimQ.Get("kind"); got != catalogclient.ClaimKindAudited {
		t.Errorf("kind = %q, want only the works the caller reviewed", got)
	}
	if got := rec.claimQ.Get("claim_state"); got != "" {
		t.Errorf("claim_state = %q, want every state — an audit history hides nothing", got)
	}
}

func TestClaimErrorsCarryTheTokenTaxonomy(t *testing.T) {
	t.Run("a token minted before catalog:edit asks for a re-login", func(t *testing.T) {
		rec := &claimPlaneRecorder{status: http.StatusForbidden, message: "missing required scope: catalog:edit"}
		svc := rec.submissionService(t)

		_, appErr := svc.Resubmit(t.Context(), "old-jwt", 90210)
		if appErr == nil {
			t.Fatal("want a refusal")
		}
		if appErr.Code != errors.CodeReauthRequired {
			t.Errorf("code = %d, want %d (re-login prompt)", appErr.Code, errors.CodeReauthRequired)
		}
	})

	t.Run("a dead session is auth-expired, not an upstream fault", func(t *testing.T) {
		rec := &claimPlaneRecorder{status: http.StatusUnauthorized, message: "invalid token"}
		svc := rec.submissionService(t)

		_, appErr := svc.Resubmit(t.Context(), "dead-jwt", 90210)
		if appErr == nil {
			t.Fatal("want a refusal")
		}
		if appErr.Code != errors.CodeAuth || appErr.StatusCode != http.StatusUnauthorized {
			t.Errorf("error = %+v, want the auth-expired envelope", appErr)
		}
	})

	t.Run("a real denial stays a denial", func(t *testing.T) {
		rec := &claimPlaneRecorder{status: http.StatusForbidden, message: "not the owner of this work"}
		svc := rec.reviewService(t)

		_, appErr := svc.Review(t.Context(), "mod-jwt", 90210, catalogclient.ClaimActionApprove, "")
		if appErr == nil {
			t.Fatal("want a refusal")
		}
		if appErr.StatusCode != http.StatusForbidden || appErr.Code == errors.CodeReauthRequired {
			t.Errorf("error = %+v, want a plain 403 — re-logging in would not help", appErr)
		}
	})
}
