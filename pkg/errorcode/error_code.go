package errorcode

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
)

// ErrorCode 业务错误码
type ErrorCode int

const (
	// 1xxx - 参数错误
	ErrParamRequired    ErrorCode = 1001 // 必填参数缺失
	ErrParamInvalid     ErrorCode = 1002 // 参数格式错误
	ErrParamOutOfRange  ErrorCode = 1003 // 参数超出范围
	ErrParamTooLong     ErrorCode = 1004 // 参数长度超限
	ErrBodyMalformed    ErrorCode = 1005 // 请求体格式错误
	ErrQueryRequired    ErrorCode = 1006 // 必填查询参数缺失

	// 2xxx - 认证授权错误
	ErrUnauthorized     ErrorCode = 2001 // 未认证
	ErrTokenExpired     ErrorCode = 2002 // Token 已过期
	ErrTokenInvalid     ErrorCode = 2003 // Token 格式错误
	ErrForbidden        ErrorCode = 2004 // 无权限
	ErrAdminRequired    ErrorCode = 2005 // 需要管理员权限

	// 3xxx - 资源不存在
	ErrUserNotFound        ErrorCode = 3001 // 用户不存在
	ErrPasskeyNotFound     ErrorCode = 3002 // Passkey 不存在
	ErrPoemNotFound        ErrorCode = 3003 // 诗歌不存在
	ErrPlanNotFound        ErrorCode = 3004 // 阅读计划不存在
	ErrConnectionNotFound  ErrorCode = 3005 // 连接不存在
	ErrCredentialNotFound  ErrorCode = 3006 // 凭证不存在

	// 4xxx - 业务冲突
	ErrConnectionExpired   ErrorCode = 4001 // 连接令牌已过期
	ErrConnectionInvalid   ErrorCode = 4002 // 连接令牌无效
	ErrConnectionStatus    ErrorCode = 4003 // 连接状态错误
	ErrNotConfirmed        ErrorCode = 4004 // 设备 A 尚未确认
	ErrDeviceNameExists    ErrorCode = 4005 // 设备名称已存在
	ErrPasskeyExists       ErrorCode = 4006 // Passkey 已存在

	// 5xxx - 验证失败
	ErrWebAuthnVerify      ErrorCode = 5001 // WebAuthn 验证失败
	ErrCredentialVerify    ErrorCode = 5002 // 凭证验证失败
	ErrSignatureInvalid    ErrorCode = 5003 // 签名无效

	// 9xxx - 服务器错误
	ErrInternal            ErrorCode = 9001 // 服务器内部错误
	ErrDatabase            ErrorCode = 9002 // 数据库错误
	ErrDatabaseQuery       ErrorCode = 9003 // 数据库查询失败
	ErrDatabaseInsert      ErrorCode = 9004 // 数据库写入失败
	ErrDatabaseUpdate      ErrorCode = 9005 // 数据库更新失败
	ErrDatabaseDelete      ErrorCode = 9006 // 数据库删除失败
	ErrExternalService     ErrorCode = 9007 // 外部服务错误
)

// Error 表示一个业务错误
type Error struct {
	Code    ErrorCode
	Title   string
	Detail  string
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.Code, e.Title, e.Detail)
}

// ToFuegoError 转换为 Fuego HTTPError
func (e *Error) ToFuegoError() error {
	switch {
	case e.Code >= 1000 && e.Code < 2000:
		return fuego.BadRequestError{Title: e.Title, Detail: e.Detail}
	case e.Code >= 2000 && e.Code < 3000:
		switch e.Code {
		case ErrForbidden, ErrAdminRequired:
			return fuego.ForbiddenError{Title: e.Title, Detail: e.Detail}
		default:
			return fuego.UnauthorizedError{Title: e.Title, Detail: e.Detail}
		}
	case e.Code >= 3000 && e.Code < 4000:
		return fuego.NotFoundError{Title: e.Title, Detail: e.Detail}
	case e.Code >= 4000 && e.Code < 5000:
		return fuego.BadRequestError{Title: e.Title, Detail: e.Detail}
	case e.Code >= 5000 && e.Code < 6000:
		// 422 Unprocessable Entity
		return fuego.HTTPError{Status: http.StatusUnprocessableEntity, Title: e.Title, Detail: e.Detail}
	default:
		return fuego.InternalServerError{Title: e.Title, Detail: e.Detail}
	}
}

