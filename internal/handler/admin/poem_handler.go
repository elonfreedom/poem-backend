package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type PoemHandler struct {
	poemService *admin.AdminPoemService
}

func NewPoemHandler(poemService *admin.AdminPoemService) *PoemHandler {
	return &PoemHandler{poemService: poemService}
}

// ImportError 单条导入错误
type ImportError struct {
	Index int    `json:"index" description:"失败记录索引"`
	Title string `json:"title" description:"诗歌标题"`
	Error string `json:"error" description:"错误原因"`
}

// BatchUpdateStatus 批量更新诗歌状态
func (h *PoemHandler) BatchUpdateStatus(c fuego.ContextWithBody[adminmodel.AdminPoemBatchUpdateStatusRequest]) (*response.APIResponse[any], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if _, err := h.poemService.BatchUpdateStatus(c.Context(), body.IDs, body.Status); err != nil {
		return nil, fuego.InternalServerError{Title: "batch update failed", Detail: err.Error()}
	}

	return response.OK[any](nil), nil
}

// ImportPoems 批量导入诗歌（JSON 数组 body）
func (h *PoemHandler) ImportPoems(c fuego.ContextWithBody[[]adminmodel.AdminPoemCreateRequest]) (*response.APIResponse[ImportResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: "请求体必须是 JSON 数组"}
	}

	result := ImportResponse{Total: len(body)}
	userID := middleware.GetUserIDFromContext(c.Context())

	for i, req := range body {
		// 校验必填字段
		if req.Title == "" || req.Author == "" || req.Content == "" || req.Status == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Index: i,
				Title: req.Title,
				Error: "缺少必填字段：title、author、content、status",
			})
			continue
		}
		if _, err := h.poemService.Create(c.Context(), &req, &userID); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Index: i,
				Title: req.Title,
				Error: err.Error(),
			})
			continue
		}
		result.Success++
	}

	return response.OK(result), nil
}

// List 获取诗歌列表
func (h *PoemHandler) List(c fuego.ContextNoBody) (*response.APIResponse[response.PageData[adminmodel.AdminPoemResponse]], error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	status := c.QueryParam("status")
	keyword := c.QueryParam("keyword")

	var categoryID *int64
	if cid, err := strconv.ParseInt(c.QueryParam("category_id"), 10, 64); err == nil {
		categoryID = &cid
	}

	result, err := h.poemService.List(c.Context(), page, pageSize, categoryID, status, keyword)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "list failed", Detail: err.Error()}
	}

	return response.PageOK(result.Items, result.Total), nil
}

// Create 创建诗歌
func (h *PoemHandler) Create(c fuego.ContextWithBody[adminmodel.AdminPoemCreateRequest]) (*response.APIResponse[adminmodel.AdminPoemResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	userID := middleware.GetUserIDFromContext(c.Context())
	result, err := h.poemService.Create(c.Context(), &body, &userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "create failed", Detail: err.Error()}
	}

	return response.OK(*result), nil
}

// GetByID 获取诗歌详情
func (h *PoemHandler) GetByID(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminPoemResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "诗歌ID必须是数字"}
	}

	result, err := h.poemService.GetByID(c.Context(), id)
	if err != nil {
		return nil, fuego.NotFoundError{Title: "not found", Detail: err.Error()}
	}

	return response.OK(*result), nil
}

// Update 更新诗歌
func (h *PoemHandler) Update(c fuego.ContextWithBody[adminmodel.AdminPoemUpdateRequest]) (*response.APIResponse[any], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "诗歌ID必须是数字"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.poemService.Update(c.Context(), id, &body); err != nil {
		return nil, fuego.InternalServerError{Title: "update failed", Detail: err.Error()}
	}

	return response.OK[any](nil), nil
}

// Delete 删除诗歌
func (h *PoemHandler) Delete(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "诗歌ID必须是数字"}
	}

	if err := h.poemService.Delete(c.Context(), id); err != nil {
		return nil, fuego.InternalServerError{Title: "delete failed", Detail: err.Error()}
	}

	return response.OK[any](nil), nil
}

// UpdateStatus 更新诗歌状态
func (h *PoemHandler) UpdateStatus(c fuego.ContextWithBody[adminmodel.AdminPoemUpdateStatusRequest]) (*response.APIResponse[any], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "诗歌ID必须是数字"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.poemService.UpdateStatus(c.Context(), id, body.Status); err != nil {
		return nil, fuego.InternalServerError{Title: "update status failed", Detail: err.Error()}
	}

	return response.OK[any](nil), nil
}

// ImportResponse 批量导入响应
type ImportResponse struct {
	Total   int            `json:"total" description:"总条数"`
	Success int            `json:"success" description:"成功数"`
	Failed  int            `json:"failed" description:"失败数"`
	Errors  []ImportError  `json:"errors" description:"失败详情"`
}
