package handler

import (
	"kun-galgame-api/internal/galgame/service"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type CalendarHandler struct {
	calendarService *service.CalendarService
}

func NewCalendarHandler(calendarService *service.CalendarService) *CalendarHandler {
	return &CalendarHandler{calendarService: calendarService}
}

func (h *CalendarHandler) GetMonth(c fiber.Ctx) error {
	page, appErr := h.calendarService.GetMonth(c.Context(), collectQuery(c), utils.IsSFW(c))
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, page)
}

func (h *CalendarHandler) GetToday(c fiber.Ctx) error {
	flag, appErr := h.calendarService.GetTodayFlag(c.Context(), utils.IsSFW(c))
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, flag)
}

func (h *CalendarHandler) GetPending(c fiber.Ctx) error {
	page, appErr := h.calendarService.GetPending(c.Context(), collectQuery(c), utils.IsSFW(c))
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, page)
}

func (h *CalendarHandler) GetTBA(c fiber.Ctx) error {
	page, appErr := h.calendarService.GetTBA(c.Context(), collectQuery(c), utils.IsSFW(c))
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, page)
}

func (h *CalendarHandler) GetUpcoming(c fiber.Ctx) error {
	page, appErr := h.calendarService.GetUpcoming(c.Context(), utils.IsSFW(c))
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, page)
}
