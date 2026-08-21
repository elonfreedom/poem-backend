package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type AnnouncementHandler struct {
	announcementService *admin.AdminAnnouncementService
}

func NewAnnouncementHandler(announcementService *admin.AdminAnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{announcementService: announcementService}
}

// List 获取公告列表
func (h *AnnouncementHandler) List(c fuego.ContextNoBody) (*response.APIResponse[[]adminmodel.AdminAnnouncementResponse], error) {
	result, err := h.announcementService.List(c.Context())
	if err != nil {
		return nil, fuego.InternalServerError{Title: "list failed", Detail: err.Error()}
	}
	return response.OK(result), nil
}

// Create 创建公告
func (h *AnnouncementHandler) Create(c fuego.ContextWithBody[adminmodel.AdminAnnouncementCreateRequest]) (*response.APIResponse[adminmodel.AdminAnnouncementResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.announcementService.Create(c.Context(), &body)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "create failed", Detail: err.Error()}
	}
	return response.OK(*result), nil
}

// Update 更新公告
func (h *AnnouncementHandler) Update(c fuego.ContextWithBody[adminmodel.AdminAnnouncementUpdateRequest]) (*response.APIResponse[any], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "公告ID必须是数字"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.announcementService.Update(c.Context(), id, &body); err != nil {
		return nil, fuego.InternalServerError{Title: "update failed", Detail: err.Error()}
	}
	return response.OK[any](nil), nil
}

// Delete 删除公告
func (h *AnnouncementHandler) Delete(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "公告ID必须是数字"}
	}

	if err := h.announcementService.Delete(c.Context(), id); err != nil {
		return nil, fuego.InternalServerError{Title: "delete failed", Detail: err.Error()}
	}
	return response.OK[any](nil), nil
}
