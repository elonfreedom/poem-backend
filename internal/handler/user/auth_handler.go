package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-webauthn/webauthn/webauthn"

	"poem-backend/internal/middleware"
	"poem-backend/internal/repository"
	userservice "poem-backend/internal/service/user"
	"poem-backend/pkg/errorcode"
	"poem-backend/pkg/response"
)

// SessionData 存储 WebAuthn 会话数据
type SessionData struct {
	Session webauthn.SessionData
	UserID  string // 仅注册流程使用
}

// AuthHandler 认证处理器
type AuthHandler struct {
	AuthService     *userservice.AuthService
	SessionStore    *SessionStore
	ConnectionStore *ConnectionStore
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(
	authService *userservice.AuthService,
	sessionRepo *repository.SessionRepository,
	connectionRepo *repository.ConnectionRepository,
) *AuthHandler {
	store := NewSessionStore(sessionRepo)
	connStore := NewConnectionStore(connectionRepo)
	return &AuthHandler{
		AuthService:     authService,
		SessionStore:    store,
		ConnectionStore: connStore,
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
func (h *AuthHandler) BeginRegistration(c fuego.ContextWithBody[BeginRegistrationRequest]) (*response.APIResponse[any], error) {
	body, err := c.Body()
	if err != nil {
		return nil, errorcode.BodyMalformed(err).ToFuegoError()
	}

	options, session, userID, sessionID, err := h.AuthService.BeginRegistration(c.Context(), body.DeviceName)
	if err != nil {
		return nil, errorcode.Internal("初始化注册", err).ToFuegoError()
	}

	// 存储会话数据
	h.SessionStore.Store(sessionID, SessionData{
		Session: *session,
		UserID:  userID,
	})

	return response.Success(&BeginRegistrationResponse{
		Options:   options,
		Session:   session,
		UserID:    userID,
		SessionID: sessionID,
	}), nil
}

// FinishRegistration 完成注册
// 请求体：标准 RegistrationResponseJSON（id, rawId, response: {clientDataJSON, attestationObject, ...}）
// 会话 ID：通过 X-Session-ID header 传递
func (h *AuthHandler) FinishRegistration(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	// 从 header 获取 session_id
	sessionID := c.Header("X-Session-ID")
	if sessionID == "" {
		return nil, errorcode.ParamRequired("X-Session-ID header").ToFuegoError()
	}

	// 获取会话数据
	sessionData, ok := h.SessionStore.Get(sessionID)
	if !ok {
		return nil, errorcode.Newf(errorcode.ErrParamInvalid, "invalid session", "会话不存在或已过期: session_id=%s", sessionID).ToFuegoError()
	}

	// 删除会话（一次性使用）
	h.SessionStore.Delete(sessionID)

	// 使用原始请求体直接传递给 WebAuthn 库
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "", io.NopCloser(c.Request().Body))
	if err != nil {
		return nil, errorcode.Internal("创建请求", err).ToFuegoError()
	}
	req.Header.Set("Content-Type", "application/json")

	// 调用服务完成注册
	result, err := h.AuthService.FinishRegistration(c.Context(), sessionData.UserID, sessionData.Session, req)
	if err != nil {
		return nil, err
	}
	return response.Success(result), nil
}

// BeginLoginResponse 开始登录响应
type BeginLoginResponse struct {
	Options   any    `json:"options" description:"WebAuthn 公钥凭证选项"`
	Session   any    `json:"session" description:"会话数据（需返回给服务端）"`
	SessionID string `json:"session_id" description:"会话 ID（通过 X-Session-ID header 回传）"`
}

// BeginLogin 开始登录（无需请求体）
func (h *AuthHandler) BeginLogin(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	options, session, sessionID, err := h.AuthService.BeginLogin(c.Context())
	if err != nil {
		return nil, err
	}

	// 存储会话数据
	h.SessionStore.Store(sessionID, SessionData{
		Session: *session,
	})

	return response.Success(&BeginLoginResponse{
		Options:   options,
		Session:   session,
		SessionID: sessionID,
	}), nil
}

// authResponseCredential 用于从请求体中提取凭证 ID（仅用于错误信息）
type authResponseCredential struct {
	ID string `json:"id"`
}

// FinishLogin 完成登录
// 请求体：标准 AuthenticationResponseJSON（id, rawId, response: {clientDataJSON, authenticatorData, signature, ...}）
// 会话 ID：通过 X-Session-ID header 传递
func (h *AuthHandler) FinishLogin(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	// 从 header 获取 session_id
	sessionID := c.Header("X-Session-ID")
	if sessionID == "" {
		return nil, errorcode.ParamRequired("X-Session-ID header").ToFuegoError()
	}

	// 获取会话数据
	sessionData, ok := h.SessionStore.Get(sessionID)
	if !ok {
		return nil, errorcode.Newf(errorcode.ErrParamInvalid, "invalid session", "会话不存在或已过期: session_id=%s", sessionID).ToFuegoError()
	}

	// 删除会话（一次性使用）
	h.SessionStore.Delete(sessionID)

	// 读取原始请求体
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, errorcode.BodyMalformed(err).ToFuegoError()
	}
	c.Request().Body.Close()

	// 使用读取到的 body 创建新的请求
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "", io.NopCloser(bytes.NewReader(bodyBytes)))
	if err != nil {
		return nil, errorcode.Internal("创建请求", err).ToFuegoError()
	}
	req.Header.Set("Content-Type", "application/json")

	// 调用服务完成登录
	result, err := h.AuthService.FinishLogin(c.Context(), sessionData.Session, req)
	if err != nil {
		// 提取凭证 ID，附加到错误信息中方便调试
		var cred authResponseCredential
		if json.Unmarshal(bodyBytes, &cred) == nil && cred.ID != "" {
			return nil, errorcode.Newf(errorcode.ErrCredentialNotFound, "登录失败", "凭证验证失败: credential_id=%s, error=%s", cred.ID, err.Error()).ToFuegoError()
		}
		return nil, errorcode.Newf(errorcode.ErrCredentialNotFound, "登录失败", "凭证验证失败: %s", err.Error()).ToFuegoError()
	}
	return response.Success(result), nil
}

