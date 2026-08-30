package user

import (
	"context"
	"strconv"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	"poem-backend/pkg/response"
)

// RequestContext 定义 handler 所需的通用 context 接口
// 兼容 fuego.ContextNoBody 和 fuego.ContextWithBody[T]
type RequestContext interface {
	Context() context.Context
	PathParam(name string) string
	QueryParam(name string) string
}

// RequireUserID 从 context 获取用户 ID，未登录返回 UnauthorizedError
func RequireUserID(c RequestContext) (string, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return "", fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}
	return userID, nil
}

// ParsePathID 解析路径参数 ID，无效返回 BadRequestError
func ParsePathID(c RequestContext, param string) (int64, error) {
	id, err := strconv.ParseInt(c.PathParam(param), 10, 64)
	if err != nil {
		return 0, fuego.BadRequestError{Title: "invalid id", Detail: "无效的 ID 参数"}
	}
	return id, nil
}

// ParsePathInt 解析路径参数为 int，无效返回 BadRequestError
func ParsePathInt(c RequestContext, param string) (int, error) {
	id, err := strconv.Atoi(c.PathParam(param))
	if err != nil {
		return 0, fuego.BadRequestError{Title: "invalid id", Detail: "无效的 ID 参数"}
	}
	return id, nil
}

// ParsePageParams 解析分页参数
func ParsePageParams(c RequestContext) (page, pageSize int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return page, pageSize
}

// StatusResponse 操作状态响应
type StatusResponse struct {
	Status  string `json:"status" description:"操作状态"`
	Message string `json:"message,omitempty" description:"提示信息"`
}

// Common status responses
var (
	StatusUpdated      = StatusResponse{Status: "updated"}
	StatusPublished    = StatusResponse{Status: "published"}
	StatusUnpublished  = StatusResponse{Status: "unpublished"}
	StatusDeleted      = StatusResponse{Status: "deleted"}
	StatusPaused       = StatusResponse{Status: "paused"}
	StatusResumed      = StatusResponse{Status: "resumed"}
	StatusFavorited    = StatusResponse{Status: "favorited"}
	StatusUnfavorited  = StatusResponse{Status: "unfavorited"}
	StatusUnsubscribed = StatusResponse{Status: "unsubscribed"}
)

// PageResponse 分页响应
type PageResponse[T any] struct {
	Items []T   `json:"items" description:"列表数据"`
	Total int64 `json:"total" description:"总条数"`
}

// OK 成功响应（统一使用 code: 0 格式）
func OK[T any](data T) *response.APIResponse[T] {
	return response.OK(data)
}

// Success 成功响应（code: 0）
// 用于用户端接口统一返回格式
func Success(data any) *response.APIResponse[any] {
	return response.Success(data)
}
