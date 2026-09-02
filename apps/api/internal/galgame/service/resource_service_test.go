package service

import (
	"net/http"
	"testing"

	"kun-galgame-api/pkg/errors"
)

func TestClaimRefusedByState(t *testing.T) {
	conflict := errors.New(errors.CodeBiz, `cannot claim a claim in state "live"`, http.StatusConflict)
	if !claimRefusedByState(conflict) {
		t.Fatalf("claimRefusedByState(409) = false, want true — a live claim is already owned, so the silent claim is skipped")
	}
	unavailable := errors.New(errors.CodeBiz, "资料库服务暂不可用", http.StatusServiceUnavailable)
	if claimRefusedByState(unavailable) {
		t.Fatalf("claimRefusedByState(503) = true, want false — an upstream outage must still WARN")
	}
	if claimRefusedByState(nil) {
		t.Fatalf("claimRefusedByState(nil) = true, want false")
	}
}
