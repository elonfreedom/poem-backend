package user

import (
	"strconv"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	"poem-backend/internal/model"
	userservice "poem-backend/internal/service/user"
)

type CheckinHandler struct {
	checkinService *userservice.CheckinService
}

func NewCheckinHandler(checkinService *userservice.CheckinService) *CheckinHandler {
	return &CheckinHandler{checkinService: checkinService}
}

// Checkin 打卡
func (h *CheckinHandler) Checkin(c fuego.ContextNoBody) (*model.CheckInResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	result, err := h.checkinService.Checkin(c.Context(), userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// GetStats 获取打卡统计
func (h *CheckinHandler) GetStats(c fuego.ContextNoBody) (*model.CheckInStatsResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	result, err := h.checkinService.GetStats(c.Context(), userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// GetCheckinList 获取打卡记录列表
func (h *CheckinHandler) GetCheckinList(c fuego.ContextNoBody) (*model.CheckInListResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	result, err := h.checkinService.GetCheckinList(c.Context(), userID, 1, 30)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// GetCalendar 获取打卡日历
func (h *CheckinHandler) GetCalendar(c fuego.ContextNoBody) (*model.CheckInCalendarResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	year, _ := strconv.Atoi(c.QueryParam("year"))
	month, _ := strconv.Atoi(c.QueryParam("month"))

	result, err := h.checkinService.GetCalendar(c.Context(), userID, year, month)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// GetRanking 获取排行榜
func (h *CheckinHandler) GetRanking(c fuego.ContextNoBody) (*model.RankingResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	result, err := h.checkinService.GetRanking(c.Context(), userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}
