package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type StatsHandler struct {
	statsService *admin.AdminStatsService
}

func NewStatsHandler(statsService *admin.AdminStatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// Overview 总览数据
func (h *StatsHandler) Overview(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminStatsOverview], error) {
	result, err := h.statsService.Overview(c.Context())
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(*result), nil
}

// Daily 每日统计
func (h *StatsHandler) Daily(c fuego.ContextNoBody) (*response.APIResponse[[]adminmodel.AdminStatsDaily], error) {
	days, _ := strconv.Atoi(c.QueryParam("days"))
	if days <= 0 || days > 365 {
		days = 30
	}

	result, err := h.statsService.Daily(c.Context(), days)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(result), nil
}

// HotPoems 热门诗歌
func (h *StatsHandler) HotPoems(c fuego.ContextNoBody) (*response.APIResponse[[]adminmodel.AdminStatsHotPoem], error) {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	result, err := h.statsService.HotPoems(c.Context(), limit)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(result), nil
}

// UserGrowth 用户增长
func (h *StatsHandler) UserGrowth(c fuego.ContextNoBody) (*response.APIResponse[[]adminmodel.AdminStatsUserGrowth], error) {
	days, _ := strconv.Atoi(c.QueryParam("days"))
	if days <= 0 || days > 365 {
		days = 30
	}

	result, err := h.statsService.UserGrowth(c.Context(), days)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(result), nil
}
