package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type CategoryHandler struct {
	categoryService *admin.AdminCategoryService
}

func NewCategoryHandler(categoryService *admin.AdminCategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List 获取分类列表
func (h *CategoryHandler) List(c fuego.ContextNoBody) (*response.APIResponse[[]adminmodel.AdminCategoryResponse], error) {
	result, err := h.categoryService.List(c.Context())
	if err != nil {
		return nil, fuego.InternalServerError{Title: "list failed", Detail: err.Error()}
	}
	return response.OK(result), nil
}

// Create 创建分类
func (h *CategoryHandler) Create(c fuego.ContextWithBody[adminmodel.AdminCategoryCreateRequest]) (*response.APIResponse[adminmodel.AdminCategoryResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.categoryService.Create(c.Context(), &body)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "create failed", Detail: err.Error()}
	}
	return response.OK(*result), nil
}

// Update 更新分类
func (h *CategoryHandler) Update(c fuego.ContextWithBody[adminmodel.AdminCategoryUpdateRequest]) (*response.APIResponse[any], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "分类ID必须是数字"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.categoryService.Update(c.Context(), id, &body); err != nil {
		return nil, fuego.InternalServerError{Title: "update failed", Detail: err.Error()}
	}
	return response.OK[any](nil), nil
}

// Delete 删除分类
func (h *CategoryHandler) Delete(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "分类ID必须是数字"}
	}

	if err := h.categoryService.Delete(c.Context(), id); err != nil {
		return nil, fuego.InternalServerError{Title: "delete failed", Detail: err.Error()}
	}
	return response.OK[any](nil), nil
}
