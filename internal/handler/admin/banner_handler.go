package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type BannerHandler struct {
	bannerService *admin.AdminBannerService
}

func NewBannerHandler(bannerService *admin.AdminBannerService) *BannerHandler {
	return &BannerHandler{bannerService: bannerService}
}

// List 获取 Banner 列表
func (h *BannerHandler) List(c fuego.ContextNoBody) (*response.APIResponse[[]adminmodel.AdminBannerResponse], error) {
	result, err := h.bannerService.List(c.Context())
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(result), nil
}

// Create 创建 Banner
func (h *BannerHandler) Create(c fuego.ContextWithBody[adminmodel.AdminBannerCreateRequest]) (*response.APIResponse[adminmodel.AdminBannerResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.bannerService.Create(c.Context(), &body)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(*result), nil
}

// Update 更新 Banner
func (h *BannerHandler) Update(c fuego.ContextWithBody[adminmodel.AdminBannerUpdateRequest]) (*response.APIResponse[response.SimpleResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "Banner ID必须是数字"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.bannerService.Update(c.Context(), id, &body); err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(response.SimpleResponse{Success: true}), nil
}

// Delete 删除 Banner
func (h *BannerHandler) Delete(c fuego.ContextNoBody) (*response.APIResponse[response.SimpleResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "Banner ID必须是数字"}
	}

	if err := h.bannerService.Delete(c.Context(), id); err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(response.SimpleResponse{Success: true}), nil
}
