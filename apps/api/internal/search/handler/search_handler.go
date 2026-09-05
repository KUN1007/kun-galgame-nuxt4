package handler

import (
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/search/dto"
	"kun-galgame-api/internal/search/service"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) QuickSearch(c fiber.Ctx) error {
	var req dto.QuickSearchRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	res, appErr := h.searchService.QuickSearch(c.Context(), req.Keywords, middleware.GetUser(c) != nil)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, res)
}

func (h *SearchHandler) Overview(c fiber.Ctx) error {
	var req dto.OverviewRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	res, appErr := h.searchService.Overview(c.Context(), req.Keywords, utils.IsSFW(c), middleware.GetUser(c) != nil)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, res)
}

func (h *SearchHandler) SearchEntities(c fiber.Ctx) error {
	var req dto.EntitySearchRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	groups, appErr := h.searchService.SearchEntities(
		c.Context(), req.Keywords, req.Family, req.Page, req.Limit, utils.IsSFW(c),
	)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	res := dto.EntitySearchResult{Groups: groups}
	for _, group := range groups {
		res.Total += group.Total
	}
	return response.OK(c, res)
}

func (h *SearchHandler) Search(c fiber.Ctx) error {
	var req dto.SearchRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	switch req.Type {
	case "topic":
		res, appErr := h.searchService.SearchTopics(c.Context(), req.Keywords, req.Page, req.Limit, middleware.GetUser(c) != nil)
		if appErr != nil {
			return response.Error(c, appErr)
		}
		return response.Paginated(c, res.Items, res.Total)
	case "galgame":
		res, appErr := h.searchService.SearchGalgames(
			c.Context(), req.Keywords, req.Page, req.Limit, false,
		)
		if appErr != nil {
			return response.Error(c, appErr)
		}
		return response.Paginated(c, res.Items, res.Total)
	case "resource":
		res, appErr := h.searchService.SearchResources(
			c.Context(), req.Keywords, req.Page, req.Limit, utils.IsSFW(c),
		)
		if appErr != nil {
			return response.Error(c, appErr)
		}
		return response.Paginated(c, res.Items, res.Total)
	case "user":
		res, appErr := h.searchService.SearchUsers(c.Context(), req.Keywords, req.Page, req.Limit, middleware.GetUser(c) != nil)
		if appErr != nil {
			return response.Error(c, appErr)
		}
		return response.Paginated(c, res.Items, res.Total)
	case "reply":
		res, appErr := h.searchService.SearchReplies(c.Context(), req.Keywords, req.Page, req.Limit, middleware.GetUser(c) != nil)
		if appErr != nil {
			return response.Error(c, appErr)
		}
		return response.Paginated(c, res.Items, res.Total)
	case "comment":
		res, appErr := h.searchService.SearchComments(c.Context(), req.Keywords, req.Page, req.Limit, middleware.GetUser(c) != nil)
		if appErr != nil {
			return response.Error(c, appErr)
		}
		return response.Paginated(c, res.Items, res.Total)
	default:
		return response.OK(c, []any{})
	}
}
