package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type CheckinHandler struct {
	checkinService *admin.AdminCheckinService
}

func NewCheckinHandler(checkinService *admin.AdminCheckinService) *CheckinHandler {
	return &CheckinHandler{checkinService: checkinService}
}

// List 获取打卡记录列表
func (h *CheckinHandler) List(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminCheckinListResponse], error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	keyword := c.QueryParam("keyword")
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	result, err := h.checkinService.ListCheckins(c.Context(), page, pageSize, keyword, startDate, endDate)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(*result), nil
}

// Stats 获取打卡数据统计
func (h *CheckinHandler) Stats(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminCheckinStats], error) {
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	result, err := h.checkinService.GetCheckinStats(c.Context(), startDate, endDate)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(*result), nil
}
