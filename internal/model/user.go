package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID           int64     `json:"id" db:"id"`
	Phone        string    `json:"phone" db:"phone"`
	Nickname     string    `json:"nickname" db:"nickname"`
	AvatarURL    string    `json:"avatar_url" db:"avatar_url"`
	Role         string    `json:"role" db:"role"` // admin, user
	WechatOpenID string    `json:"wechat_openid" db:"wechat_openid"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserResponse 用户响应（隐藏敏感信息）
type UserResponse struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse 转换为响应格式
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Phone:     u.Phone,
		Nickname:  u.Nickname,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Code     string `json:"code"`
	WechatID string `json:"wechat_id"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
