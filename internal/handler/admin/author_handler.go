package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type AuthorHandler struct {
	authorService *admin.AdminAuthorService
}

func NewAuthorHandler(authorService *admin.AdminAuthorService) *AuthorHandler {
	return &AuthorHandler{authorService: authorService}
}

// List 获取作者列表
func (h *AuthorHandler) List(c fuego.ContextNoBody) (*response.APIResponse[response.PageData[adminmodel.AdminAuthorResponse]], error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	keyword := c.QueryParam("keyword")

	result, err := h.authorService.List(c.Context(), page, pageSize, keyword)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.PageOK(result.Items, result.Total), nil
}

// GetByID 获取作者详情
func (h *AuthorHandler) GetByID(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminAuthorResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "作者ID必须是数字"}
	}

	result, err := h.authorService.GetByID(c.Context(), id)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(*result), nil
}

// Create 创建作者
func (h *AuthorHandler) Create(c fuego.ContextWithBody[adminmodel.AdminAuthorCreateRequest]) (*response.APIResponse[adminmodel.AdminAuthorResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.authorService.Create(c.Context(), &body)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(*result), nil
}

// Update 更新作者
func (h *AuthorHandler) Update(c fuego.ContextWithBody[adminmodel.AdminAuthorUpdateRequest]) (*response.APIResponse[response.SimpleResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "作者ID必须是数字"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.authorService.Update(c.Context(), id, &body); err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(response.SimpleResponse{Success: true}), nil
}

// Delete 删除作者
func (h *AuthorHandler) Delete(c fuego.ContextNoBody) (*response.APIResponse[response.SimpleResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "作者ID必须是数字"}
	}

	if err := h.authorService.Delete(c.Context(), id); err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(response.SimpleResponse{Success: true}), nil
}

// Options 作者下拉搜索（用于诗歌表单）
func (h *AuthorHandler) Options(c fuego.ContextNoBody) (*response.APIResponse[[]adminmodel.AdminAuthorOptionResponse], error) {
	keyword := c.QueryParam("keyword")
	result, err := h.authorService.SearchOptions(c.Context(), keyword)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(result), nil
}

// BatchMatch 批量匹配诗歌关联作者
func (h *AuthorHandler) BatchMatch(c fuego.ContextWithBody[adminmodel.AdminAuthorBatchMatchRequest]) (*response.APIResponse[adminmodel.AdminAuthorBatchMatchResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.authorService.BatchMatchPoems(c.Context(), body.PoetryIDs)
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(*result), nil
}
