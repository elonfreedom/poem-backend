package admin

import (
	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type ConfigHandler struct {
	configService *admin.AdminConfigService
}

func NewConfigHandler(configService *admin.AdminConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

// List 获取配置列表
func (h *ConfigHandler) List(c fuego.ContextNoBody) (*response.APIResponse[[]adminmodel.AdminConfigResponse], error) {
	result, err := h.configService.List(c.Context())
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(result), nil
}

// GetByKey 获取单个配置
func (h *ConfigHandler) GetByKey(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminConfigResponse], error) {
	key := c.PathParam("key")
	if key == "" {
		return nil, fuego.BadRequestError{Title: "invalid key", Detail: "配置键不能为空"}
	}

	result, err := h.configService.GetByKey(c.Context(), key)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(*result), nil
}

// Update 更新配置
func (h *ConfigHandler) Update(c fuego.ContextWithBody[adminmodel.AdminConfigUpdateRequest]) (*response.APIResponse[response.SimpleResponse], error) {
	key := c.QueryParam("key")
	if key == "" {
		return nil, fuego.BadRequestError{Title: "invalid key", Detail: "配置键不能为空"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.configService.Update(c.Context(), key, &body); err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(response.SimpleResponse{Success: true}), nil
}
