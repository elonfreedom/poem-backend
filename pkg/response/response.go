package response

import (
	"net/http"

	"github.com/go-fuego/fuego"
)

// APIResponse 统一响应格式
type APIResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

// Success 成功响应
func Success[T any](c *fuego.Context, data T) error {
	return c.JSON(http.StatusOK, APIResponse[T]{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *fuego.Context, code int, message string) error {
	return c.JSON(code, APIResponse[any]{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400 错误
func BadRequest(c *fuego.Context, message string) error {
	return Error(c, http.StatusBadRequest, message)
}

// Unauthorized 401 错误
func Unauthorized(c *fuego.Context, message string) error {
	return Error(c, http.StatusUnauthorized, message)
}

// Forbidden 403 错误
func Forbidden(c *fuego.Context, message string) error {
	return Error(c, http.StatusForbidden, message)
}

// NotFound 404 错误
func NotFound(c *fuego.Context, message string) error {
	return Error(c, http.StatusNotFound, message)
}

// InternalError 500 错误
func InternalError(c *fuego.Context, message string) error {
	return Error(c, http.StatusInternalServerError, message)
}
