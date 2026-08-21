package model

import (
	"time"
)

// RecoveryCode 恢复码模型
type RecoveryCode struct {
	ID        int64      `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"` // UUID v7
	CodeHash  string     `json:"-" db:"code_hash"`      // bcrypt 哈希
	Used      bool       `json:"used" db:"used"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
}

// RecoveryCodeResponse 恢复码响应
type RecoveryCodeResponse struct {
	RecoveryCode string    `json:"recovery_code"`
	ExpireAt     time.Time `json:"expire_at"`
}

// RecoveryRequest 账号恢复请求
type RecoveryRequest struct {
	RecoveryCode string `json:"recovery_code" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
}
