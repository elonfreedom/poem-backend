package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type TagHandler struct {
	tagService *admin.AdminTagService
}

func NewTagHandler(tagService *admin.AdminTagService) *TagHandler {
	return &TagHandler{tagService: tagService}
}

// List 获取标签列表
func (h *TagHandler) List(c fuego.ContextNoBody) (*response.APIResponse[[]adminmodel.AdminTagResponse], error) {
	result, err := h.tagService.List(c.Context())
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(result), nil
}

// Create 创建标签
func (h *TagHandler) Create(c fuego.ContextWithBody[adminmodel.AdminTagCreateRequest]) (*response.APIResponse[adminmodel.AdminTagResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.tagService.Create(c.Context(), &body)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(*result), nil
}

// Delete 删除标签
func (h *TagHandler) Delete(c fuego.ContextNoBody) (*response.APIResponse[response.SimpleResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "标签ID必须是数字"}
	}

	if err := h.tagService.Delete(c.Context(), id); err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(response.SimpleResponse{Success: true}), nil
}
