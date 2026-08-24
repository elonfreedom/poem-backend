package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/adminmodel"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type UserHandler struct {
	userService *admin.AdminUserService
}

func NewUserHandler(userService *admin.AdminUserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List 获取前端用户列表
func (h *UserHandler) List(c fuego.ContextNoBody) (*response.APIResponse[response.PageData[adminmodel.AdminUserListItem]], error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	keyword := c.QueryParam("keyword")
	status := c.QueryParam("status")

	items, total, err := h.userService.ListUsers(c.Context(), page, pageSize, keyword, status)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "list failed", Detail: err.Error()}
	}

	return response.PageOK(items, total), nil
}

// GetByID 获取用户详情
func (h *UserHandler) GetByID(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.AdminUserDetailResponse], error) {
	id := c.QueryParam("id")
	if id == "" {
		// 也支持路径参数
		id = c.PathParam("id")
	}
	if id == "" {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "用户ID不能为空"}
	}

	detail, err := h.userService.GetUserDetail(c.Context(), id)
	if err != nil {
		return nil, fuego.NotFoundError{Title: "not found", Detail: err.Error()}
	}

	return response.OK(*detail), nil
}

// UpdateStatus 更新用户状态（禁用/启用）
func (h *UserHandler) UpdateStatus(c fuego.ContextWithBody[adminmodel.AdminUserUpdateStatusRequest]) (*response.APIResponse[any], error) {
	id := c.QueryParam("id")
	if id == "" {
		id = c.PathParam("id")
	}
	if id == "" {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "用户ID不能为空"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.userService.UpdateUserStatus(c.Context(), id, body.Status); err != nil {
		return nil, fuego.InternalServerError{Title: "update failed", Detail: err.Error()}
	}

	return response.OK[any](nil), nil
}
