package tools

import (
	"fmt"
	"log"

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

// ==================== 诗文去重工具 ====================

// DedupScan 扫描重复诗文组（分页返回）
// 根据 match_fields 组合对诗文分组，找出重复项，分页返回组摘要 + 组内诗文详情
func (h *ToolsHandler) DedupScan(c fuego.ContextWithBody[adminmodel.AdminToolDedupScanRequest]) (*response.APIResponse[adminmodel.AdminToolDedupScanResponse], error) {
	body, err := c.Body()
	if err != nil {
		log.Printf("[DedupScan] 请求体解析失败: %v", err)
		return nil, fuego.BadRequestError{Title: "invalid_request", Detail: fmt.Sprintf("请求体解析失败: %v", err)}
	}

	result, err := h.poemService.DedupScan(c.Context(), body.MatchFields, body.StatusFilter, body.DynastyFilter, body.Page, body.PageSize)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(*result), nil
}

// DedupExecute 执行去重（归档 + 删除）
func (h *ToolsHandler) DedupExecute(c fuego.ContextWithBody[adminmodel.AdminToolDedupExecuteRequest]) (*response.APIResponse[adminmodel.AdminToolDedupExecuteResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.poemService.DedupExecute(c.Context(), body.ArchiveIDs, body.DeleteIDs)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(*result), nil
}

// DedupMerge 智能合并重复诗文
// 将 merge_ids 中的诗文数据合并到 keep_id 对应的诗中，仅填充保留诗的空缺字段
// 合并后将被合并的诗归档（status→archived）
func (h *ToolsHandler) DedupMerge(c fuego.ContextWithBody[adminmodel.AdminToolDedupMergeRequest]) (*response.APIResponse[adminmodel.AdminToolDedupMergeResponse], error) {
	body, err := c.Body()
	if err != nil {
		log.Printf("[DedupMerge] 请求体解析失败: %v", err)
		return nil, fuego.BadRequestError{Title: "invalid_request", Detail: fmt.Sprintf("请求体解析失败: %v", err)}
	}

	result, err := h.poemService.DedupMerge(c.Context(), body.KeepID, body.MergeIDs)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(*result), nil
}

// ==================== 作者查重工具 ====================

// AuthorDedupScan 扫描重复作者组
func (h *ToolsHandler) AuthorDedupScan(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminToolAuthorDedupScanResponse], error) {
	matchBy := c.QueryParam("match_by")
	result, err := h.authorService.AuthorDedupScan(c.Context(), matchBy)
	if err != nil {
		return nil, err
	}
	return response.OK(*result), nil
}

// AuthorDedupMerge 合并重复作者
func (h *ToolsHandler) AuthorDedupMerge(c fuego.ContextWithBody[adminmodel.AdminToolAuthorDedupMergeRequest]) (*response.APIResponse[adminmodel.AdminToolAuthorDedupMergeResponse], error) {
	body, err := c.Body()
	if err != nil {
		log.Printf("[AuthorDedupMerge] 请求体解析失败: %v", err)
		return nil, fuego.BadRequestError{Title: "invalid_request", Detail: fmt.Sprintf("请求体解析失败: %v", err)}
	}

	result, err := h.authorService.AuthorDedupMerge(c.Context(), body.KeepID, body.MergeIDs)
	if err != nil {
		return nil, err
	}
	return response.OK(*result), nil
}

// ==================== 清理作者繁体名工具 ====================

// CleanupAuthorNamesResponse 清理作者繁体名响应
type CleanupAuthorNamesResponse struct {
	Cleaned int64  `json:"cleaned" description:"清理的记录数"`
	Message string `json:"message" description:"处理结果描述"`
}

// CleanupAuthorNames 清理 name = name_traditional 的作者记录
func (h *ToolsHandler) CleanupAuthorNames(c fuego.ContextNoBody) (*response.APIResponse[CleanupAuthorNamesResponse], error) {
	cleaned, message, err := h.authorService.CleanupAuthorNames(c.Context())
	if err != nil {
		return nil, err
	}
	return response.OK(CleanupAuthorNamesResponse{
		Cleaned: cleaned,
		Message: message,
	}), nil
}

// ==================== 作者姓名转简体工具 ====================

// EnsureAuthorNamesSimplifiedResponse 作者姓名转简体响应
type EnsureAuthorNamesSimplifiedResponse struct {
	Processed int64  `json:"processed" description:"处理的记录数"`
	Message   string `json:"message" description:"处理结果描述"`
}

// EnsureAuthorNamesSimplified 确保 authors 表的 name 字段为简体字
func (h *ToolsHandler) EnsureAuthorNamesSimplified(c fuego.ContextNoBody) (*response.APIResponse[EnsureAuthorNamesSimplifiedResponse], error) {
	processed, message, err := h.authorService.EnsureAuthorNamesSimplified(c.Context())
	if err != nil {
		return nil, err
	}
	return response.OK(EnsureAuthorNamesSimplifiedResponse{
		Processed: processed,
		Message:   message,
	}), nil
}

// ==================== 作者姓名转繁体工具 ====================

// ConvertAuthorNamesTraditionalResponse 作者姓名转繁体响应
type ConvertAuthorNamesTraditionalResponse struct {
	Processed int64  `json:"processed" description:"处理的记录数"`
	Message   string `json:"message" description:"处理结果描述"`
}

// ConvertAuthorNamesTraditional 将作者姓名从简体转为繁体，写入 name_traditional
func (h *ToolsHandler) ConvertAuthorNamesTraditional(c fuego.ContextNoBody) (*response.APIResponse[ConvertAuthorNamesTraditionalResponse], error) {
	processed, message, err := h.authorService.ConvertAuthorNamesToTraditional(c.Context())
	if err != nil {
		return nil, err
	}
	return response.OK(ConvertAuthorNamesTraditionalResponse{
		Processed: processed,
		Message:   message,
	}), nil
}
