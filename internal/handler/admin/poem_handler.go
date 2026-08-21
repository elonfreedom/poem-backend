package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/model"
	"poem-backend/internal/repository"
)

type PoemHandler struct {
	poemRepo *repository.PoemRepository
}

func NewPoemHandler(poemRepo *repository.PoemRepository) *PoemHandler {
	return &PoemHandler{poemRepo: poemRepo}
}

// ImportResponse 导入响应
type ImportResponse struct {
	Total   int      `json:"total" description:"总条数"`
	Success int      `json:"success" description:"成功数"`
	Failed  int      `json:"failed" description:"失败数"`
	Errors  []string `json:"errors" description:"错误详情"`
}

// ImportPoems 导入诗歌（JSON 文件）
func (h *PoemHandler) ImportPoems(c fuego.ContextNoBody) (*ImportResponse, error) {
	// 获取上传的文件
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid file", Detail: "请选择要上传的 JSON 文件"}
	}
	defer file.Close()

	// 检查文件类型
	if !strings.HasSuffix(header.Filename, ".json") {
		return nil, fuego.BadRequestError{Title: "invalid file type", Detail: "仅支持 JSON 文件"}
	}

	// 限制文件大小（10MB）
	if header.Size > 10*1024*1024 {
		return nil, fuego.BadRequestError{Title: "file too large", Detail: "文件大小不能超过 10MB"}
	}

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "read error", Detail: "读取文件失败"}
	}

	// 解析 JSON
	var poems []model.Poem
	if err := json.Unmarshal(data, &poems); err != nil {
		// 尝试解析为原始格式（唐诗/宋词）
		var rawPoems []map[string]any
		if err := json.Unmarshal(data, &rawPoems); err != nil {
			return nil, fuego.BadRequestError{Title: "invalid json", Detail: "JSON 格式错误，请检查文件格式"}
		}
		// 转换原始格式
		poems = convertRawPoems(rawPoems)
	}

	// 获取朝代参数
	dynasty := c.QueryParam("dynasty")

	// 批量导入
	result := ImportResponse{
		Total:  len(poems),
		Errors: []string{},
	}

	now := time.Now()
	for i, p := range poems {
		// 设置默认值
		if p.Status == "" {
			p.Status = "published"
		}
		if p.Dynasty == "" && dynasty != "" {
			p.Dynasty = dynasty
		}
		p.CreatedAt = now
		p.UpdatedAt = now

		// 验证必填字段
		if p.Title == "" || p.Author == "" || p.Content == "" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("第 %d 条：标题、作者、内容为必填项", i+1))
			continue
		}

		// 插入数据库
		if err := h.poemRepo.Create(context.Background(), &p); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("第 %d 条：%v", i+1, err))
			continue
		}
		result.Success++
	}

	return &result, nil
}

// convertRawPoems 转换原始 JSON 格式为 Poem 模型
func convertRawPoems(rawPoems []map[string]any) []model.Poem {
	poems := make([]model.Poem, 0, len(rawPoems))
	for _, raw := range rawPoems {
		poem := model.Poem{}

		// 标题
		if title, ok := raw["title"].(string); ok {
			poem.Title = title
		}
		// 宋词用 rhythmic 作为标题
		if rhythmic, ok := raw["rhythmic"].(string); ok && poem.Title == "" {
			poem.Title = rhythmic
		}

		// 作者
		if author, ok := raw["author"].(string); ok {
			poem.Author = author
		}

		// 内容（paragraphs 数组）
		if paragraphs, ok := raw["paragraphs"].([]any); ok {
			content := make([]string, 0, len(paragraphs))
			for _, p := range paragraphs {
				if s, ok := p.(string); ok {
					content = append(content, s)
				}
			}
			poem.Content = strings.Join(content, "\n")
		}

		// 标签（宋词 rhythmic）
		if rhythmic, ok := raw["rhythmic"].(string); ok {
			poem.Tags = []string{rhythmic}
		}

		poems = append(poems, poem)
	}
	return poems
}

// List 获取诗歌列表
func (h *PoemHandler) List(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Create 创建诗歌
func (h *PoemHandler) Create(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// GetByID 获取诗歌详情
func (h *PoemHandler) GetByID(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Update 更新诗歌
func (h *PoemHandler) Update(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// Delete 删除诗歌
func (h *PoemHandler) Delete(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}

// UpdateStatus 更新状态
func (h *PoemHandler) UpdateStatus(c fuego.ContextNoBody) (any, error) {
	return nil, fuego.InternalServerError{Title: "not implemented", Detail: "接口开发中"}
}
