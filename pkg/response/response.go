package response

import (
	"encoding/json"
	"net/http"

	"github.com/go-fuego/fuego"
)

// APIResponse 统一响应格式
type APIResponse[T any] struct {
	Code    int    `json:"code" description:"状态码，0 表示成功"`
	Message string `json:"message" description:"提示信息"`
	Data    T      `json:"data,omitempty" description:"响应数据"`
}

// PageData 分页数据（适配 vben-admin）
type PageData[T any] struct {
	Items []T   `json:"items" description:"列表数据"`
	Total int64 `json:"total" description:"总条数"`
}

// OK 成功响应
func OK[T any](data T) *APIResponse[T] {
	return &APIResponse[T]{Code: 0, Message: "ok", Data: data}
}

// PageOK 分页成功响应
func PageOK[T any](items []T, total int64) *APIResponse[PageData[T]] {
	return &APIResponse[PageData[T]]{Code: 0, Message: "ok", Data: PageData[T]{Items: items, Total: total}}
}

// Err 错误响应
func Err(code int, message string) *APIResponse[any] {
	return &APIResponse[any]{Code: code, Message: message}
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, code int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}

// Success 成功响应 - 返回数据和错误
func Success[T any](c fuego.ContextNoBody, data T) (any, error) {
	return data, nil
}

// Error 错误响应
func Error(c fuego.ContextNoBody, code int, message string) error {
	return writeJSON(c.Response(), code, APIResponse[any]{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400 错误
func BadRequest(c fuego.ContextNoBody, message string) error {
	return Error(c, http.StatusBadRequest, message)
}

// Unauthorized 401 错误
func Unauthorized(c fuego.ContextNoBody, message string) error {
	return Error(c, http.StatusUnauthorized, message)
}

// Forbidden 403 错误
func Forbidden(c fuego.ContextNoBody, message string) error {
	return Error(c, http.StatusForbidden, message)
}

// NotFound 404 错误
func NotFound(c fuego.ContextNoBody, message string) error {
	return Error(c, http.StatusNotFound, message)
}

// InternalError 500 错误
func InternalError(c fuego.ContextNoBody, message string) error {
	return Error(c, http.StatusInternalServerError, message)
}
