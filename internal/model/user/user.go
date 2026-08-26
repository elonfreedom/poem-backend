package usermodel

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// User 用户模型
type User struct {
	ID           string                `json:"id" db:"id" description:"用户唯一标识 (UUID v7)"`
	Nickname     string                `json:"nickname" db:"nickname" description:"用户昵称"`
	Email        *string               `json:"email,omitempty" db:"email" description:"邮箱地址（可选）"`
	Role         string                `json:"role" db:"role" description:"用户角色: admin, user"`
	Status       string                `json:"status" db:"status" description:"用户状态: active, disabled"`
	PasswordHash *string               `json:"-" db:"password_hash" description:"密码哈希（后台登录用）"`
	Credentials  []webauthn.Credential `json:"-" db:"-" description:"WebAuthn 凭证（不持久化）"`
	CreatedAt    time.Time             `json:"created_at" db:"created_at" description:"创建时间"`
	UpdatedAt    time.Time             `json:"updated_at" db:"updated_at" description:"更新时间"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        string    `json:"id" description:"用户唯一标识 (UUID v7)"`
	Nickname  string    `json:"nickname" description:"用户昵称"`
	Email     string    `json:"email,omitempty" description:"邮箱地址（脱敏显示）"`
	Role      string    `json:"role" description:"用户角色: admin, user"`
	Status    string    `json:"status" description:"用户状态: active, disabled"`
	CreatedAt time.Time `json:"created_at" description:"创建时间"`
}

// LoginResponse 登录响应（用户端 Passkey 登录）
type LoginResponse struct {
	Token string       `json:"token" description:"JWT 认证令牌"`
	User  UserResponse `json:"user" description:"用户信息"`
}

// UpdateProfileRequest 更新个人信息请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" validate:"omitempty,min=2,max=20" description:"用户昵称（2-20个字符）"`
}

// BindEmailRequest 绑定邮箱请求
type BindEmailRequest struct {
	Email string `json:"email" validate:"required,email" description:"邮箱地址"`
}

// ToResponse 转换为响应格式
func (u *User) ToResponse() UserResponse {
	resp := UserResponse{
		ID:        u.ID,
		Nickname:  u.Nickname,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
	if u.Email != nil {
		resp.Email = maskEmail(*u.Email)
	}
	return resp
}

// maskEmail 脱敏邮箱：abc***@example.com
func maskEmail(email string) string {
	at := -1
	for i, c := range email {
		if c == '@' {
			at = i
			break
		}
	}
	if at <= 3 {
		return "***" + email[at:]
	}
	return email[:3] + "***" + email[at:]
}

// WebAuthnUser 实现 go-webauthn 的 User 接口
func (u *User) WebAuthnID() []byte {
	return []byte(u.ID)
}

func (u *User) WebAuthnName() string {
	return u.Nickname
}

func (u *User) WebAuthnDisplayName() string {
	return u.Nickname
}

func (u *User) WebAuthnIcon() string {
	return ""
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
