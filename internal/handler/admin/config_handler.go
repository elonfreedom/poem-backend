package admin

import (
	"github.com/go-fuego/fuego"
)

type ConfigHandler struct{}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// List 获取配置列表
func (h *ConfigHandler) List(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// GetByKey 获取单个配置
func (h *ConfigHandler) GetByKey(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Update 更新配置
func (h *ConfigHandler) Update(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}
