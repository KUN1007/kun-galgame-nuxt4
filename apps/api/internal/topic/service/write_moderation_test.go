package service

import (
	"context"
	"kun-galgame-api/internal/middleware"
	"testing"

	"kun-galgame-api/internal/topic/dto"
	"kun-galgame-api/internal/trust/gate"
	"kun-galgame-api/pkg/trustclient"
)

type scriptedChecker struct{ decision string }

func (c scriptedChecker) Check(_ context.Context, _ trustclient.CheckRequest) (*trustclient.CheckResult, error) {
	return &trustclient.CheckResult{Decision: c.decision, Matched: []string{"坏词"}}, nil
}

func TestCreateTopicDeniedNothingPersisted(t *testing.T) {
	svc := NewTopicWriteService(
		nil, nil, nil, nil, nil, nil,
		gate.NewCheckService(scriptedChecker{decision: gate.DecisionDeny}),
		gate.NewScanService(nil),
	)
	id, appErr := svc.Create(context.Background(), 7, &dto.CreateTopicRequest{
		Title: "标题", Content: "正文", Category: "others", Sections: []string{"闲聊"},
	})
	if appErr == nil {
		t.Fatal("deny must return an error")
	}
	if appErr.StatusCode != 422 {
		t.Fatalf("deny status = %d, want 422", appErr.StatusCode)
	}
	if id != 0 {
		t.Fatalf("deny must persist nothing, got id %d", id)
	}
}

func TestCreateCommentDeniedNothingPersisted(t *testing.T) {
	svc := NewCommentService(
		nil, nil, nil, nil, nil,
		gate.NewCheckService(scriptedChecker{decision: gate.DecisionDeny}),
		gate.NewScanService(nil),
	)
	resp, appErr := svc.CreateComment(context.Background(), &middleware.UserInfo{ID: 7}, 1, 1, 2, nil, "违禁内容")
	if appErr == nil {
		t.Fatal("deny must return an error")
	}
	if appErr.StatusCode != 422 {
		t.Fatalf("deny status = %d, want 422", appErr.StatusCode)
	}
	if resp != nil {
		t.Fatalf("deny must persist nothing, got %+v", resp)
	}
}
