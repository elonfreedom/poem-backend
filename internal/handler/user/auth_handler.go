package user

import (
	"github.com/go-fuego/fuego"

	"poem-backend/internal/model"
	userservice "poem-backend/internal/service/user"
)

type AuthHandler struct {
	authService *userservice.AuthService
}

func NewAuthHandler(authService *userservice.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// BeginRegistrationRequest 开始注册请求
type BeginRegistrationRequest struct {
	DeviceName string `json:"device_name" description:"设备名称（可选）"`
}

// BeginRegistrationResponse 开始注册响应
type BeginRegistrationResponse struct {
	Options any    `json:"options" description:"WebAuthn 公钥凭证选项"`
	Session any    `json:"session" description:"会话数据（需返回给服务端）"`
	UserID  string `json:"user_id" description:"临时用户 ID"`
}

// BeginRegistration 开始注册
func (h *AuthHandler) BeginRegistration(c fuego.ContextWithBody[BeginRegistrationRequest]) (*BeginRegistrationResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, err
	}

	options, session, userID, err := h.authService.BeginRegistration(c.Context(), body.DeviceName)
	if err != nil {
		return nil, err
	}

	return &BeginRegistrationResponse{
		Options: options,
		Session: session,
		UserID:  userID,
	}, nil
}

// FinishRegistrationRequest 完成注册请求
type FinishRegistrationRequest struct {
	SessionID string `json:"session_id" description:"会话 ID"`
	// 实际应用中需要完整的 WebAuthn 响应
	Credential any `json:"credential" description:"WebAuthn 凭证"`
}

// FinishRegistration 完成注册
func (h *AuthHandler) FinishRegistration(c fuego.ContextWithBody[FinishRegistrationRequest]) (*model.LoginResponse, error) {
	// 实际应用中需要完整解析 WebAuthn 响应
	return &model.LoginResponse{
		Token: "jwt-token-placeholder",
		User: model.UserResponse{
			ID:       "user-id",
			Nickname: "诗友",
		},
	}, nil
}

// BeginLoginResponse 开始登录响应
type BeginLoginResponse struct {
	Options any `json:"options" description:"WebAuthn 公钥凭证选项"`
	Session any `json:"session" description:"会话数据（需返回给服务端）"`
}

// BeginLogin 开始登录（无需请求体）
func (h *AuthHandler) BeginLogin(c fuego.ContextNoBody) (*BeginLoginResponse, error) {
	options, session, err := h.authService.BeginLogin(c.Context())
	if err != nil {
		return nil, err
	}

	return &BeginLoginResponse{
		Options: options,
		Session: session,
	}, nil
}

// FinishLoginRequest 完成登录请求
type FinishLoginRequest struct {
	SessionID  string `json:"session_id" description:"会话 ID"`
	Credential any    `json:"credential" description:"WebAuthn 凭证"`
}

// FinishLogin 完成登录
func (h *AuthHandler) FinishLogin(c fuego.ContextWithBody[FinishLoginRequest]) (*model.LoginResponse, error) {
	// 实际应用中需要完整解析 WebAuthn 响应
	return &model.LoginResponse{
		Token: "jwt-token-placeholder",
		User: model.UserResponse{
			ID:       "user-id",
			Nickname: "诗友",
		},
	}, nil
}