// getWebAuthnSession 从连接数据中提取 WebAuthn 会话
// 兼容两种情况：
//   - 内存缓存中的 *webauthn.SessionData（指针类型，直接返回）
//   - 数据库反序列化后的 map[string]interface{}（需重新反序列化）
func getWebAuthnSession(data any) (webauthn.SessionData, error) {
	switch v := data.(type) {
	case *webauthn.SessionData:
		if v == nil {
			return webauthn.SessionData{}, fmt.Errorf("session data is nil")
		}
		return *v, nil
	case webauthn.SessionData:
		return v, nil
	default:
		// 数据库反序列化后变成 map[string]interface{}，需要重新序列化/反序列化
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return webauthn.SessionData{}, fmt.Errorf("marshal session data failed: %w", err)
		}
		var session webauthn.SessionData
		if err := json.Unmarshal(jsonBytes, &session); err != nil {
			return webauthn.SessionData{}, fmt.Errorf("unmarshal session data failed: %w", err)
		}
		return session, nil
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
func (h *AuthHandler) AddDeviceBegin(c fuego.ContextWithBody[AddDeviceBeginRequest]) (*response.APIResponse[any], error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, errorcode.Unauthorized().ToFuegoError()
	}

	body, err := c.Body()
	if err != nil {
		return nil, errorcode.BodyMalformed(err).ToFuegoError()
	}

	// 调用服务生成连接令牌、WebAuthn 选项和会话
	token, options, session, expiresAt, err := h.AuthService.BeginAddDevice(c.Context(), userID, body.DeviceName)
	if err != nil {
		return nil, errorcode.Internal("生成连接令牌", err).ToFuegoError()
	}

	// 存储连接状态、会话和注册选项
	h.ConnectionStore.Store(token, &Connection{
		Token:           token,
		UserID:          userID,
		Status:          ConnectionStatusWaiting,
		WebAuthnSession: session, // webauthn.SessionData（finish 时验证 credential）
		WebAuthnOptions: options, // protocol.CredentialCreation（设备 B 创建 credential）
		CreatedAt:       time.Now(),
		ExpiresAt:       expiresAt,
	})

	return response.Success(&AddDeviceBeginResponse{
		ConnectionToken: token,
		Options:         options,
		ExpiresAt:       expiresAt.Format("2006-01-02T15:04:05Z"),
	}), nil
}

