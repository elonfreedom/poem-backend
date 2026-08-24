package user

import (
	"strconv"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	usermodel "poem-backend/internal/model/user"
	userservice "poem-backend/internal/service/user"
)

type CheckinHandler struct {
	checkinService *userservice.CheckinService
}

func NewCheckinHandler(checkinService *userservice.CheckinService) *CheckinHandler {
	return &CheckinHandler{checkinService: checkinService}
}

// CheckinRequest 打卡请求
type CheckinRequest struct {
	Date   string `json:"date" description:"打卡日期（YYYY-MM-DD，可选，默认今天）"`
	PoemID *int64 `json:"poem_id" description:"关联诗歌ID（可选）"`
}

// Checkin 打卡
func (h *CheckinHandler) Checkin(c fuego.ContextWithBody[CheckinRequest]) (*usermodel.CheckInResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.checkinService.Checkin(c.Context(), userID, body.Date, body.PoemID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// GetStats 获取打卡统计
func (h *CheckinHandler) GetStats(c fuego.ContextNoBody) (*usermodel.CheckInStatsResponse, error) {
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
func (h *CheckinHandler) GetCheckinList(c fuego.ContextNoBody) (*usermodel.CheckInListResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	result, err := h.checkinService.GetCheckinList(c.Context(), userID, 1, 30, startDate, endDate)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// GetCalendar 获取打卡日历
func (h *CheckinHandler) GetCalendar(c fuego.ContextNoBody) (*usermodel.CheckInCalendarResponse, error) {
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
func (h *CheckinHandler) GetRanking(c fuego.ContextNoBody) (*usermodel.RankingResponse, error) {
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
