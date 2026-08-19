package model

import (
	"time"
)

// Favorite 收藏模型
type Favorite struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	PoemID    int64     `json:"poem_id" db:"poem_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FavoriteResponse 收藏响应
type FavoriteResponse struct {
	ID        int64        `json:"id"`
	Poem      PoemResponse `json:"poem"`
	CreatedAt time.Time    `json:"created_at"`
}

// FavoriteRequest 收藏请求
type FavoriteRequest struct {
	PoemID int64 `json:"poem_id" validate:"required"`
}
