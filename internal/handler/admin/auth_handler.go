package admin

import (
	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
)

// APIResponse 统一响应格式（适配 vben-admin）
type APIResponse struct {
	Code    int    `json:"code" description:"状态码，0 表示成功"`
	Message string `json:"message" description:"提示信息"`
	Data    any    `json:"data" description:"响应数据"`
}

type AuthHandler struct {
	adminAuthService *admin.AdminAuthService
}

func NewAuthHandler(adminAuthService *admin.AdminAuthService) *AuthHandler {
	return &AuthHandler{adminAuthService: adminAuthService}
}

// Login 后台管理员登录（适配 vben-admin）
func (h *AuthHandler) Login(c fuego.ContextWithBody[adminmodel.AdminLoginRequest]) (*APIResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	// vben-admin 使用 username 字段，实际为邮箱
	result, err := h.adminAuthService.Login(c.Context(), body.Username, body.Password)
	if err != nil {
		return nil, fuego.UnauthorizedError{Title: "login failed", Detail: err.Error()}
	}

	return &APIResponse{Code: 0, Message: "ok", Data: result}, nil
}

// GetUserInfo 获取管理员用户信息（适配 vben-admin /user/info）
func (h *AuthHandler) GetUserInfo(c fuego.ContextNoBody) (*APIResponse, error) {
	// 从 context 获取用户ID
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录或登录已过期"}
	}

	result, err := h.adminAuthService.GetUserInfo(c.Context(), userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "get user info failed", Detail: err.Error()}
	}

	return &APIResponse{Code: 0, Message: "ok", Data: result}, nil
}

// GetAccessCodes 获取权限码（适配 vben-admin /auth/codes）
func (h *AuthHandler) GetAccessCodes(c fuego.ContextNoBody) (*APIResponse, error) {
	// 后台暂不实现细粒度权限，返回空数组
	return &APIResponse{Code: 0, Message: "ok", Data: []string{}}, nil
}

// Logout 退出登录（适配 vben-admin /auth/logout）
func (h *AuthHandler) Logout(c fuego.ContextNoBody) (*APIResponse, error) {
	// 后台使用 JWT 无状态认证，客户端清除 token 即可
	return &APIResponse{Code: 0, Message: "ok"}, nil
}
