package service

import (
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/pkg/errors"
)

func hideDecision(topic *topicModel.Topic, userID int, canHide bool) (int, string, *errors.AppError) {
	if topic.Status == 0 {
		if topic.UserID == userID {
			return 1, "author", nil
		}
		if canHide {
			return 1, "moderator", nil
		}
		return 0, "", errors.ErrForbidden("您没有权限操作此话题")
	}
	if topic.UserID == userID {
		if topic.HiddenBy != "author" && !canHide {
			return 1, topic.HiddenBy, errors.ErrForbidden("该话题已被管理员隐藏, 无法自行取消隐藏")
		}
		return 0, "", nil
	}
	if canHide {
		return 0, "", nil
	}
	return 1, topic.HiddenBy, errors.ErrForbidden("您没有权限操作此话题")
}
