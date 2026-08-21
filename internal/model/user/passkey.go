package usermodel

import "time"

// Passkey 通行密钥模型
type Passkey struct {
	ID           int64     `json:"id" db:"id" description:"Passkey ID"`
	UserID       string    `json:"user_id" db:"user_id" description:"用户ID (UUID v7)"`
	CredentialID []byte    `json:"-" db:"credential_id" description:"凭证ID（二进制）"`
	PublicKey    []byte    `json:"-" db:"public_key" description:"公钥"`
	SignCount    uint32    `json:"sign_count" db:"sign_count" description:"签名计数器"`
	DeviceName   string    `json:"device_name" db:"device_name" description:"设备名称"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" description:"创建时间"`
	LastUsedAt   time.Time `json:"last_used_at" db:"last_used_at" description:"最后使用时间"`
}

// PasskeyResponse Passkey 响应
type PasskeyResponse struct {
	ID         int64     `json:"id" description:"Passkey ID"`
	DeviceName string    `json:"device_name" description:"设备名称"`
	CreatedAt  time.Time `json:"created_at" description:"创建时间"`
	LastUsedAt time.Time `json:"last_used_at" description:"最后使用时间"`
}

// ToResponse 转换为响应格式
func (p *Passkey) ToResponse() PasskeyResponse {
	return PasskeyResponse{
		ID:         p.ID,
		DeviceName: p.DeviceName,
		CreatedAt:  p.CreatedAt,
		LastUsedAt: p.LastUsedAt,
	}
}
