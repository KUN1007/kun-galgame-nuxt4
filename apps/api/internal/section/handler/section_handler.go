package handler

import (
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/section/dto"
	"kun-galgame-api/internal/section/service"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type SectionHandler struct {
	sectionService *service.SectionService
}

func NewSectionHandler(sectionService *service.SectionService) *SectionHandler {
	return &SectionHandler{sectionService: sectionService}
}

func (h *SectionHandler) GetSectionTopics(c fiber.Ctx) error {
	var req dto.SectionTopicsRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	resp, appErr := h.sectionService.GetSectionTopics(c.Context(), &req, middleware.GetUser(c) != nil)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, fiber.Map{
		"topics": resp.Topics,
		"total":  resp.Total,
	})
}

func (h *SectionHandler) GetCategories(c fiber.Ctx) error {
	var req dto.CategoriesRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	stats, appErr := h.sectionService.GetCategoryStats(c.Context(), req.Category)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, stats)
}
