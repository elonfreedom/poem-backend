package user

import (
	"github.com/go-fuego/fuego"

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
func (h *UserHandler) GetProfile(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	profile, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(profile), nil
}

// UpdateProfile 更新个人信息
func (h *UserHandler) UpdateProfile(c fuego.ContextWithBody[usermodel.UpdateProfileRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	profile, err := h.userService.UpdateProfile(c.Context(), userID, &body)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(profile), nil
}

// GetPasskeys 获取 Passkey 列表
func (h *UserHandler) GetPasskeys(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	passkeys, err := h.userService.GetPasskeys(c.Context(), userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(passkeys), nil
}

// DeletePasskey 删除 Passkey
func (h *UserHandler) DeletePasskey(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	passkeyID, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	if err := h.userService.DeletePasskey(c.Context(), userID, passkeyID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusDeleted), nil
}
