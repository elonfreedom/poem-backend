package model

import (
	"time"
)

// Favorite 收藏模型（复合主键：user_id + poem_id）
type Favorite struct {
	UserID    string    `json:"user_id" db:"user_id" description:"用户ID (UUID v7)"`
	PoemID    int64     `json:"poem_id" db:"poem_id" description:"诗歌ID"`
	CreatedAt time.Time `json:"created_at" db:"created_at" description:"收藏时间"`
}

// FavoriteResponse 收藏响应
type FavoriteResponse struct {
	Poem      PoemListItem `json:"poem" description:"诗歌信息"`
	CreatedAt time.Time    `json:"created_at" description:"收藏时间"`
}

// FavoriteRequest 收藏请求
type FavoriteRequest struct {
	PoemID int64 `json:"poem_id" validate:"required,min=1" description:"诗歌ID"`
}

// FavoriteListResponse 收藏列表响应
type FavoriteListResponse struct {
	Total int                `json:"total" description:"总数"`
	List  []FavoriteResponse `json:"list" description:"收藏列表"`
}