// AddDeviceStatusResponse 查询连接状态响应
type AddDeviceStatusResponse struct {
	Status     string `json:"status" description:"连接状态：waiting/connected/confirmed/rejected/expired"`
	DeviceName string `json:"device_name" description:"新设备名称（connected 时返回）"`
}

// AddDeviceStatus 查询连接状态（设备 A 长轮询）
// 当状态立即返回时直接响应；否则持有连接直到状态变化或 30s 超时
func (h *AuthHandler) AddDeviceStatus(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, errorcode.Unauthorized().ToFuegoError()
	}

	token := c.QueryParam("token")
	if token == "" {
		return nil, errorcode.QueryRequired("token").ToFuegoError()
	}

	conn, ok := h.ConnectionStore.Get(token)
	if !ok {
		return response.Success(AddDeviceStatusResponse{Status: string(ConnectionStatusExpired)}), nil
	}

	// 验证所有权
	if conn.UserID != userID {
		return nil, errorcode.Forbidden(fmt.Sprintf("无权访问此连接: token=%s", token)).ToFuegoError()
	}

	// 检查是否过期
	if time.Now().After(conn.ExpiresAt) {
		return response.Success(AddDeviceStatusResponse{Status: string(ConnectionStatusExpired)}), nil
	}

	// 长轮询：等待状态变化或超时
	statusCh, cleanup := h.ConnectionStore.Subscribe(token, connectionLongPollTimeout)
	defer cleanup()

	newStatus := <-statusCh
	if newStatus == ConnectionStatusTimeout {
		// 超时：返回当前状态
		conn, ok = h.ConnectionStore.Get(token)
		if !ok {
			return response.Success(AddDeviceStatusResponse{Status: string(ConnectionStatusExpired)}), nil
		}
		resp := AddDeviceStatusResponse{Status: string(conn.Status)}
		if conn.Status == ConnectionStatusConnected || conn.Status == ConnectionStatusConfirmed {
			resp.DeviceName = conn.DeviceName
		}
		return response.Success(resp), nil
	}

	// 状态变化：返回新状态
	resp := AddDeviceStatusResponse{Status: string(newStatus)}
	if newStatus == ConnectionStatusConnected || newStatus == ConnectionStatusConfirmed {
		if conn, ok := h.ConnectionStore.Get(token); ok {
			resp.DeviceName = conn.DeviceName
		}
	}
	return response.Success(resp), nil
}

// AddDeviceStatusPublic 查询连接状态（公开接口，设备 B 长轮询）
// 无需认证，仅通过 token 查询
// 当状态立即返回时直接响应；否则持有连接直到状态变化或 30s 超时
func (h *AuthHandler) AddDeviceStatusPublic(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	token := c.QueryParam("token")
	if token == "" {
		return nil, errorcode.QueryRequired("token").ToFuegoError()
	}

	conn, ok := h.ConnectionStore.Get(token)
	if !ok || time.Now().After(conn.ExpiresAt) {
		return response.Success(map[string]any{"status": string(ConnectionStatusExpired)}), nil
	}

	// 设备 B 首次访问时更新心跳
	status := conn.Status
	if status == ConnectionStatusWaiting || status == ConnectionStatusConnected {
		h.ConnectionStore.UpdateHeartbeat(token)
	}

	// 长轮询：等待状态变化或超时
	statusCh, cleanup := h.ConnectionStore.Subscribe(token, connectionLongPollTimeout)
	defer cleanup()

	newStatus := <-statusCh
	if newStatus == ConnectionStatusTimeout {
		// 超时：返回当前状态
		conn, ok := h.ConnectionStore.Get(token)
		if !ok || time.Now().After(conn.ExpiresAt) {
			return response.Success(map[string]any{"status": string(ConnectionStatusExpired)}), nil
		}
		data := map[string]any{"status": string(conn.Status)}
		if conn.Status == ConnectionStatusConnected || conn.Status == ConnectionStatusConfirmed {
			data["device_name"] = conn.DeviceName
		}
		return response.Success(data), nil
	}

	// 状态变化：返回新状态
	data := map[string]any{"status": string(newStatus)}
	if newStatus == ConnectionStatusConnected || newStatus == ConnectionStatusConfirmed {
		if conn, ok := h.ConnectionStore.Get(token); ok {
			data["device_name"] = conn.DeviceName
		}
	}
	return response.Success(data), nil
}

