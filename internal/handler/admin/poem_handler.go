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
	poemService         *admin.AdminPoemService
	importRecordService *admin.ImportRecordService
}

func NewPoemHandler(poemService *admin.AdminPoemService) *PoemHandler {
	return &PoemHandler{poemService: poemService}
}

func NewPoemHandlerWithImportRecord(poemService *admin.AdminPoemService, importRecordService *admin.ImportRecordService) *PoemHandler {
	return &PoemHandler{poemService: poemService, importRecordService: importRecordService}
}

// ImportError 单条导入错误（失败或跳过）
type ImportError struct {
	Index int    `json:"index" description:"记录索引"`
	Title string `json:"title" description:"诗歌标题"`
	Error string `json:"error" description:"原因（失败/重复跳过）"`
}

// BatchUpdateStatus 批量更新诗歌状态
func (h *PoemHandler) BatchUpdateStatus(c fuego.ContextWithBody[adminmodel.AdminPoemBatchUpdateStatusRequest]) (*response.APIResponse[response.SimpleResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if _, err := h.poemService.BatchUpdateStatus(c.Context(), body.IDs, body.Status); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(response.SimpleResponse{Success: true}), nil
}

// ImportPoemsRequest 批量导入诗歌请求（支持顶层 source 作为默认来源）
// 兼容两种格式：
// 1. 数组格式：[{title, author, ...}, ...]（原有格式）
// 2. 对象格式：{source: "唐诗三百首", poems: [{title, author, ...}, ...]}
type ImportPoemsRequest struct {
	Source string                              `json:"source" description:"默认来源（如《唐诗三百首》），未单独指定 source 的诗文继承此值"`
	Poems  []adminmodel.AdminPoemCreateRequest `json:"poems" description:"诗歌列表"`
}

// ImportPoems 批量导入诗歌
// 支持两种请求格式：
// 1. JSON 数组：[{title, author, content, status, source?}, ...]
// 2. JSON 对象：{source: "唐诗三百首", poems: [{title, author, content, status}, ...]}
func (h *PoemHandler) ImportPoems(c fuego.ContextWithBody[ImportPoemsRequest]) (*response.APIResponse[ImportResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: "请求体格式错误"}
	}

	// 对象格式：{source, poems}
	var poems []adminmodel.AdminPoemCreateRequest
	var defaultSource string

	if len(body.Poems) > 0 {
		defaultSource = body.Source
		poems = body.Poems
	} else {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: "请求体必须包含 poems 数组"}
	}

	result := ImportResponse{Total: len(poems)}
	userID := middleware.GetUserIDFromContext(c.Context())

	// 导入开始时立即创建记录（processing 状态），超时后也能查看进度
	var recordID int64
	if h.importRecordService != nil {
		recordID, _ = h.importRecordService.Create(c.Context(), "", body.Source, len(poems), 0, 0, []adminmodel.ImportError{}, &userID)
	}

	// 用于批次内去重：标题+作者
	seen := make(map[string]bool)

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

		// 提取正文首句（第一个换行符前的内容，去除首尾空白）
		firstLine := firstContentLine(req.Content)

		// 批次内去重：标题+作者+正文首句相同视为重复
		dedupKey := req.Title + "|" + req.Author + "|" + firstLine
		if seen[dedupKey] {
			result.Skipped++
			result.Errors = append(result.Errors, ImportError{
				Index: i,
				Title: req.Title,
				Error: "批次内重复（标题+作者+首句相同），已跳过",
			})
			continue
		}
		seen[dedupKey] = true

		// 未指定 source 时使用默认值
		if req.Source == "" && defaultSource != "" {
			req.Source = defaultSource
		}

		// 数据库去重：标题+作者+正文首句已存在则跳过
		exists, err := h.poemService.ExistsByTitleAuthorFirstLine(c.Context(), req.Title, req.Author, firstLine)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Index: i,
				Title: req.Title,
				Error: "查重失败：" + err.Error(),
			})
			continue
		}
		if exists {
			result.Skipped++
			result.Errors = append(result.Errors, ImportError{
				Index: i,
				Title: req.Title,
				Error: "数据库中已存在（标题+作者+首句相同），已跳过",
			})
			continue
		}

		created, err := h.poemService.Create(c.Context(), &req, &userID)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Index: i,
				Title: req.Title,
				Error: err.Error(),
			})
			continue
		}
		result.Success++
		if created != nil {
			result.IDs = append(result.IDs, created.ID)
		}
	}

	// 导入结束后更新记录为最终状态（失败不影响主流程）
	if h.importRecordService != nil && recordID > 0 {
		errors := make([]adminmodel.ImportError, len(result.Errors))
		for i, e := range result.Errors {
			errors[i] = adminmodel.ImportError{Index: e.Index, Title: e.Title, Error: e.Error}
		}
		status := computeImportStatus(result.Success, result.Failed)
		_ = h.importRecordService.UpdateStatus(c.Context(), recordID, result.Success, result.Failed, status, errors)
	}

	return response.OK(result), nil
}

