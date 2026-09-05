package handler

import (
	"github.com/gofiber/fiber/v3"
	"kun-galgame-api/internal/admin/service"
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/response"
	"strconv"
)

type TopicAdminHandler struct{ service *service.TopicAdminService }

func NewTopicAdminHandler(s *service.TopicAdminService) *TopicAdminHandler {
	return &TopicAdminHandler{service: s}
}
func (h *TopicAdminHandler) ListHidden(c fiber.Ctx) error {
	page := fiber.Query[int](c, "page", 1)
	limit := fiber.Query[int](c, "limit", 30)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	hiddenBy := c.Query("hidden_by")
	if hiddenBy != "" && hiddenBy != "author" && hiddenBy != "moderator" && hiddenBy != "trust" {
		return response.Error(c, errors.ErrBadRequest("非法的隐藏来源"))
	}
	keywords := c.Query("keywords")
	if len([]rune(keywords)) > 107 {
		return response.Error(c, errors.ErrBadRequest("关键词过长"))
	}
	r, e := h.service.ListHidden(c.Context(), page, limit, hiddenBy, keywords)
	if e != nil {
		return response.Error(c, e)
	}
	return response.OK(c, r)
}
func (h *TopicAdminHandler) PurgeStats(c fiber.Ctx) error {
	id, e := topicID(c)
	if e != nil {
		return response.Error(c, e)
	}
	r, a := h.service.PurgeStats(c.Context(), id)
	if a != nil {
		return response.Error(c, a)
	}
	return response.OK(c, r)
}
func (h *TopicAdminHandler) Delete(c fiber.Ctx) error {
	op, a := middleware.MustGetUser(c)
	if a != nil {
		return response.Error(c, a)
	}
	id, e := topicID(c)
	if e != nil {
		return response.Error(c, e)
	}
	r, a := h.service.Delete(c.Context(), op.ID, id)
	if a != nil {
		return response.Error(c, a)
	}
	return response.OK(c, r)
}
func topicID(c fiber.Ctx) (int, *errors.AppError) {
	id, err := strconv.Atoi(c.Params("tid"))
	if err != nil || id <= 0 {
		return 0, errors.ErrBadRequest("非法的话题 ID")
	}
	return id, nil
}
