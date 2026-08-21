package model

import (
	"time"
)

// Category 分类模型
type Category struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Sort      int       `json:"sort" db:"sort"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CategoryResponse 分类响应
type CategoryResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

// Tag 标签模型
type Tag struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TagResponse 标签响应
type TagResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
