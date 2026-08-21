package admin

import (
	"time"
)

// AdminLoginRequest 后台登录请求（适配 vben-admin，username 即邮箱）
type AdminLoginRequest struct {
	Username      string `json:"username" validate:"required" description:"管理员邮箱/用户名"`
	Password      string `json:"password" validate:"required,min=6,max=64" description:"密码（6-64个字符）"`
	SelectAccount string `json:"selectAccount,omitempty" description:"vben-admin 选择账户字段（忽略）"`
	Captcha       bool   `json:"captcha,omitempty" description:"vben-admin 验证码字段（忽略）"`
}

// AdminLoginResponse 后台登录响应（适配 vben-admin）
type AdminLoginResponse struct {
	AccessToken string         `json:"accessToken" description:"JWT 认证令牌"`
	User        AdminUserResponse `json:"user" description:"用户信息"`
}

// AdminUserResponse 后台用户响应
type AdminUserResponse struct {
	ID        string    `json:"id" description:"用户唯一标识"`
	Nickname  string    `json:"nickname" description:"用户昵称"`
	Email     string    `json:"email,omitempty" description:"邮箱地址（脱敏显示）"`
	Role      string    `json:"role" description:"用户角色"`
	CreatedAt time.Time `json:"created_at" description:"创建时间"`
}

// AdminUserInfoResponse 用户信息响应（适配 vben-admin /user/info）
type AdminUserInfoResponse struct {
	UserId   string      `json:"userId" description:"用户ID"`
	Username string      `json:"username" description:"用户名"`
	RealName string      `json:"realName" description:"真实姓名"`
	Avatar   string      `json:"avatar" description:"头像"`
	Desc     string      `json:"desc" description:"描述/角色"`
	HomePath string      `json:"homePath" description:"首页路径"`
	Roles    []AdminRoleInfo `json:"roles" description:"角色列表"`
}

type AdminRoleInfo struct {
	RoleName string `json:"roleName" description:"角色名称"`
	Value    string `json:"value" description:"角色值"`
}
