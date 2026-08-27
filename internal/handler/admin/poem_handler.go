package admin

import (
	"encoding/json"
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

// ImportPoemsRequest 批量导入诗歌请求（支持顶层 source 作为默认来源）
// 兼容两种格式：
// 1. 数组格式：[{title, author, ...}, ...]（原有格式）
// 2. 对象格式：{source: "唐诗三百首", poems: [{title, author, ...}, ...]}
type ImportPoemsRequest struct {
	Source string                           `json:"source" description:"默认来源（如《唐诗三百首》），未单独指定 source 的诗文继承此值"`
	Poems  []adminmodel.AdminPoemCreateRequest `json:"poems" description:"诗歌列表"`
}

// ImportPoems 批量导入诗歌
// 支持两种请求格式：
// 1. JSON 数组：[{title, author, content, status, source?}, ...]
// 2. JSON 对象：{source: "唐诗三百首", poems: [{title, author, content, status}, ...]}
func (h *PoemHandler) ImportPoems(c fuego.ContextWithBody[any]) (*response.APIResponse[ImportResponse], error) {
	// 读取原始 body 字节
	rawBody, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: "请求体格式错误"}
	}

	// 将 body 转为 map 判断格式
	bodyMap, ok := rawBody.(map[string]any)
	if !ok {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: "请求体必须是 JSON 数组或对象"}
	}

	var poems []adminmodel.AdminPoemCreateRequest
	var defaultSource string

	// 判断是对象格式还是数组格式
	if poemsRaw, exists := bodyMap["poems"]; exists {
		// 对象格式：{source, poems}
		if source, ok := bodyMap["source"].(string); ok {
			defaultSource = source
		}
		// 解析 poems 数组
		poemsJSON, err := json.Marshal(poemsRaw)
		if err != nil {
			return nil, fuego.BadRequestError{Title: "invalid poems", Detail: "poems 字段格式错误"}
		}
		if err := json.Unmarshal(poemsJSON, &poems); err != nil {
			return nil, fuego.BadRequestError{Title: "invalid poems", Detail: "poems 数组解析失败"}
		}
	} else {
		// 尝试解析为数组（原有格式）
		reqBody, err := json.Marshal(rawBody)
		if err != nil {
			return nil, fuego.BadRequestError{Title: "invalid body", Detail: "请求体解析失败"}
		}
		// 尝试解析为数组
		if err := json.Unmarshal(reqBody, &poems); err != nil {
			// 尝试解析为对象格式
			var req ImportPoemsRequest
			if err := json.Unmarshal(reqBody, &req); err != nil {
				return nil, fuego.BadRequestError{Title: "invalid body", Detail: "请求体必须是 JSON 数组或 {source, poems} 对象"}
			}
			defaultSource = req.Source
			poems = req.Poems
		}
	}

	result := ImportResponse{Total: len(poems)}
	userID := middleware.GetUserIDFromContext(c.Context())

	for i, req := range poems {
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
		// 未指定 source 时使用默认值
		if req.Source == "" && defaultSource != "" {
			req.Source = defaultSource
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
	Total   int           `json:"total" description:"总条数"`
	Success int           `json:"success" description:"成功数"`
	Failed  int           `json:"failed" description:"失败数"`
	Errors  []ImportError `json:"errors" description:"失败详情"`
}