// AddDeviceConnectRequest 设备 B 连接请求
type AddDeviceConnectRequest struct {
	Token      string `json:"token" description:"连接令牌"`
	DeviceName string `json:"device_name" description:"新设备名称"`
}

// AddDeviceConnect 设备 B 连接（扫码后调用）
func (h *AuthHandler) AddDeviceConnect(c fuego.ContextWithBody[AddDeviceConnectRequest]) (*response.APIResponse[any], error) {
	body, err := c.Body()
	if err != nil {
		return nil, errorcode.BodyMalformed(err).ToFuegoError()
	}

	if body.Token == "" {
		return nil, errorcode.ParamRequired("token").ToFuegoError()
	}

	conn, ok := h.ConnectionStore.Get(body.Token)
	if !ok {
		return nil, errorcode.ConnectionNotFound(body.Token).ToFuegoError()
	}

	if time.Now().After(conn.ExpiresAt) {
		return nil, errorcode.ConnectionExpired(body.Token).ToFuegoError()
	}

	if conn.Status != ConnectionStatusWaiting {
		return nil, errorcode.ConnectionStatusInvalid("waiting", string(conn.Status)).ToFuegoError()
	}

	// 更新连接状态为已连接，同时更新心跳
	h.ConnectionStore.Update(body.Token, func(c *Connection) {
		c.Status = ConnectionStatusConnected
		c.DeviceName = body.DeviceName
		c.LastActiveAt = time.Now()
	})

	// 返回 WebAuthn 注册选项（设备 B 需要此数据创建 credential）
	return response.Success(map[string]any{
		"status":  string(ConnectionStatusConnected),
		"message": "已连接，等待设备 A 确认",
		"options": conn.WebAuthnOptions, // protocol.CredentialCreation（含 publicKey）
	}), nil
}

// AddDeviceConfirmRequest 确认授权请求
type AddDeviceConfirmRequest struct {
	Token     string `json:"connection_token" description:"连接令牌"`
	Confirmed bool   `json:"confirmed" description:"是否确认授权"`
}

// AddDeviceConfirm 设备 A 确认/拒绝授权
func (h *AuthHandler) AddDeviceConfirm(c fuego.ContextWithBody[AddDeviceConfirmRequest]) (*response.APIResponse[any], error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, errorcode.Unauthorized().ToFuegoError()
	}

	body, err := c.Body()
	if err != nil {
		return nil, errorcode.BodyMalformed(err).ToFuegoError()
	}

	if body.Token == "" {
		return nil, errorcode.ParamRequired("connection_token").ToFuegoError()
	}

	conn, ok := h.ConnectionStore.Get(body.Token)
	if !ok {
		return nil, errorcode.ConnectionNotFound(body.Token).ToFuegoError()
	}

	if conn.UserID != userID {
		return nil, errorcode.Forbidden(fmt.Sprintf("无权操作此连接: token=%s", body.Token)).ToFuegoError()
	}

	if conn.Status != ConnectionStatusConnected {
		return nil, errorcode.ConnectionStatusInvalid("connected", string(conn.Status)).ToFuegoError()
	}

	if body.Confirmed {
		h.ConnectionStore.Update(body.Token, func(c *Connection) {
			c.Status = ConnectionStatusConfirmed
		})
		return response.Success(StatusResponse{Status: "confirmed", Message: "已确认授权"}), nil
	}

	// 拒绝
	h.ConnectionStore.Update(body.Token, func(c *Connection) {
		c.Status = ConnectionStatusRejected
	})
	return response.Success(StatusResponse{Status: "rejected", Message: "已拒绝"}), nil
}

