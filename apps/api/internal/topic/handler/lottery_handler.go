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

type LotteryHandler struct {
	lotteryService *service.LotteryService
}

func NewLotteryHandler(lotteryService *service.LotteryService) *LotteryHandler {
	return &LotteryHandler{lotteryService: lotteryService}
}

func (h *LotteryHandler) GetLotteriesByTopic(c fiber.Ctx) error {
	var req dto.GetLotteryByTopicRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	lotteries, appErr := h.lotteryService.GetLotteriesByTopic(c.Context(), req.TopicID, middleware.GetUser(c), utils.IsSFW(c))
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, lotteries)
}

func (h *LotteryHandler) GetEntrants(c fiber.Ctx) error {
	var req struct {
		LotteryID int `query:"lottery_id" validate:"required,min=1"`
	}
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	entrants, appErr := h.lotteryService.GetEntrants(c.Context(), req.LotteryID, middleware.GetUser(c))
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, entrants)
}

func (h *LotteryHandler) CreateLottery(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.CreateLotteryRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	canModerate := perm.CanUser(user.ID, user.Roles, perm.LotteryCreateAny)
	if appErr := h.lotteryService.CreateLottery(c.Context(), user.ID, canModerate, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "抽奖创建成功")
}

func (h *LotteryHandler) UpdateLottery(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.UpdateLotteryRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	canModerate := perm.CanUser(user.ID, user.Roles, perm.LotteryManageAny)
	if appErr := h.lotteryService.UpdateLottery(c.Context(), user.ID, canModerate, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "抽奖更新成功")
}

func (h *LotteryHandler) DeleteLottery(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req struct {
		LotteryID int `query:"lottery_id" validate:"required,min=1"`
	}
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	canModerate := perm.CanUser(user.ID, user.Roles, perm.LotteryManageAny)
	if appErr := h.lotteryService.DeleteLottery(user.ID, canModerate, req.LotteryID); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "抽奖已删除")
}

func (h *LotteryHandler) Enter(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.LotteryIDRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.lotteryService.Enter(c.Context(), user, req.LotteryID); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "参与成功, 祝您好运")
}

func (h *LotteryHandler) Withdraw(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.LotteryIDRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	if appErr := h.lotteryService.Withdraw(user.ID, req.LotteryID); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "已退出抽奖")
}

func (h *LotteryHandler) Draw(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.LotteryIDRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	canModerate := perm.CanUser(user.ID, user.Roles, perm.LotteryManageAny)
	if appErr := h.lotteryService.DrawNow(c.Context(), user.ID, canModerate, req.LotteryID); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "开奖完成")
}

func (h *LotteryHandler) Cancel(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.LotteryIDRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	canModerate := perm.CanUser(user.ID, user.Roles, perm.LotteryManageAny)
	if appErr := h.lotteryService.Cancel(user.ID, canModerate, req.LotteryID); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "抽奖已取消")
}

// Claim is a POST that returns a plaintext activation code. It must never
// become a GET: Nuxt inlines every payload it fetches during SSR into the
// __NUXT__ blob, so a code fetched on page load is a code in the page source.
func (h *LotteryHandler) Claim(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.LotteryIDRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	code, appErr := h.lotteryService.ClaimCode(user.ID, req.LotteryID)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OK(c, dto.LotteryClaimResponse{Code: code})
}

func (h *LotteryHandler) SetFulfillment(c fiber.Ctx) error {
	user, appErr := middleware.MustGetUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	var req dto.LotteryFulfillRequest
	if appErr := utils.ParseAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	canModerate := perm.CanUser(user.ID, user.Roles, perm.LotteryManageAny)
	if appErr := h.lotteryService.SetFulfillment(user.ID, canModerate, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.OKMessage(c, "履约状态已更新")
}
