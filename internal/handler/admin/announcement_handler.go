package admin

import (
	"github.com/go-fuego/fuego"
)

type AnnouncementHandler struct{}

func NewAnnouncementHandler() *AnnouncementHandler {
	return &AnnouncementHandler{}
}

// List 获取公告列表
func (h *AnnouncementHandler) List(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Create 创建公告
func (h *AnnouncementHandler) Create(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Update 更新公告
func (h *AnnouncementHandler) Update(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Delete 删除公告
func (h *AnnouncementHandler) Delete(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}
