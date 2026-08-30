package admin

import (
	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type AuthHandler struct {
	adminAuthService *admin.AdminAuthService
}

func NewAuthHandler(adminAuthService *admin.AdminAuthService) *AuthHandler {
	return &AuthHandler{adminAuthService: adminAuthService}
}

// Login 后台管理员登录（适配 vben-admin）
func (h *AuthHandler) Login(c fuego.ContextWithBody[adminmodel.AdminLoginRequest]) (*response.APIResponse[adminmodel.AdminLoginResponse], error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.adminAuthService.Login(c.Context(), body.Username, body.Password)
	if err != nil {
		return nil, fuego.UnauthorizedError{Title: "login failed", Detail: err.Error()}
	}

	return response.OK(*result), nil
}

// GetUserInfo 获取管理员用户信息（适配 vben-admin /user/info）
func (h *AuthHandler) GetUserInfo(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminUserInfoResponse], error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录或登录已过期"}
	}

	result, err := h.adminAuthService.GetUserInfo(c.Context(), userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.OK(*result), nil
}

// GetAccessCodes 获取权限码（适配 vben-admin /auth/codes）
func (h *AuthHandler) GetAccessCodes(c fuego.ContextNoBody) (*response.APIResponse[[]string], error) {
	return response.OK([]string{}), nil
}

// Logout 退出登录（适配 vben-admin /auth/logout）
func (h *AuthHandler) Logout(c fuego.ContextNoBody) (*response.APIResponse[response.SimpleResponse], error) {
	return response.OK(response.SimpleResponse{Success: true}), nil
}
