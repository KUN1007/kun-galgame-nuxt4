package service

import (
	"slices"
	"strconv"

	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/topic/dto"
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/internal/topic/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/perm"
)

type topicViewer struct {
	ID             int
	Roles          []string
	Authenticated  bool
	ViewHidden     bool
	ViewRestricted bool
}

func readDecision(topic *topicModel.Topic, viewer topicViewer, grants []topicModel.TopicAccessGrant) bool {
	author := viewer.Authenticated && viewer.ID == topic.UserID
	if topic.Status == 1 && !author && !(viewer.Authenticated && viewer.ViewHidden) {
		return false
	}
	if author || (viewer.Authenticated && viewer.ViewRestricted) {
		return true
	}
	switch topic.AccessScope {
	case "public":
		return true
	case "login":
		return viewer.Authenticated
	case "role", "users":
		if !viewer.Authenticated {
			return false
		}
		for _, grant := range grants {
			if topic.AccessScope == "role" && grant.SubjectType == "role" && slices.Contains(viewer.Roles, grant.SubjectValue) {
				return true
			}
			if topic.AccessScope == "users" && grant.SubjectType == "user" && grant.SubjectValue == strconv.Itoa(viewer.ID) {
				return true
			}
		}
	}
	return false
}

func requireTopicRead(repo *repository.TopicRepository, topic *topicModel.Topic, user *middleware.UserInfo) ([]topicModel.TopicAccessGrant, *errors.AppError) {
	viewer := topicViewer{}
	if user != nil {
		viewer = topicViewer{ID: user.ID, Roles: user.Roles, Authenticated: true,
			ViewHidden:     perm.CanUser(user.ID, user.Roles, perm.TopicViewHidden),
			ViewRestricted: perm.CanUser(user.ID, user.Roles, perm.TopicViewRestricted)}
	}
	if topic.Status == 1 && !(viewer.Authenticated && (viewer.ID == topic.UserID || viewer.ViewHidden)) {
		return nil, errors.ErrNotFound("未找到该话题")
	}
	var grants []topicModel.TopicAccessGrant
	if topic.AccessScope == "role" || topic.AccessScope == "users" {
		var err error
		grants, err = repo.FindAccessGrants(topic.ID)
		if err != nil {
			return nil, errors.ErrInternal("获取话题权限失败")
		}
	}
	if !readDecision(topic, viewer, grants) {
		return nil, errors.ErrNotFound("未找到该话题")
	}
	return grants, nil
}

func topicDetailGrants(topic *topicModel.Topic, user *middleware.UserInfo, grants []topicModel.TopicAccessGrant) *dto.TopicAccessGrants {
	if user == nil || (user.ID != topic.UserID && !perm.CanUser(user.ID, user.Roles, perm.TopicEditAny)) {
		return nil
	}
	out := &dto.TopicAccessGrants{Roles: []string{}, UserIDs: []int{}}
	for _, grant := range grants {
		switch grant.SubjectType {
		case "role":
			out.Roles = append(out.Roles, grant.SubjectValue)
		case "user":
			if id, err := strconv.Atoi(grant.SubjectValue); err == nil {
				out.UserIDs = append(out.UserIDs, id)
			}
		}
	}
	return out
}
