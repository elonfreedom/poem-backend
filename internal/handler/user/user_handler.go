package user

import (
	"strconv"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	usermodel "poem-backend/internal/model/user"
	userservice "poem-backend/internal/service/user"
	"poem-backend/pkg/response"
)

type UserHandler struct {
	userService *userservice.UserService
}

func NewUserHandler(userService *userservice.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile 获取个人信息
func (h *UserHandler) GetProfile(c fuego.ContextNoBody) (map[string]any, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	profile, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return response.Success(profile), nil
}

// UpdateProfile 更新个人信息
func (h *UserHandler) UpdateProfile(c fuego.ContextWithBody[usermodel.UpdateProfileRequest]) (map[string]any, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	profile, err := h.userService.UpdateProfile(c.Context(), userID, &body)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return response.Success(profile), nil
}

// GetPasskeys 获取 Passkey 列表
func (h *UserHandler) GetPasskeys(c fuego.ContextNoBody) (map[string]any, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	passkeys, err := h.userService.GetPasskeys(c.Context(), userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return response.Success(passkeys), nil
}

// DeletePasskey 删除 Passkey
func (h *UserHandler) DeletePasskey(c fuego.ContextNoBody) (map[string]any, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	passkeyID, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "无效的 Passkey ID"}
	}

	if err := h.userService.DeletePasskey(c.Context(), userID, passkeyID); err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return response.Success(map[string]string{"status": "deleted"}), nil
}
