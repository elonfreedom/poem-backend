# 用户模块

## 注册/登录

| 登录方式 | 说明 |
|---------|------|
| 手机号+验证码 | 主要登录方式，发送短信验证码 |
| 微信登录 | 通过微信开放平台 OAuth |
| Apple 登录 | 通过 Sign in with Apple |

## 个人信息管理

- 查看个人信息
- 修改昵称、头像
- 修改手机号

## 角色权限

| 角色 | 权限 |
|-----|------|
| admin | 所有功能 + 后台管理 |
| user | 浏览、收藏、打卡 |

## 数据模型

```go
// User 用户模型
type User struct {
    ID           int64
    Phone        string
    Nickname     string
    AvatarURL    string
    Role         string    // admin, user
    WechatOpenID string
    AppleUserID  string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// LoginRequest 登录请求
type LoginRequest struct {
    Phone    string `json:"phone" validate:"required"`
    Code     string `json:"code"`
    WechatID string `json:"wechat_id"`
    AppleID  string `json:"apple_id"`
}

// LoginResponse 登录响应
type LoginResponse struct {
    Token string       `json:"token"`
    User  UserResponse `json:"user"`
}
```

## API 接口

### 公开接口（无需认证）

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/public/login/phone | 手机号验证码登录 |
| POST | /api/public/login/wechat | 微信登录 |
| POST | /api/public/login/apple | Apple 登录 |
| POST | /api/public/sms/send | 发送验证码 |

### 用户接口（需认证）

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/user/profile | 获取个人信息 |
| PUT | /api/user/profile | 更新个人信息 |
