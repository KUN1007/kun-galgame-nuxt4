package service

import (
	"encoding/json"
	"strings"
	"testing"

	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/topic/dto"
	topicModel "kun-galgame-api/internal/topic/model"
)

func TestReadDecision(t *testing.T) {
	grants := []topicModel.TopicAccessGrant{
		{SubjectType: "role", SubjectValue: "creator"},
		{SubjectType: "user", SubjectValue: "4"},
	}
	viewers := []struct {
		name    string
		viewer  topicViewer
		allowed [4]bool
	}{
		{"anonymous", topicViewer{}, [4]bool{true, false, false, false}},
		{"unrelated login", topicViewer{ID: 2, Authenticated: true, Roles: []string{"user"}}, [4]bool{true, true, false, false}},
		{"second role granted", topicViewer{ID: 3, Authenticated: true, Roles: []string{"user", "creator"}}, [4]bool{true, true, true, false}},
		{"user granted", topicViewer{ID: 4, Authenticated: true}, [4]bool{true, true, false, true}},
		{"author", topicViewer{ID: 1, Authenticated: true}, [4]bool{true, true, true, true}},
		{"restricted permission", topicViewer{ID: 5, Authenticated: true, ViewRestricted: true}, [4]bool{true, true, true, true}},
		{"hidden permission", topicViewer{ID: 6, Authenticated: true, ViewHidden: true}, [4]bool{true, true, false, false}},
		{"both permissions", topicViewer{ID: 7, Authenticated: true, ViewHidden: true, ViewRestricted: true}, [4]bool{true, true, true, true}},
		{"hidden role granted", topicViewer{ID: 8, Authenticated: true, ViewHidden: true, Roles: []string{"creator"}}, [4]bool{true, true, true, false}},
		{"hidden user granted", topicViewer{ID: 4, Authenticated: true, ViewHidden: true}, [4]bool{true, true, false, true}},
	}
	for i, scope := range []string{"public", "login", "role", "users"} {
		for _, tt := range viewers {
			for _, hidden := range []bool{false, true} {
				name := scope + "/" + tt.name
				if hidden {
					name += "/hidden"
				}
				t.Run(name, func(t *testing.T) {
					topic := &topicModel.Topic{UserID: 1, AccessScope: scope}
					want := tt.allowed[i]
					if hidden {
						topic.Status = 1
						want = want && (tt.viewer.ID == 1 || tt.viewer.ViewHidden)
					}
					if got := readDecision(topic, tt.viewer, grants); got != want {
						t.Fatalf("allowed = %v, want %v", got, want)
					}
				})
			}
		}
	}
}

func TestReadDecisionGrantTypesAndDecimalIDs(t *testing.T) {
	viewer := topicViewer{ID: 4, Authenticated: true, Roles: []string{"creator"}}
	for _, tt := range []struct{ scope, kind, value string }{
		{"users", "user", "5"}, {"users", "user", "04"}, {"users", "role", "4"},
		{"role", "user", "creator"}, {"role", "role", "admin"}, {"unknown", "user", "4"},
	} {
		t.Run(tt.scope+"/"+tt.kind+"/"+tt.value, func(t *testing.T) {
			if readDecision(&topicModel.Topic{UserID: 1, AccessScope: tt.scope}, viewer, []topicModel.TopicAccessGrant{{SubjectType: tt.kind, SubjectValue: tt.value}}) {
				t.Fatal("unexpected grant match")
			}
		})
	}
}

func TestRequireTopicReadWithoutGrants(t *testing.T) {
	for _, tt := range []struct {
		name, scope string
		status      int
		viewer      *middleware.UserInfo
		denied      bool
	}{
		{"public anonymous", "public", 0, nil, false},
		{"login anonymous", "login", 0, nil, true},
		{"login viewer", "login", 0, &middleware.UserInfo{ID: 2}, false},
		{"hidden public", "public", 1, nil, true},
		{"hidden role", "role", 1, nil, true},
		{"hidden users", "users", 1, nil, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := requireTopicRead(nil, &topicModel.Topic{UserID: 1, AccessScope: tt.scope, Status: tt.status}, tt.viewer)
			if (err != nil) != tt.denied {
				t.Fatalf("error = %v", err)
			}
			if err != nil && err.StatusCode != 404 {
				t.Fatalf("status = %d", err.StatusCode)
			}
		})
	}
}

func TestTopicDetailGrantsPrivacy(t *testing.T) {
	topic := &topicModel.Topic{UserID: 1, AccessScope: "users"}
	grants := []topicModel.TopicAccessGrant{{SubjectType: "user", SubjectValue: "2"}, {SubjectType: "role", SubjectValue: "creator"}}
	for _, tt := range []struct {
		name    string
		viewer  *middleware.UserInfo
		visible bool
	}{
		{"anonymous", nil, false},
		{"granted viewer", &middleware.UserInfo{ID: 2, Roles: []string{"creator"}}, false},
		{"author", &middleware.UserInfo{ID: 1}, true},
		{"editor", &middleware.UserInfo{ID: 3, Roles: []string{"admin"}}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			detail := dto.TopicDetail{AccessScope: topic.AccessScope, AccessGrants: topicDetailGrants(topic, tt.viewer, grants)}
			data, err := json.Marshal(detail)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data), `"access_grants"`) != tt.visible {
				t.Fatalf("unexpected grants visibility: %s", data)
			}
			if !strings.Contains(string(data), `"access_scope":"users"`) {
				t.Fatal("missing access_scope")
			}
			if tt.visible && (len(detail.AccessGrants.UserIDs) != 1 || detail.AccessGrants.UserIDs[0] != 2 || len(detail.AccessGrants.Roles) != 1) {
				t.Fatalf("grants = %+v", detail.AccessGrants)
			}
		})
	}
}
