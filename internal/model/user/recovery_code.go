package usermodel

import "time"

// RecoveryCode 恢复码模型
type RecoveryCode struct {
	ID        int64      `json:"id" db:"id" description:"恢复码 ID"`
	UserID    string     `json:"user_id" db:"user_id" description:"用户 ID"`
	CodeHash  string     `json:"-" db:"code_hash" description:"恢复码哈希（不返回给前端）"`
	Used      bool       `json:"used" db:"used" description:"是否已使用"`
	CreatedAt time.Time  `json:"created_at" db:"created_at" description:"创建时间"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at" description:"使用时间"`
}

// RecoveryCodeResponse 恢复码响应
type RecoveryCodeResponse struct {
	RecoveryCode string    `json:"recovery_code" description:"恢复码"`
	ExpireAt     time.Time `json:"expire_at" description:"过期时间"`
}

// RecoveryRequest 账号恢复请求
type RecoveryRequest struct {
	RecoveryCode string `json:"recovery_code" validate:"required" description:"恢复码"`
	Email        string `json:"email" validate:"required,email" description:"邮箱"`
}
