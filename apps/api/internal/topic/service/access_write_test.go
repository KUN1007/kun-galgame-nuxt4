package service

import (
	"reflect"
	"testing"

	topicModel "kun-galgame-api/internal/topic/model"
)

func TestNormalizeTopicAccess(t *testing.T) {
	for _, tt := range []struct {
		name, scope string
		roles       []string
		users       []int
		create      bool
		wantScope   string
		want        []topicModel.TopicAccessGrant
		invalid     bool
	}{
		{name: "create default", create: true, wantScope: "public"},
		{name: "update required", invalid: true},
		{name: "invalid scope", scope: "private", invalid: true},
		{name: "public clears", scope: "public", roles: []string{"invalid"}, users: []int{-1}, wantScope: "public"},
		{name: "login clears", scope: "login", roles: []string{"user"}, users: []int{2}, wantScope: "login"},
		{name: "roles required", scope: "role", invalid: true},
		{name: "invalid role", scope: "role", roles: []string{"guest"}, invalid: true},
		{name: "too many roles", scope: "role", roles: make([]string, 9), invalid: true},
		{name: "roles deduplicated", scope: "role", roles: []string{"creator", "creator"}, wantScope: "role", want: []topicModel.TopicAccessGrant{{SubjectType: "role", SubjectValue: "creator"}}},
		{name: "users required", scope: "users", invalid: true},
		{name: "too many users", scope: "users", users: make([]int, 51), invalid: true},
		{name: "zero user", scope: "users", users: []int{0}, invalid: true},
		{name: "negative user", scope: "users", users: []int{-2}, invalid: true},
		{name: "deduplicate and drop author", scope: "users", users: []int{1, 12, 12}, wantScope: "users", want: []topicModel.TopicAccessGrant{{SubjectType: "user", SubjectValue: "12"}}},
		{name: "author only", scope: "users", users: []int{1}, wantScope: "users"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scope, grants, err := normalizeTopicAccess(tt.scope, tt.roles, tt.users, 1, tt.create)
			if (err != nil) != tt.invalid {
				t.Fatalf("error = %v", err)
			}
			if err != nil {
				if err.StatusCode != 400 {
					t.Fatalf("status = %d", err.StatusCode)
				}
				return
			}
			if scope != tt.wantScope || (len(grants) != 0 || len(tt.want) != 0) && !reflect.DeepEqual(grants, tt.want) {
				t.Fatalf("got %q %+v", scope, grants)
			}
		})
	}
	for _, role := range []string{"user", "creator", "moderator", "admin", "ren"} {
		if _, _, err := normalizeTopicAccess("role", []string{role}, nil, 1, false); err != nil {
			t.Fatalf("role %s: %v", role, err)
		}
	}
	users := make([]int, 50)
	for i := range users {
		users[i] = i + 2
	}
	if _, grants, err := normalizeTopicAccess("users", nil, users, 1, false); err != nil || len(grants) != 50 {
		t.Fatalf("maximum users: %d %v", len(grants), err)
	}
}
