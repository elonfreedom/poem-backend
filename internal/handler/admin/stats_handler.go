package admin

import (
	"github.com/go-fuego/fuego"
)

type StatsHandler struct{}

func NewStatsHandler() *StatsHandler {
	return &StatsHandler{}
}

// Overview 总览数据
func (h *StatsHandler) Overview(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Daily 每日统计
func (h *StatsHandler) Daily(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// HotPoems 热门诗歌
func (h *StatsHandler) HotPoems(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// UserGrowth 用户增长
func (h *StatsHandler) UserGrowth(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}
