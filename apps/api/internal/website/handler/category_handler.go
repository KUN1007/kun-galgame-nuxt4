package handler

import (
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/website/dto"
	"kun-galgame-api/internal/website/service"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) GetWebsiteCategories(c fiber.Ctx) error {
	return response.OK(c, h.categoryService.GetAll())
}

func (h *CategoryHandler) CreateWebsiteCategory(c fiber.Ctx) error {
	if _, appErr := middleware.MustGetUser(c); appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.CreateWebsiteCategoryRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	if appErr := h.categoryService.Create(&req); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "分类创建成功")
}

func (h *CategoryHandler) DeleteWebsiteCategory(c fiber.Ctx) error {
	if _, appErr := middleware.MustGetUser(c); appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.DeleteWebsiteCategoryRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	if appErr := h.categoryService.Delete(req.CategoryID); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "分类已删除")
}

func (h *CategoryHandler) GetWebsiteCategory(c fiber.Ctx) error {
	name := c.Params("name")
	detail, appErr := h.categoryService.GetDetail(name, utils.IsSFW(c))
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, detail)
}

func (h *CategoryHandler) UpdateWebsiteCategory(c fiber.Ctx) error {
	if _, appErr := middleware.MustGetUser(c); appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.UpdateWebsiteCategoryRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.categoryService.Update(&req); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "分类更新成功")
}
