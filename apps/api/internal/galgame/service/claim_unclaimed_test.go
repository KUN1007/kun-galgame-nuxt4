package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kun-galgame-api/pkg/catalogclient"
)

type claimActionRecorder struct {
	actions []string
	bodies  []string
	fail    map[string]int
}

func (r *claimActionRecorder) catalog(t *testing.T) *catalogclient.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		action := req.URL.Path[strings.LastIndexByte(req.URL.Path, '/')+1:]
		buf := make([]byte, req.ContentLength)
		_, _ = req.Body.Read(buf)
		body := string(buf)
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/v2/me/claims":
			action = "claim"
		case req.Method == http.MethodPatch && strings.HasPrefix(req.URL.Path, "/v2/me/claims/"):
			action = "publish"
			if strings.Contains(body, `"pending"`) {
				action = "submit"
			}
			if strings.Contains(body, `"withdrawn"`) {
				action = "withdraw"
			}
		}
		r.actions = append(r.actions, action)
		r.bodies = append(r.bodies, body)

		w.Header().Set("Content-Type", "application/json")
		if status, bad := r.fail[action]; bad {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"code":"INVALID_STATE_TRANSITION","detail":"上游拒绝"}`))
			return
		}
		state := "live"
		if action == "submit" {
			state = "pending"
		}
		if action == "withdraw" {
			state = "draft"
		}
		_, _ = w.Write([]byte(`{"object":"claim","id":"4649","state":"` + state + `"}`))
	}))
	t.Cleanup(srv.Close)
	return catalogclient.New(catalogclient.Config{BaseURL: srv.URL})
}

func TestAdoptAndPublish_AnchorsTheClaimAtTheCatalogWorkID(t *testing.T) {
	rec := &claimActionRecorder{}
	res, appErr := adoptAndPublish(t.Context(), rec.catalog(t), "user-jwt", 4649)
	if appErr != nil {
		t.Fatalf("adoptAndPublish: %v", appErr)
	}
	if res.To != "live" {
		t.Errorf("to_state = %q, want live", res.To)
	}
	if len(rec.actions) != 2 || rec.actions[0] != "claim" || rec.actions[1] != "publish" {
		t.Fatalf("actions = %v, want claim then publish", rec.actions)
	}
	if !strings.Contains(rec.bodies[0], `"site_work_id":"4649"`) && !strings.Contains(rec.bodies[0], `"product_work_id":4649`) {
		t.Errorf("claim body = %q, want site_work_id 4649 — catalog refuses a claim without one",
			rec.bodies[0])
	}
}

func TestAdoptAndPublish_RetriesOverAnAlreadyAdoptedWork(t *testing.T) {
	rec := &claimActionRecorder{fail: map[string]int{"claim": http.StatusConflict}}
	if _, appErr := adoptAndPublish(t.Context(), rec.catalog(t), "user-jwt", 4649); appErr != nil {
		t.Fatalf("adoptAndPublish: %v — a refused claim over the caller's own draft must not "+
			"abort the publish, or the first failed attempt strands the work forever", appErr)
	}
	if len(rec.actions) != 2 {
		t.Errorf("actions = %v, want the publish to run anyway", rec.actions)
	}
}

func TestAdoptAndPublish_ReportsTheClaimErrorWhenBothFail(t *testing.T) {
	rec := &claimActionRecorder{fail: map[string]int{
		"claim": http.StatusForbidden, "publish": http.StatusConflict,
	}}
	_, appErr := adoptAndPublish(t.Context(), rec.catalog(t), "user-jwt", 4649)
	if appErr == nil {
		t.Fatal("adoptAndPublish: want an error when neither transaction lands")
	}
	if appErr.Message != "你没有权限执行此操作" {
		t.Errorf("message = %q, want the claim's refusal — the publish only failed because "+
			"the claim never happened, so its error describes a symptom", appErr.Message)
	}
}