// computeImportStatus 根据成功/失败数计算导入状态
func computeImportStatus(success, failed int) string {
	switch {
	case failed == 0:
		return "success"
	case success == 0:
		return "failed"
	default:
		return "partial"
	}
}

// List 获取诗歌列表
func (h *PoemHandler) List(c fuego.ContextNoBody) (*response.APIResponse[response.PageData[adminmodel.AdminPoemResponse]], error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	status := c.QueryParam("status")
	keyword := c.QueryParam("keyword")
	dynasty := c.QueryParam("dynasty")

	var categoryID *int64
	if cid, err := strconv.ParseInt(c.QueryParam("category_id"), 10, 64); err == nil {
		categoryID = &cid
	}

	var authorID *int64
	if aid, err := strconv.ParseInt(c.QueryParam("author_id"), 10, 64); err == nil {
		authorID = &aid
	}

	result, err := h.poemService.List(c.Context(), page, pageSize, categoryID, status, keyword, dynasty, authorID)
	if err != nil {
		return nil, err // 透传 Service 错误
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
		return nil, err // 透传 Service 错误
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
		return nil, err // 透传 Service 错误
	}

	return response.OK(*result), nil
}

// Update 更新诗歌
func (h *PoemHandler) Update(c fuego.ContextWithBody[adminmodel.AdminPoemUpdateRequest]) (*response.APIResponse[response.SimpleResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "诗歌ID必须是数字"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.poemService.Update(c.Context(), id, &body); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(response.SimpleResponse{Success: true}), nil
}

// Delete 删除诗歌
func (h *PoemHandler) Delete(c fuego.ContextNoBody) (*response.APIResponse[response.SimpleResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "诗歌ID必须是数字"}
	}

	if err := h.poemService.Delete(c.Context(), id); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(response.SimpleResponse{Success: true}), nil
}

// UpdateStatus 更新诗歌状态
func (h *PoemHandler) UpdateStatus(c fuego.ContextWithBody[adminmodel.AdminPoemUpdateStatusRequest]) (*response.APIResponse[response.SimpleResponse], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "诗歌ID必须是数字"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.poemService.UpdateStatus(c.Context(), id, body.Status); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(response.SimpleResponse{Success: true}), nil
}

// firstContentLine 提取正文首句（第一个换行符前的非空内容）
func firstContentLine(content string) string {
	for i, c := range content {
		if c == '\n' || c == '\r' {
			return content[:i]
		}
	}
	return content
}

// ImportResponse 批量导入响应
type ImportResponse struct {
	Total   int           `json:"total" description:"总条数"`
	Success int           `json:"success" description:"成功数"`
	Skipped int           `json:"skipped" description:"跳过数（重复）"`
	Failed  int           `json:"failed" description:"失败数"`
	Errors  []ImportError `json:"errors" description:"跳过/失败详情"`
	IDs    []int64       `json:"ids" description:"成功导入的诗歌 ID 列表"`
}
