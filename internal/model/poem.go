package model

import (
	"time"
)

// Poem 诗歌模型
type Poem struct {
	ID           int64     `json:"id" db:"id" description:"诗歌ID（自增）"`
	Title        string    `json:"title" db:"title" description:"诗歌标题"`
	Author       string    `json:"author" db:"author" description:"作者"`
	Dynasty      string    `json:"dynasty" db:"dynasty" description:"朝代"`
	Content      string    `json:"content" db:"content" description:"原文内容"`
	Translation  string    `json:"translation" db:"translation" description:"现代译文"`
	Appreciation string    `json:"appreciation" db:"appreciation" description:"赏析"`
	CategoryID   *int64    `json:"category_id,omitempty" db:"category_id" description:"分类ID"`
	Tags         []string  `json:"tags" db:"tags" description:"标签列表"`
	CoverURL     string    `json:"cover_url" db:"cover_url" description:"封面图片URL"`
	Status       string    `json:"status" db:"status" description:"状态: draft, published, archived"`
	CreatedBy    *string   `json:"created_by,omitempty" db:"created_by" description:"创建者ID"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" description:"创建时间"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at" description:"更新时间"`
}

// PoemResponse 诗歌响应
type PoemResponse struct {
	ID           int64    `json:"id" description:"诗歌ID"`
	Title        string   `json:"title" description:"诗歌标题"`
	Author       string   `json:"author" description:"作者"`
	Dynasty      string   `json:"dynasty" description:"朝代"`
	Content      string   `json:"content" description:"原文内容"`
	Translation  string   `json:"translation,omitempty" description:"现代译文"`
	Appreciation string   `json:"appreciation,omitempty" description:"赏析"`
	Category     string   `json:"category" description:"分类名称"`
	Tags         []string `json:"tags" description:"标签列表"`
	CoverURL     string   `json:"cover_url,omitempty" description:"封面图片URL"`
	IsFavorited  bool     `json:"is_favorited" description:"是否已收藏"`
}

// PoemListItem 诗歌列表项
type PoemListItem struct {
	ID       int64  `json:"id" description:"诗歌ID"`
	Title    string `json:"title" description:"诗歌标题"`
	Author   string `json:"author" description:"作者"`
	Dynasty  string `json:"dynasty" description:"朝代"`
	Category string `json:"category" description:"分类名称"`
	CoverURL string `json:"cover_url" description:"封面图片URL"`
}

// PoemListResponse 诗歌列表响应
type PoemListResponse struct {
	Total int            `json:"total" description:"总数"`
	List  []PoemListItem `json:"list" description:"诗歌列表"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword  string `json:"keyword" validate:"required,min=1" description:"搜索关键词"`
	Page     int    `json:"page" validate:"min=1" description:"页码"`
	PageSize int    `json:"page_size" validate:"min=1,max=50" description:"每页数量"`
}
