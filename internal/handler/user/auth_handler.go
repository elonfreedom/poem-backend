package user

import (
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-webauthn/webauthn/webauthn"

	"poem-backend/internal/middleware"
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
	authService     *userservice.AuthService
	sessionStore    *SessionStore
	connectionStore *ConnectionStore
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *userservice.AuthService) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		sessionStore:    NewSessionStore(),
		connectionStore: NewConnectionStore(),
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
	SessionID string `json:"session_id" description:"会话 ID（通过 X-Session-ID header 回传）"`
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

// FinishRegistration 完成注册
// 请求体：标准 RegistrationResponseJSON（id, rawId, response: {clientDataJSON, attestationObject, ...}）
// 会话 ID：通过 X-Session-ID header 传递
func (h *AuthHandler) FinishRegistration(c fuego.ContextNoBody) (*usermodel.LoginResponse, error) {
	// 从 header 获取 session_id
	sessionID := c.Header("X-Session-ID")
	if sessionID == "" {
		return nil, fuego.BadRequestError{Title: "missing session", Detail: "X-Session-ID header is required"}
	}

	// 获取会话数据
	sessionData, ok := h.sessionStore.Get(sessionID)
	if !ok {
		return nil, fuego.BadRequestError{Title: "invalid session", Detail: "会话不存在或已过期"}
	}

	// 删除会话（一次性使用）
	h.sessionStore.Delete(sessionID)

	// 使用原始请求体直接传递给 WebAuthn 库
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "", io.NopCloser(c.Request().Body))
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
	SessionID string `json:"session_id" description:"会话 ID（通过 X-Session-ID header 回传）"`
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