// AddDeviceRejectRequest 设备 B 拒绝/放弃请求（公开接口）
type AddDeviceRejectRequest struct {
	Token string `json:"token" description:"连接令牌"`
}

// AddDeviceReject 设备 B 主动放弃创建 Passkey（公开接口）
// 无需认证，设备 B 点击"取消"或关闭页面时调用
func (h *AuthHandler) AddDeviceReject(c fuego.ContextWithBody[AddDeviceRejectRequest]) (*response.APIResponse[any], error) {
	body, err := c.Body()
	if err != nil {
		return nil, errorcode.BodyMalformed(err).ToFuegoError()
	}

	if body.Token == "" {
		return nil, errorcode.ParamRequired("token").ToFuegoError()
	}

	conn, ok := h.ConnectionStore.Get(body.Token)
	if !ok {
		return nil, errorcode.ConnectionNotFound(body.Token).ToFuegoError()
	}

	// waiting、connected、confirmed 状态都可以拒绝
	// confirmed 后设备 B 仍可能因 Passkey 冲突需要取消
	if conn.Status != ConnectionStatusWaiting && conn.Status != ConnectionStatusConnected && conn.Status != ConnectionStatusConfirmed {
		return nil, errorcode.ConnectionStatusInvalid("waiting/connected/confirmed", string(conn.Status)).ToFuegoError()
	}

	// 更新状态为 rejected
	h.ConnectionStore.Update(body.Token, func(c *Connection) {
		c.Status = ConnectionStatusRejected
	})

	return response.Success(StatusResponse{Status: "rejected", Message: "设备 B 已取消"}), nil
}

// AddDeviceFinishRequest 新设备完成注册请求
type AddDeviceFinishRequest struct {
	Token      string `json:"connection_token" description:"连接令牌"`
	Credential any    `json:"credential" description:"WebAuthn PublicKeyCredential JSON"`
	DeviceName string `json:"device_name" description:"新设备名称"`
}

// AddDeviceFinish 新设备完成注册（设备 B 调用）
func (h *AuthHandler) AddDeviceFinish(c fuego.ContextWithBody[AddDeviceFinishRequest]) (*response.APIResponse[any], error) {
	body, err := c.Body()
	if err != nil {
		return nil, errorcode.BodyMalformed(err).ToFuegoError()
	}

	if body.Token == "" {
		return nil, errorcode.ParamRequired("connection_token").ToFuegoError()
	}

	conn, ok := h.ConnectionStore.Get(body.Token)
	if !ok {
		return nil, errorcode.ConnectionNotFound(body.Token).ToFuegoError()
	}

	if conn.Status != ConnectionStatusConfirmed {
		return nil, errorcode.NotConfirmed(string(conn.Status)).ToFuegoError()
	}

	if time.Now().After(conn.ExpiresAt) {
		return nil, errorcode.ConnectionExpired(body.Token).ToFuegoError()
	}

	// 获取 WebAuthn 会话数据（兼容内存指针和数据库反序列化后的 map）
	session, err := getWebAuthnSession(conn.WebAuthnSession)
	if err != nil {
		return nil, errorcode.Internal("获取 WebAuthn 会话", err).ToFuegoError()
	}

	// 从解析后的 credential 字段重建 WebAuthn 库需要的请求体
	// 注意：c.Body() 已消费原始请求体，不能再用 c.Request().Body
	credentialJSON, err := json.Marshal(body.Credential)
	if err != nil {
		return nil, errorcode.Internal("序列化凭证数据", err).ToFuegoError()
	}
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "", bytes.NewReader(credentialJSON))
	if err != nil {
		return nil, errorcode.Internal("创建请求", err).ToFuegoError()
	}
	req.Header.Set("Content-Type", "application/json")

	// 完成注册，验证 credential 并保存 passkey
	result, err := h.AuthService.FinishAddDevice(c.Context(), conn.UserID, session, req, body.DeviceName)
	if err != nil {
		return nil, errorcode.Internal("Passkey 注册", err).ToFuegoError()
	}

	// 标记连接完成
	h.ConnectionStore.Update(body.Token, func(c *Connection) {
		c.Status = ConnectionStatusCompleted
	})

	return response.Success(result), nil
}
