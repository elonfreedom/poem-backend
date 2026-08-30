package user

import (
	"strconv"

	"github.com/go-fuego/fuego"

	userservice "poem-backend/internal/service/user"
	"poem-backend/pkg/response"
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
func (h *CheckinHandler) Checkin(c fuego.ContextWithBody[CheckinRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.checkinService.Checkin(c.Context(), userID, body.Date, body.PoemID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetStats 获取打卡统计
func (h *CheckinHandler) GetStats(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	result, err := h.checkinService.GetStats(c.Context(), userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetCheckinList 获取打卡记录列表（支持日期范围查询，用于热力图展示一年数据）
func (h *CheckinHandler) GetCheckinList(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	// 热力图场景：查询日期范围内所有记录，不分页（pageSize 根据日期范围计算）
	pageSize := 365
	if startDate != "" && endDate != "" {
		// 有日期范围时，返回全部记录
		pageSize = 365
	}

	result, err := h.checkinService.GetCheckinList(c.Context(), userID, 1, pageSize, startDate, endDate)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetCalendar 获取打卡日历
func (h *CheckinHandler) GetCalendar(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	year, _ := strconv.Atoi(c.QueryParam("year"))
	month, _ := strconv.Atoi(c.QueryParam("month"))

	result, err := h.checkinService.GetCalendar(c.Context(), userID, year, month)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetRanking 获取排行榜
func (h *CheckinHandler) GetRanking(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	result, err := h.checkinService.GetRanking(c.Context(), userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}
