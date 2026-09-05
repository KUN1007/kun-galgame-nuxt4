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

type TopicHandler struct {
	topicService      *service.TopicService
	topicWriteService *service.TopicWriteService
}

func NewTopicHandler(
	topicService *service.TopicService,
	topicWriteService *service.TopicWriteService,
) *TopicHandler {
	return &TopicHandler{
		topicService:      topicService,
		topicWriteService: topicWriteService,
	}
}

func (h *TopicHandler) MyInteractions(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, h.topicService.GetMyInteractions(user.ID))
}

func (h *TopicHandler) GetList(c fiber.Ctx) error {
	var req dto.ListTopicsRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if req.SortField == "" {
		req.SortField = "status_update_time"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	isNSFW := !utils.IsSFW(c)

	items, total, appErr := h.topicService.GetList(c.Context(), &req, isNSFW, middleware.GetUser(c) != nil)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, dto.TopicListResponse{Topics: items, Total: total})
}

func (h *TopicHandler) GetResourceList(c fiber.Ctx) error {
	var req dto.ListTopicsRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if req.SortField == "" {
		req.SortField = "status_update_time"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	isNSFW := !utils.IsSFW(c)
	items, _, appErr := h.topicService.GetResourceList(c.Context(), &req, isNSFW, middleware.GetUser(c) != nil)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, items)
}

func (h *TopicHandler) GetDetail(c fiber.Ctx) error {
	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的话题 ID"))
	}

	userInfo := middleware.GetUser(c)

	detail, appErr := h.topicService.GetDetail(c.Context(), tid, userInfo)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, detail)
}

func (h *TopicHandler) GetTopicReactionHistory(c fiber.Ctx) error {
	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的话题 ID"))
	}

	records, appErr := h.topicService.GetTopicReactionHistory(c.Context(), tid)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, records)
}

func (h *TopicHandler) Create(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.CreateTopicRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	topicID, appErr := h.topicWriteService.Create(c.Context(), user.ID, &req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, topicID)
}

func (h *TopicHandler) Update(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无��的话题 ID"))
	}

	var req dto.UpdateTopicRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.topicWriteService.Update(c.Context(), user.ID, perm.CanUser(user.ID, user.Roles, perm.TopicEditAny), tid, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "话题更新成功")
}

func (h *TopicHandler) ToggleLike(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的话题 ID"))
	}

	if appErr := h.topicWriteService.ToggleLike(c.Context(), user.ID, tid); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "操作成功")
}

func (h *TopicHandler) ToggleReaction(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的话题 ID"))
	}

	var req dto.ReactionRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.topicWriteService.ToggleReaction(c.Context(), user.ID, tid, req.Reaction); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "操作成功")
}

func (h *TopicHandler) ToggleDislike(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无��的话题 ID"))
	}

	if appErr := h.topicWriteService.ToggleDislike(c.Context(), user.ID, tid); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "��作成功")
}

func (h *TopicHandler) Upvote(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效��话�� ID"))
	}

	var body struct {
		Description string `json:"description"`
	}
	_ = c.Bind().Body(&body)

	if appErr := h.topicWriteService.Upvote(c.Context(), user.ID, tid, body.Description); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "推话题成功")
}

func (h *TopicHandler) GetUpvotes(c fiber.Ctx) error {
	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的话题 ID"))
	}

	records, appErr := h.topicService.GetTopicUpvotes(c.Context(), tid)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OK(c, records)
}

func (h *TopicHandler) ToggleFavorite(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的话题 ID"))
	}

	if appErr := h.topicWriteService.ToggleFavorite(c.Context(), user.ID, tid); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "操作成功")
}

func (h *TopicHandler) ToggleHide(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("���效的话题 ID"))
	}

	if appErr := h.topicWriteService.ToggleHide(c.Context(), user.ID, perm.CanUser(user.ID, user.Roles, perm.TopicHide), tid); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "操作成功")
}

func (h *TopicHandler) SetBestAnswer(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	tid, err := strconv.Atoi(c.Params("tid"))
	if err != nil {
		return response.Error(c, errors.ErrBadRequest("无效的话�� ID"))
	}

	var req dto.BestAnswerRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.topicWriteService.SetBestAnswer(c.Context(), user.ID, perm.CanUser(user.ID, user.Roles, perm.TopicSetBestAnswer), tid, req.ReplyID); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.OKMessage(c, "已设置最佳回答")
}
