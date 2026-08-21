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
