package admin

import (
	"github.com/go-fuego/fuego"
)

type BannerHandler struct{}

func NewBannerHandler() *BannerHandler {
	return &BannerHandler{}
}

// List 获取 Banner 列表
func (h *BannerHandler) List(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Create 创建 Banner
func (h *BannerHandler) Create(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Update 更新 Banner
func (h *BannerHandler) Update(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Delete 删除 Banner
func (h *BannerHandler) Delete(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}
