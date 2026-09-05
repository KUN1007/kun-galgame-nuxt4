package service

import (
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/pkg/errors"
	"slices"
	"strconv"
)

func normalizeTopicAccess(scope string, roles []string, userIDs []int, authorID int, create bool) (string, []topicModel.TopicAccessGrant, *errors.AppError) {
	if scope == "" && create {
		scope = "public"
	}
	grants := []topicModel.TopicAccessGrant{}
	seen := map[string]bool{}
	switch scope {
	case "public", "login":
	case "role":
		if len(roles) < 1 || len(roles) > 8 {
			return "", nil, errors.ErrBadRequest("访问角色须为 1 至 8 个")
		}
		for _, role := range roles {
			// "user" is rejected on purpose: OAuth never puts it in the roles
			// claim (docs/oauth/11-roles.md §2 — implicit default identity), so
			// a user grant matches nobody and the topic silently goes invisible
			// to everyone the author picked. The login scope is that feature.
			if !slices.Contains([]string{"creator", "moderator", "admin", "ren"}, role) {
				return "", nil, errors.ErrBadRequest("无效的访问角色")
			}
			if !seen[role] {
				grants = append(grants, topicModel.TopicAccessGrant{SubjectType: "role", SubjectValue: role})
				seen[role] = true
			}
		}
	case "users":
		if len(userIDs) < 1 || len(userIDs) > 50 {
			return "", nil, errors.ErrBadRequest("访问用户须为 1 至 50 个")
		}
		for _, id := range userIDs {
			if id <= 0 {
				return "", nil, errors.ErrBadRequest("无效的访问用户 ID")
			}
			value := strconv.Itoa(id)
			if id != authorID && !seen[value] {
				grants = append(grants, topicModel.TopicAccessGrant{SubjectType: "user", SubjectValue: value})
				seen[value] = true
			}
		}
	default:
		return "", nil, errors.ErrBadRequest("无效的话题访问范围")
	}
	return scope, grants, nil
}
