package handler

import (
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/topic/dto"
	"kun-galgame-api/internal/topic/service"
	"kun-galgame-api/pkg/perm"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type PollHandler struct {
	pollService *service.PollService
}

func NewPollHandler(pollService *service.PollService) *PollHandler {
	return &PollHandler{pollService: pollService}
}

func (h *PollHandler) CreatePoll(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.CreatePollRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.pollService.CreatePoll(c.Context(), user.ID, perm.CanUser(user.ID, user.Roles, perm.PollCreateAny), &req); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "投票创建成功")
}

func (h *PollHandler) UpdatePoll(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.UpdatePollRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.pollService.UpdatePoll(c.Context(), user.ID, perm.CanUser(user.ID, user.Roles, perm.PollEditAny), &req); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "投票更新成功")
}

func (h *PollHandler) GetPollsByTopic(c fiber.Ctx) error {
	var req dto.GetPollByTopicRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	userInfo := middleware.GetUser(c)

	polls, appErr := h.pollService.GetPollsByTopic(c.Context(), req.TopicID, userInfo)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, polls)
}

func (h *PollHandler) Vote(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	_ = user

	var req dto.VoteRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.pollService.Vote(c.Context(), user, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "投票成功")
}

func (h *PollHandler) DeletePoll(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req struct {
		PollID int `query:"poll_id" validate:"required,min=1"`
	}
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.pollService.DeletePoll(c.Context(), user.ID, perm.CanUser(user.ID, user.Roles, perm.PollDeleteAny), req.PollID); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "投票已删除")
}

func (h *PollHandler) GetVoteLog(c fiber.Ctx) error {
	var req dto.GetPollLogRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	userInfo := middleware.GetUser(c)

	entries, total, appErr := h.pollService.GetVoteLog(c.Context(), req.PollID, req.Page, req.Limit, userInfo)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, fiber.Map{"logs": entries, "total": total})
}
