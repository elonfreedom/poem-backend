package response

import (
	"encoding/json"
	"net/http"

	"github.com/go-fuego/fuego"
)

// APIResponse 统一响应格式（适配 vben-admin）
type APIResponse[T any] struct {
	Code    int    `json:"code" description:"状态码，0 表示成功，非 0 表示失败"`
	Message string `json:"message" description:"提示信息"`
	Error   string `json:"error,omitempty" description:"错误描述（失败时）"`
	Data    T      `json:"data" description:"响应数据"`
}

// PageData 分页数据（适配 vben-admin）
type PageData[T any] struct {
	Items []T   `json:"items" description:"列表数据"`
	Total int64 `json:"total" description:"总条数"`
}

// SimpleResponse 简单响应（用于删除、更新等无数据返回的操作）
type SimpleResponse struct {
	Success bool `json:"success" description:"操作是否成功"`
}

// OK 成功响应（code: 0）
func OK[T any](data T) *APIResponse[T] {
	return &APIResponse[T]{Code: 0, Message: "ok", Data: data}
}

// Success 成功响应（code: 0, message: "ok"）
// 用于用户端接口统一返回格式，与 Admin 端保持一致
func Success(data any) *APIResponse[any] {
	return &APIResponse[any]{Code: 0, Message: "ok", Data: data}
}

// PageOK 分页成功响应
func PageOK[T any](items []T, total int64) *APIResponse[PageData[T]] {
	return &APIResponse[PageData[T]]{Code: 0, Message: "ok", Data: PageData[T]{Items: items, Total: total}}
}

// Err 错误响应（根据错误类型返回不同错误码）
func Err(code int, message string) *APIResponse[any] {
	if code == 0 {
		code = 400 // 默认 400
	}
	return &APIResponse[any]{Code: code, Message: message, Error: message}
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, code int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}

// SuccessHandler 成功响应 - 返回数据和错误（Fuego handler 用）
func SuccessHandler[T any](c fuego.ContextNoBody, data T) (any, error) {
	return data, nil
}

// Error 错误响应（HTTP 200 + body code）
func Error(c fuego.ContextNoBody, code int, message string) error {
	return writeJSON(c.Response(), http.StatusOK, APIResponse[any]{
		Code:    code,
		Message: message,
		Error:   message,
	})
}

// BadRequest 400 参数错误
func BadRequest(c fuego.ContextNoBody, message string) error {
	return Error(c, 400, message)
}

// Unauthorized 401 未登录/登录过期
func Unauthorized(c fuego.ContextNoBody, message string) error {
	return Error(c, 401, message)
}

// Forbidden 403 权限不足
func Forbidden(c fuego.ContextNoBody, message string) error {
	return Error(c, 403, message)
}

// NotFound 404 资源不存在
func NotFound(c fuego.ContextNoBody, message string) error {
	return Error(c, 404, message)
}

// InternalError 500 服务器内部错误
func InternalError(c fuego.ContextNoBody, message string) error {
	return Error(c, 500, message)
}
