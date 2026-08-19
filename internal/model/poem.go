package model

import (
	"time"
)

// Poem 诗歌模型
type Poem struct {
	ID           int64     `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	Author       string    `json:"author" db:"author"`
	Dynasty      string    `json:"dynasty" db:"dynasty"`
	Content      string    `json:"content" db:"content"`
	Translation  string    `json:"translation" db:"translation"`
	Appreciation string    `json:"appreciation" db:"appreciation"`
	Category     string    `json:"category" db:"category"`
	Tags         []string  `json:"tags" db:"tags"`
	CoverURL     string    `json:"cover_url" db:"cover_url"`
	Status       string    `json:"status" db:"status"` // draft, published, archived
	CreatedBy    int64     `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// PoemResponse 诗歌响应
type PoemResponse struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	Dynasty      string    `json:"dynasty"`
	Content      string    `json:"content"`
	Translation  string    `json:"translation,omitempty"`
	Appreciation string    `json:"appreciation,omitempty"`
	Category     string    `json:"category"`
	Tags         []string  `json:"tags"`
	CoverURL     string    `json:"cover_url,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// ToResponse 转换为响应格式
func (p *Poem) ToResponse() PoemResponse {
	return PoemResponse{
		ID:           p.ID,
		Title:        p.Title,
		Author:       p.Author,
		Dynasty:      p.Dynasty,
		Content:      p.Content,
		Translation:  p.Translation,
		Appreciation: p.Appreciation,
		Category:     p.Category,
		Tags:         p.Tags,
		CoverURL:     p.CoverURL,
		Status:       p.Status,
		CreatedAt:    p.CreatedAt,
	}
}

// CreatePoemRequest 创建诗歌请求
type CreatePoemRequest struct {
	Title        string   `json:"title" validate:"required"`
	Author       string   `json:"author"`
	Dynasty      string   `json:"dynasty"`
	Content      string   `json:"content" validate:"required"`
	Translation  string   `json:"translation"`
	Appreciation string   `json:"appreciation"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	CoverURL     string   `json:"cover_url"`
}

// UpdatePoemRequest 更新诗歌请求
type UpdatePoemRequest struct {
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	Dynasty      string   `json:"dynasty"`
	Content      string   `json:"content"`
	Translation  string   `json:"translation"`
	Appreciation string   `json:"appreciation"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	CoverURL     string   `json:"cover_url"`
	Status       string   `json:"status"`
}

// PoemListRequest 诗歌列表请求
type PoemListRequest struct {
	Page     int    `query:"page" default:"1"`
	PageSize int    `query:"page_size" default:"10"`
	Category string `query:"category"`
	Status   string `query:"status"`
	Keyword  string `query:"keyword"`
}
