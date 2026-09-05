package handler

import (
	"strconv"

	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/topic/dto"
	"kun-galgame-api/internal/topic/service"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/perm"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

func (h *CommentHandler) CreateComment(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.CreateCommentRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	created, appErr := h.commentService.CreateComment(
		c.Context(), user,
		req.TopicID, req.ReplyID, req.TargetUserID, req.ParentCommentID, req.Content,
	)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, created)
}

func (h *CommentHandler) UpdateComment(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.UpdateCommentRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	updated, appErr := h.commentService.UpdateComment(c.Context(), user.ID,
		perm.CanUser(user.ID, user.Roles, perm.CommentTopicEdit), &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, updated)
}

func (h *CommentHandler) ToggleCommentLike(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.CommentInteractionRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.commentService.ToggleCommentLike(c.Context(), user.ID, req.CommentID); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "操作成功")
}

func (h *CommentHandler) DeleteComment(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	commentID, err := strconv.Atoi(c.Query("commentId"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的评论 ID"))
	}

	if appErr := h.commentService.DeleteComment(c.Context(), user.ID, perm.CanUser(user.ID, user.Roles, perm.CommentTopicDelete), commentID); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "评论已删除")
}