// FinishLogin 完成登录
// 请求体：标准 AuthenticationResponseJSON（id, rawId, response: {clientDataJSON, authenticatorData, signature, ...}）
// 会话 ID：通过 X-Session-ID header 传递
func (h *AuthHandler) FinishLogin(c fuego.ContextNoBody) (*usermodel.LoginResponse, error) {
	// 从 header 获取 session_id
	sessionID := c.Header("X-Session-ID")
	if sessionID == "" {
		return nil, fuego.BadRequestError{Title: "missing session", Detail: "X-Session-ID header is required"}
	}

	// 获取会话数据
	sessionData, ok := h.sessionStore.Get(sessionID)
	if !ok {
		return nil, fuego.BadRequestError{Title: "invalid session", Detail: "会话不存在或已过期"}
	}

	// 删除会话（一次性使用）
	h.sessionStore.Delete(sessionID)

	// 使用原始请求体直接传递给 WebAuthn 库
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "", io.NopCloser(c.Request().Body))
	if err != nil {
		return nil, fuego.InternalServerError{Title: "failed to create request", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	// 调用服务完成登录
	return h.authService.FinishLogin(c.Context(), sessionData.Session, req)
}

// ==================== 跨设备 Passkey ====================

// AddDeviceBeginRequest 开始添加设备请求
type AddDeviceBeginRequest struct {
	DeviceName string `json:"device_name" description:"新设备名称（可选）"`
}

// AddDeviceBeginResponse 开始添加设备响应
type AddDeviceBeginResponse struct {
	ConnectionToken string `json:"connection_token" description:"连接令牌（5分钟有效）"`
	Options         any    `json:"options" description:"WebAuthn 注册选项（给新设备使用）"`
	ExpiresAt       string `json:"expires_at" description:"过期时间（RFC3339）"`
}

// AddDeviceBegin 开始添加新设备（设备 A 调用）
// 生成连接令牌和 WebAuthn 注册选项
func (h *AuthHandler) AddDeviceBegin(c fuego.ContextWithBody[AddDeviceBeginRequest]) (*AddDeviceBeginResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	// 调用服务生成连接令牌、WebAuthn 选项和会话
	token, options, session, expiresAt, err := h.authService.BeginAddDevice(c.Context(), userID, body.DeviceName)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "failed to begin add device", Detail: err.Error()}
	}

	// 存储连接状态、会话和注册选项
	h.connectionStore.Store(token, &Connection{
		Token:           token,
		UserID:          userID,
		Status:          ConnectionStatusWaiting,
		WebAuthnSession: session,  // webauthn.SessionData（finish 时验证 credential）
		WebAuthnOptions: options, // protocol.CredentialCreation（设备 B 创建 credential）
		CreatedAt:       time.Now(),
		ExpiresAt:       expiresAt,
	})

	return &AddDeviceBeginResponse{
		ConnectionToken: token,
		Options:         options,
		ExpiresAt:       expiresAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// AddDeviceStatusResponse 查询连接状态响应
type AddDeviceStatusResponse struct {
	Status      string `json:"status" description:"连接状态：waiting/connected/confirmed/rejected/expired"`
	DeviceName string `json:"device_name" description:"新设备名称（connected 时返回）"`
}

// AddDeviceStatus 查询连接状态（设备 A 轮询）
func (h *AuthHandler) AddDeviceStatus(c fuego.ContextNoBody) (*AddDeviceStatusResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	token := c.QueryParam("token")
	if token == "" {
		return nil, fuego.BadRequestError{Title: "missing token", Detail: "token 参数必填"}
	}

	conn, ok := h.connectionStore.Get(token)
	if !ok {
		return &AddDeviceStatusResponse{Status: string(ConnectionStatusExpired)}, nil
	}

	// 验证所有权
	if conn.UserID != userID {
		return nil, fuego.ForbiddenError{Title: "forbidden", Detail: "无权访问此连接"}
	}

	// 检查是否过期
	if time.Now().After(conn.ExpiresAt) {
		return &AddDeviceStatusResponse{Status: string(ConnectionStatusExpired)}, nil
	}

	resp := &AddDeviceStatusResponse{
		Status: string(conn.Status),
	}
	if conn.Status == ConnectionStatusConnected || conn.Status == ConnectionStatusConfirmed {
		resp.DeviceName = conn.DeviceName
	}
	return resp, nil
}

// AddDeviceStatusPublic 查询连接状态（公开接口，设备 B 轮询）
// 无需认证，仅通过 token 查询
// 返回统一格式：{code, message, data: {status, device_name}}
func (h *AuthHandler) AddDeviceStatusPublic(c fuego.ContextNoBody) (map[string]any, error) {
	token := c.QueryParam("token")
	if token == "" {
		return nil, fuego.BadRequestError{Title: "missing token", Detail: "token 参数必填"}
	}

	conn, ok := h.connectionStore.Get(token)
	status := ConnectionStatusExpired
	if ok {
		// 检查是否过期
		if !time.Now().After(conn.ExpiresAt) {
			status = conn.Status
		}
	}

	data := map[string]any{
		"status": string(status),
	}
	if status == ConnectionStatusConnected || status == ConnectionStatusConfirmed {
		data["device_name"] = conn.DeviceName
	}

	return map[string]any{
		"code":    200,
		"message": "success",
		"data":    data,
	}, nil
}

// AddDeviceConnectRequest 设备 B 连接请求
type AddDeviceConnectRequest struct {
	Token      string `json:"token" description:"连接令牌"`
	DeviceName string `json:"device_name" description:"新设备名称"`
}

// AddDeviceConnect 设备 B 连接（扫码后调用）
func (h *AuthHandler) AddDeviceConnect(c fuego.ContextWithBody[AddDeviceConnectRequest]) (map[string]any, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if body.Token == "" {
		return nil, fuego.BadRequestError{Title: "missing token", Detail: "token 必填"}
	}

	conn, ok := h.connectionStore.Get(body.Token)
	if !ok {
		return nil, fuego.BadRequestError{Title: "invalid token", Detail: "连接令牌无效或已过期"}
	}

	if time.Now().After(conn.ExpiresAt) {
		return nil, fuego.BadRequestError{Title: "expired", Detail: "连接已过期"}
	}

	if conn.Status != ConnectionStatusWaiting {
		return nil, fuego.BadRequestError{Title: "invalid status", Detail: "连接状态错误"}
	}

	// 更新连接状态为已连接
	h.connectionStore.Update(body.Token, func(c *Connection) {
		c.Status = ConnectionStatusConnected
		c.DeviceName = body.DeviceName
	})

	// 返回 WebAuthn 注册选项（设备 B 需要此数据创建 credential）
	return map[string]any{
		"code":    200,
		"message": "success",
		"data": map[string]any{
			"status":  string(ConnectionStatusConnected),
			"message": "已连接，等待设备 A 确认",
			"options": conn.WebAuthnOptions, // protocol.CredentialCreation（含 publicKey）
		},
	}, nil
}

// AddDeviceConfirmRequest 确认授权请求
type AddDeviceConfirmRequest struct {
	Token     string `json:"connection_token" description:"连接令牌"`
	Confirmed bool   `json:"confirmed" description:"是否确认授权"`
}

// AddDeviceConfirm 设备 A 确认/拒绝授权
func (h *AuthHandler) AddDeviceConfirm(c fuego.ContextWithBody[AddDeviceConfirmRequest]) (map[string]string, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	conn, ok := h.connectionStore.Get(body.Token)
	if !ok {
		return nil, fuego.BadRequestError{Title: "invalid token", Detail: "连接令牌无效或已过期"}
	}

	if conn.UserID != userID {
		return nil, fuego.ForbiddenError{Title: "forbidden", Detail: "无权操作此连接"}
	}

	if conn.Status != ConnectionStatusConnected {
		return nil, fuego.BadRequestError{Title: "invalid status", Detail: "连接状态错误，无法确认"}
	}

	if body.Confirmed {
		h.connectionStore.Update(body.Token, func(c *Connection) {
			c.Status = ConnectionStatusConfirmed
		})
		return map[string]string{"status": "confirmed", "message": "已确认授权"}, nil
	}

	// 拒绝
	h.connectionStore.Update(body.Token, func(c *Connection) {
		c.Status = ConnectionStatusRejected
	})
	return map[string]string{"status": "rejected", "message": "已拒绝"}, nil
}

// AddDeviceFinishRequest 新设备完成注册请求
type AddDeviceFinishRequest struct {
	Token      string `json:"connection_token" description:"连接令牌"`
	Credential any    `json:"credential" description:"WebAuthn PublicKeyCredential JSON"`
	DeviceName string `json:"device_name" description:"新设备名称"`
}

// AddDeviceFinish 新设备完成注册（设备 B 调用）
func (h *AuthHandler) AddDeviceFinish(c fuego.ContextWithBody[AddDeviceFinishRequest]) (*usermodel.LoginResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if body.Token == "" {
		return nil, fuego.BadRequestError{Title: "missing token", Detail: "connection_token 必填"}
	}

	conn, ok := h.connectionStore.Get(body.Token)
	if !ok {
		return nil, fuego.BadRequestError{Title: "invalid token", Detail: "连接令牌无效或已过期"}
	}

	if conn.Status != ConnectionStatusConfirmed {
		return nil, fuego.BadRequestError{Title: "not confirmed", Detail: "设备 A 尚未确认授权"}
	}

	if time.Now().After(conn.ExpiresAt) {
		return nil, fuego.BadRequestError{Title: "expired", Detail: "连接已过期"}
	}

	// 获取 WebAuthn 会话数据（指针类型）
	sessionPtr, ok := conn.WebAuthnSession.(*webauthn.SessionData)
	if !ok || sessionPtr == nil {
		return nil, fuego.InternalServerError{Title: "invalid session", Detail: "WebAuthn 会话数据无效"}
	}
	session := *sessionPtr

	// 使用原始请求体创建一个新的 http.Request（WebAuthn 库需要解析 body）
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "", io.NopCloser(c.Request().Body))
	if err != nil {
		return nil, fuego.InternalServerError{Title: "failed to create request", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	// 完成注册，验证 credential 并保存 passkey
	result, err := h.authService.FinishAddDevice(c.Context(), conn.UserID, session, req, body.DeviceName)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "failed to finish registration", Detail: err.Error()}
	}

	// 标记连接完成
	h.connectionStore.Update(body.Token, func(c *Connection) {
		c.Status = ConnectionStatusCompleted
	})

	return result, nil
}
