package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

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

// ImportPoems 导入诗歌（JSON 文件）
func (h *PoemHandler) ImportPoems(c fuego.ContextNoBody) (*response.APIResponse[ImportResponse], error) {
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid file", Detail: "请选择要上传的 JSON 文件"}
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".json") {
		return nil, fuego.BadRequestError{Title: "invalid file type", Detail: "仅支持 JSON 文件"}
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "read error", Detail: "读取文件失败"}
	}

	var poems []adminmodel.AdminPoemCreateRequest
	if err := json.Unmarshal(data, &poems); err != nil {
		var rawPoems []map[string]any
		if err := json.Unmarshal(data, &rawPoems); err != nil {
			return nil, fuego.BadRequestError{Title: "invalid json", Detail: "JSON 格式错误"}
		}
		poems = convertRawPoems(rawPoems)
	}

	result := ImportResponse{Total: len(poems)}
	userID := middleware.GetUserIDFromContext(c.Context())

	for i, req := range poems {
		if req.Title == "" || req.Author == "" || req.Content == "" {
			result.Failed++
			result.Errors = append(result.Errors, "第"+strconv.Itoa(i+1)+"条：标题、作者、内容为必填项")
			continue
		}
		if _, err := h.poemService.Create(c.Context(), &req, &userID); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, "第"+strconv.Itoa(i+1)+"条："+err.Error())
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

// ImportResponse 导入响应
type ImportResponse struct {
	Total   int      `json:"total" description:"总条数"`
	Success int      `json:"success" description:"成功数"`
	Failed  int      `json:"failed" description:"失败数"`
	Errors  []string `json:"errors" description:"错误详情"`
}

// convertRawPoems 转换原始 JSON 格式为 AdminPoemCreateRequest
func convertRawPoems(rawPoems []map[string]any) []adminmodel.AdminPoemCreateRequest {
	poems := make([]adminmodel.AdminPoemCreateRequest, 0, len(rawPoems))
	for _, raw := range rawPoems {
		req := adminmodel.AdminPoemCreateRequest{}
		if title, ok := raw["title"].(string); ok {
			req.Title = title
		}
		if rhythmic, ok := raw["rhythmic"].(string); ok && req.Title == "" {
			req.Title = rhythmic
		}
		if author, ok := raw["author"].(string); ok {
			req.Author = author
		}
		if paragraphs, ok := raw["paragraphs"].([]any); ok {
			var content []string
			for _, p := range paragraphs {
				if s, ok := p.(string); ok {
					content = append(content, s)
				}
			}
			req.Content = strings.Join(content, "\n")
		}
		if rhythmic, ok := raw["rhythmic"].(string); ok {
			req.Tags = []string{rhythmic}
		}
		poems = append(poems, req)
	}
	return poems
}

// 避免 fmt 未使用
var _ = fmt.Sprintf