// New 创建业务错误
func New(code ErrorCode, title, detail string) *Error {
	return &Error{Code: code, Title: title, Detail: detail}
}

// Newf 创建带格式化的业务错误
func Newf(code ErrorCode, title, format string, args ...any) *Error {
	return &Error{Code: code, Title: title, Detail: fmt.Sprintf(format, args...)}
}

// ========== 便捷构造函数 ==========

// ParamRequired 必填参数缺失
func ParamRequired(field string) *Error {
	return Newf(ErrParamRequired, "missing parameter", "%s 不能为空", field)
}

// ParamInvalid 参数格式错误
func ParamInvalid(field, reason string) *Error {
	return Newf(ErrParamInvalid, "invalid parameter", "%s 格式错误: %s", field, reason)
}

// ParamOutOfRange 参数超出范围
func ParamOutOfRange(field string, min, max int) *Error {
	return Newf(ErrParamOutOfRange, "parameter out of range", "%s 超出范围: 允许 %d-%d", field, min, max)
}

// BodyMalformed 请求体格式错误
func BodyMalformed(err error) *Error {
	return Newf(ErrBodyMalformed, "malformed body", "请求体 JSON 解析失败: %v", err)
}

// QueryRequired 必填查询参数缺失
func QueryRequired(field string) *Error {
	return Newf(ErrQueryRequired, "missing parameter", "查询参数 %s 不能为空", field)
}

// Unauthorized 未认证
func Unauthorized() *Error {
	return New(ErrUnauthorized, "unauthorized", "未登录，请先完成认证")
}

// Forbidden 无权限
func Forbidden(reason string) *Error {
	return Newf(ErrForbidden, "forbidden", "%s", reason)
}

// UserNotFound 用户不存在
func UserNotFound(id string) *Error {
	return Newf(ErrUserNotFound, "user not found", "用户不存在: id=%s", id)
}

// ConnectionNotFound 连接不存在
func ConnectionNotFound(token string) *Error {
	return Newf(ErrConnectionNotFound, "connection not found", "连接令牌无效或已过期: token=%s", token)
}

// ConnectionExpired 连接已过期
func ConnectionExpired(token string) *Error {
	return Newf(ErrConnectionExpired, "connection expired", "连接已过期: token=%s", token)
}

// ConnectionStatusInvalid 连接状态错误
func ConnectionStatusInvalid(expected, actual string) *Error {
	return Newf(ErrConnectionStatus, "invalid connection status", "连接状态错误: 期望 %s, 实际 %s", expected, actual)
}

// NotConfirmed 设备 A 尚未确认
func NotConfirmed(status string) *Error {
	return Newf(ErrNotConfirmed, "not confirmed", "设备 A 尚未确认授权: 当前状态=%s", status)
}

// DatabaseError 数据库错误
func DatabaseError(operation string, err error) *Error {
	return Newf(ErrDatabase, "database error", "%s 失败: %v", operation, err)
}

// DatabaseQueryFailed 数据库查询失败
func DatabaseQueryFailed(resource string, err error) *Error {
	return Newf(ErrDatabaseQuery, "database error", "查询 %s 失败: %v", resource, err)
}

// Internal 服务器内部错误
func Internal(operation string, err error) *Error {
	return Newf(ErrInternal, "internal error", "%s 失败: %v", operation, err)
}

// WebAuthnVerifyFailed WebAuthn 验证失败
func WebAuthnVerifyFailed(err error) *Error {
	return Newf(ErrWebAuthnVerify, "webauthn verify failed", "WebAuthn 验证失败: %v", err)
}
