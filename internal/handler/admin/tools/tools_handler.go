package tools

import (
	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/convert"
	"poem-backend/pkg/response"
)

// ToolsHandler 工具模块处理器
type ToolsHandler struct {
	poemService   *admin.AdminPoemService
	authorService *admin.AdminAuthorService
}

// NewToolsHandler 创建工具模块处理器
func NewToolsHandler(poemService *admin.AdminPoemService, authorService *admin.AdminAuthorService) *ToolsHandler {
	return &ToolsHandler{
		poemService:   poemService,
		authorService: authorService,
	}
}

// BatchConvertSimplifiedResponse 批量生成简体响应
type BatchConvertSimplifiedResponse struct {
	Processed int `json:"processed" description:"成功处理记录数"`
}

// BatchConvertSimplified 一键为存量诗歌生成简体（繁体 → 简体）
// 扫描 title_sc/author_sc/content_sc 为空的记录，自动生成简体字段
func (h *ToolsHandler) BatchConvertSimplified(c fuego.ContextNoBody) (*response.APIResponse[BatchConvertSimplifiedResponse], error) {
	processed, err := h.poemService.EnsureSimplifiedForAllPoems(c.Context())
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(BatchConvertSimplifiedResponse{Processed: processed}), nil
}

// BatchGeneratePinyinResponse 批量生成拼音响应
type BatchGeneratePinyinResponse struct {
	Processed int `json:"processed" description:"成功处理记录数"`
}

// BatchGeneratePinyin 一键为存量诗歌生成拼音
// 扫描 title_pinyin 为空的记录，自动生成拼音字段（带声调）
func (h *ToolsHandler) BatchGeneratePinyin(c fuego.ContextNoBody) (*response.APIResponse[BatchGeneratePinyinResponse], error) {
	processed, err := h.poemService.EnsurePinyinForAllPoems(c.Context())
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(BatchGeneratePinyinResponse{Processed: processed}), nil
}

// GenerateAuthorsFromPoems 从已有诗歌作品中提取所有不重复的作者名，自动创建作者记录
// 扫描 poems 表的 author 字段，去重后插入 authors 表（跳过已存在的）
func (h *ToolsHandler) GenerateAuthorsFromPoems(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminToolGenerateAuthorsResponse], error) {
	result, err := h.authorService.GenerateAuthorsFromPoems(c.Context())
	if err != nil {
		return nil, err // 透传 Service 错误
	}
	return response.OK(*result), nil
}

// ==================== 简繁体工具 ====================

// DetectCharsType 检测文本的中文字符类型（简体/繁体/混合/未知）
func (h *ToolsHandler) DetectCharsType(c fuego.ContextWithBody[adminmodel.AdminToolDetectCharsTypeRequest]) (*response.APIResponse[adminmodel.AdminToolDetectCharsTypeResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	charType := convert.DetectCharsType(body.Text)
	return response.OK(adminmodel.AdminToolDetectCharsTypeResponse{Type: string(charType)}), nil
}

// ConvertChars 将文本转换为简体或繁体
func (h *ToolsHandler) ConvertChars(c fuego.ContextWithBody[adminmodel.AdminToolConvertCharsRequest]) (*response.APIResponse[adminmodel.AdminToolConvertCharsResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := convert.ConvertChars(body.Text, body.Target)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "convert failed", Detail: err.Error()}
	}

	return response.OK(adminmodel.AdminToolConvertCharsResponse{Text: result}), nil
}

// BatchConvertChars 批量转换指定诗歌的字符类型
func (h *ToolsHandler) BatchConvertChars(c fuego.ContextWithBody[adminmodel.AdminToolBatchConvertCharsRequest]) (*response.APIResponse[adminmodel.AdminToolBatchConvertCharsResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.poemService.BatchConvertChars(c.Context(), body.PoetryIDs, body.Target)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(*result), nil
}
