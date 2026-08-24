package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/go-fuego/fuego"
	"github.com/go-webauthn/webauthn/webauthn"

	usermodel "poem-backend/internal/model/user"
	userservice "poem-backend/internal/service/user"
)

// SessionData 存储 WebAuthn 会话数据
type SessionData struct {
	Session webauthn.SessionData
	UserID  string // 仅注册流程使用
}

// SessionStore 线程安全的 WebAuthn 会话存储
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]SessionData
}

// NewSessionStore 创建新的会话存储
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]SessionData),
	}
}

// Store 存储会话
func (s *SessionStore) Store(id string, data SessionData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = data
}

// Get 获取会话
func (s *SessionStore) Get(id string) (SessionData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.sessions[id]
	return data, ok
}

// Delete 删除会话
func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

// AuthHandler 认证处理器
type AuthHandler struct {
	authService  *userservice.AuthService
	sessionStore *SessionStore
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *userservice.AuthService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		sessionStore: NewSessionStore(),
	}
}

// BeginRegistrationRequest 开始注册请求
type BeginRegistrationRequest struct {
	DeviceName string `json:"device_name" description:"设备名称（可选）"`
}

// BeginRegistrationResponse 开始注册响应
type BeginRegistrationResponse struct {
	Options   any    `json:"options" description:"WebAuthn 公钥凭证选项"`
	Session   any    `json:"session" description:"会话数据（需返回给服务端）"`
	UserID    string `json:"user_id" description:"临时用户 ID"`
	SessionID string `json:"session_id" description:"会话 ID（需返回给服务端）"`
}

// BeginRegistration 开始注册
func (h *AuthHandler) BeginRegistration(c fuego.ContextWithBody[BeginRegistrationRequest]) (*BeginRegistrationResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, err
	}

	options, session, userID, sessionID, err := h.authService.BeginRegistration(c.Context(), body.DeviceName)
	if err != nil {
		return nil, err
	}

	// 存储会话数据
	h.sessionStore.Store(sessionID, SessionData{
		Session: *session,
		UserID:  userID,
	})

	return &BeginRegistrationResponse{
		Options:   options,
		Session:   session,
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}

// FinishRegistrationRequest 完成注册请求
type FinishRegistrationRequest struct {
	SessionID  string `json:"session_id" description:"会话 ID"`
	Credential any    `json:"credential" description:"WebAuthn 凭证"`
}

// FinishRegistration 完成注册
func (h *AuthHandler) FinishRegistration(c fuego.ContextWithBody[FinishRegistrationRequest]) (*usermodel.LoginResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	// 获取会话数据
	sessionData, ok := h.sessionStore.Get(body.SessionID)
	if !ok {
		return nil, fuego.BadRequestError{Title: "invalid session", Detail: "会话不存在或已过期"}
	}

	// 删除会话（一次性使用）
	h.sessionStore.Delete(body.SessionID)

	// 将凭证序列化为 JSON
	credentialJSON, err := json.Marshal(body.Credential)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid credential", Detail: "凭证格式错误"}
	}

	// 构造包含凭证的请求
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "", bytes.NewReader(credentialJSON))
	if err != nil {
		return nil, fuego.InternalServerError{Title: "failed to create request", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	// 调用服务完成注册
	return h.authService.FinishRegistration(c.Context(), sessionData.UserID, sessionData.Session, req)
}

// BeginLoginResponse 开始登录响应
type BeginLoginResponse struct {
	Options   any    `json:"options" description:"WebAuthn 公钥凭证选项"`
	Session   any    `json:"session" description:"会话数据（需返回给服务端）"`
	SessionID string `json:"session_id" description:"会话 ID（需返回给服务端）"`
}

// BeginLogin 开始登录（无需请求体）
func (h *AuthHandler) BeginLogin(c fuego.ContextNoBody) (*BeginLoginResponse, error) {
	options, session, sessionID, err := h.authService.BeginLogin(c.Context())
	if err != nil {
		return nil, err
	}

	// 存储会话数据
	h.sessionStore.Store(sessionID, SessionData{
		Session: *session,
	})

	return &BeginLoginResponse{
		Options:   options,
		Session:   session,
		SessionID: sessionID,
	}, nil
}

// FinishLoginRequest 完成登录请求
type FinishLoginRequest struct {
	SessionID  string `json:"session_id" description:"会话 ID"`
	Credential any    `json:"credential" description:"WebAuthn 凭证"`
}

// FinishLogin 完成登录
func (h *AuthHandler) FinishLogin(c fuego.ContextWithBody[FinishLoginRequest]) (*usermodel.LoginResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	// 获取会话数据
	sessionData, ok := h.sessionStore.Get(body.SessionID)
	if !ok {
		return nil, fuego.BadRequestError{Title: "invalid session", Detail: "会话不存在或已过期"}
	}

	// 删除会话（一次性使用）
	h.sessionStore.Delete(body.SessionID)

	// 将凭证序列化为 JSON
	credentialJSON, err := json.Marshal(body.Credential)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid credential", Detail: "凭证格式错误"}
	}

	// 构造包含凭证的请求
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "", bytes.NewReader(credentialJSON))
	if err != nil {
		return nil, fuego.InternalServerError{Title: "failed to create request", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	// 调用服务完成登录
	return h.authService.FinishLogin(c.Context(), sessionData.Session, req)
}
