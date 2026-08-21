package handler

import (
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/website/dto"
	"kun-galgame-api/internal/website/service"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type TagGroupHandler struct {
	tagGroupService *service.TagGroupService
}

func NewTagGroupHandler(tagGroupService *service.TagGroupService) *TagGroupHandler {
	return &TagGroupHandler{tagGroupService: tagGroupService}
}

func (h *TagGroupHandler) GetWebsiteTagGroups(c fiber.Ctx) error {
	return response.OK(c, h.tagGroupService.GetAll())
}

func (h *TagGroupHandler) CreateWebsiteTagGroup(c fiber.Ctx) error {
	if _, appErr := middleware.MustGetUser(c); appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.CreateWebsiteTagGroupRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	if appErr := h.tagGroupService.Create(&req); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "标签分组创建成功")
}

func (h *TagGroupHandler) UpdateWebsiteTagGroup(c fiber.Ctx) error {
	if _, appErr := middleware.MustGetUser(c); appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.UpdateWebsiteTagGroupRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	if appErr := h.tagGroupService.Update(&req); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "标签分组更新成功")
}

func (h *TagGroupHandler) DeleteWebsiteTagGroup(c fiber.Ctx) error {
	if _, appErr := middleware.MustGetUser(c); appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.DeleteWebsiteTagGroupRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	if appErr := h.tagGroupService.Delete(req.GroupID); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "标签分组已删除")
}
