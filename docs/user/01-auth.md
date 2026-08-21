# 用户系统

## 功能概述

用户系统采用 **Passkey（通行密钥）** 作为主要注册登录方式，结合 **邮箱** 作为账号恢复手段。用户无需手机号、无需密码，一键生物识别完成登录。

## 用户故事

- 作为新用户，我想通过指纹/人脸一键注册登录，无需输入任何信息
- 作为老用户，我想通过指纹/人脸快速登录，无需记住密码
- 作为用户，我想在换设备时通过邮箱找回我的账号
- 作为用户，我想随时修改昵称，个性化我的资料

## 功能详情

### 1. Passkey 注册登录（主要方式）

**注册流程**：
1. 用户点击"开始使用"
2. 客户端向服务端请求注册 challenge
3. 调用 WebAuthn API → 系统弹出 Face ID / 指纹验证
4. 验证通过，公钥发送到服务端
5. 服务端创建用户账号，绑定 Passkey 公钥
6. 返回 JWT Token，进入首页

**登录流程**：
1. 用户点击"登录"
2. 客户端向服务端请求认证 challenge
3. 调用 WebAuthn API → 系统弹出 Face ID / 指纹验证
4. 验证通过，签名发送到服务端
5. 服务端验证签名，返回 JWT Token
6. 进入首页

**业务规则**：
- 注册和登录对用户来说是同一个操作（无感知区分）
- 首次使用自动创建账号，后续使用直接登录
- 同一设备可绑定多个 Passkey（如 Face ID + Touch ID）
- 新用户自动生成随机昵称（如"诗友1234"）
- 用户 ID 使用 **UUID v7**（时间有序，全局唯一）
- JWT Token 有效期 72 小时
- 不支持 Passkey 的设备降级为游客模式（可选）

### 2. 邮箱恢复（辅助方式）

**绑定邮箱**：
- 用户可在个人中心绑定邮箱
- 邮箱仅用于账号恢复，不作为登录方式
- 绑定后不可更改（或需验证原邮箱）

**账号恢复流程**：
1. 用户在旧设备或个人中心获取"恢复码"
2. 在新设备输入恢复码 + 绑定邮箱
3. 系统发送验证邮件到绑定邮箱
4. 用户点击邮件链接确认
5. 新设备注册新 Passkey，账号数据迁移完成

**业务规则**：
- 恢复码在绑定邮箱时生成，用户需妥善保存
- 未绑定邮箱的账号换设备后数据无法恢复
- 恢复码一次性使用，使用后重新生成

### 3. 个人信息管理

| 功能 | 说明 |
|-----|------|
| 查看个人信息 | 昵称、绑定邮箱（脱敏）、Passkey 数量 |
| 修改昵称 | 2-20 个字符（可选，不修改保持默认） |
| 绑定邮箱 | 用于账号恢复 |
| 管理 Passkey | 查看/删除已绑定的 Passkey |
| 查看恢复码 | 绑定邮箱后可查看 |

## 数据模型

```go
// User 用户模型
type User struct {
    ID        string    // UUID v7，时间有序
    Nickname  string    // 昵称（默认自动生成）
    Email     string    // 邮箱（可选，用于恢复）
    Role      string    // admin, user
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Passkey 通行密钥模型
type Passkey struct {
    ID           int64
    UserID       string    // UUID v7
    CredentialID string    // 凭证 ID（唯一）
    PublicKey    []byte    // 公钥
    SignCount    uint32    // 签名计数器（防重放）
    DeviceName   string    // 设备名称（如"iPhone 15"）
    CreatedAt    time.Time
    LastUsedAt   time.Time
}

// RecoveryCode 恢复码模型
type RecoveryCode struct {
    ID        int64
    UserID    string    // UUID v7
    Code      string    // 恢复码（加密存储）
    Used      bool      // 是否已使用
    CreatedAt time.Time
    UsedAt    *time.Time
}

// UserResponse 用户响应
type UserResponse struct {
    ID           string `json:"id"`
    Nickname     string `json:"nickname"`
    Email        string `json:"email,omitempty"` // 脱敏显示
    Role         string `json:"role"`
    PasskeyCount int    `json:"passkey_count"`
    HasEmail     bool   `json:"has_email"`
}

// PasskeyResponse Passkey 响应
type PasskeyResponse struct {
    ID         int64     `json:"id"`
    DeviceName string    `json:"device_name"`
    CreatedAt  time.Time `json:"created_at"`
    LastUsedAt time.Time `json:"last_used_at"`
}

// UpdateProfileRequest 更新个人信息请求
type UpdateProfileRequest struct {
    Nickname string `json:"nickname" validate:"omitempty,min=2,max=20"`
}

// BindEmailRequest 绑定邮箱请求
type BindEmailRequest struct {
    Email string `json:"email" validate:"required,email"`
}

// VerifyEmailRequest 验证邮箱请求
type VerifyEmailRequest struct {
    Token string `json:"token" validate:"required"`
}

// RecoveryRequest 账号恢复请求
type RecoveryRequest struct {
    RecoveryCode string `json:"recovery_code" validate:"required"`
    Email        string `json:"email" validate:"required,email"`
}

// RecoveryCodeResponse 恢复码响应
type RecoveryCodeResponse struct {
    RecoveryCode string `json:"recovery_code"`
    ExpireAt     time.Time `json:"expire_at"`
}

// ========== WebAuthn 相关 ==========

// WebAuthnRegistrationOptions 注册选项
type WebAuthnRegistrationOptions struct {
    Challenge string `json:"challenge"`
    UserID    string `json:"user_id"`
    Username  string `json:"username"`
}

// WebAuthnLoginOptions 登录选项
type WebAuthnLoginOptions struct {
    Challenge string `json:"challenge"`
}

// WebAuthnRegistrationResult 注册结果
type WebAuthnRegistrationResult struct {
    CredentialID string `json:"credential_id"`
    PublicKey    string `json:"public_key"` // base64
    SignCount    uint32 `json:"sign_count"`
    DeviceName   string `json:"device_name"`
}

// WebAuthnLoginResult 登录结果
type WebAuthnLoginResult struct {
    CredentialID string `json:"credential_id"`
    SignCount    uint32 `json:"sign_count"`
}
```

## API 接口

### 公开接口（无需认证）

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/public/passkey/register/begin | 开始 Passkey 注册 |
| POST | /api/public/passkey/register/finish | 完成 Passkey 注册 |
| POST | /api/public/passkey/login/begin | 开始 Passkey 登录 |
| POST | /api/public/passkey/login/finish | 完成 Passkey 登录 |
| POST | /api/public/recovery/request | 请求账号恢复（发送邮件） |
| POST | /api/public/recovery/verify | 验证恢复，绑定新 Passkey |

### 用户接口（需认证）

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/user/profile | 获取个人信息 |
| PUT | /api/user/profile | 更新个人信息（昵称） |
| POST | /api/user/email/bind | 绑定邮箱 |
| GET | /api/user/email/verify | 验证邮箱（点击邮件链接） |
| GET | /api/user/passkeys | 查看 Passkey 列表 |
| DELETE | /api/user/passkeys/:id | 删除 Passkey |
| GET | /api/user/recovery-code | 获取恢复码 |
| POST | /api/user/recovery-code/regenerate | 重新生成恢复码 |

## 接口流程

### Passkey 注册（首次使用）

```
客户端                              服务端
  │                                   │
  │  POST /passkey/register/begin    │
  │  (携带设备名称)                    │
  │ ─────────────────────────────────>│
  │                                   │ 创建临时用户
  │  { challenge, user_id, username } │ 生成 challenge
  │ <─────────────────────────────────│
  │                                   │
  │  调用 WebAuthn API                │
  │  Face ID / 指纹验证               │
  │                                   │
  │  POST /passkey/register/finish   │
  │  (携带 credential_id, public_key) │
  │ ─────────────────────────────────│
  │                                   │ 保存 Passkey
  │  { token, user }                  │ 返回 JWT
  │ <─────────────────────────────────│
```

### Passkey 登录（后续使用）

```
客户端                              服务端
  │                                   │
  │  POST /passkey/login/begin       │
  │ ─────────────────────────────────>│
  │                                   │ 查找用户所有 Passkey
  │  { challenge, allow_credentials } │ 生成 challenge
  │ <─────────────────────────────────│
  │                                   │
  │  调用 WebAuthn API                │
  │  Face ID / 指纹验证               │
  │                                   │
  │  POST /passkey/login/finish      │
  │  (携带 credential_id, signature)  │
  │ ─────────────────────────────────│
  │                                   │ 验证签名
  │  { token, user }                  │ 返回 JWT
  │ <─────────────────────────────────│
```

### 账号恢复（换设备）

```
1. 用户在新设备点击"恢复账号"
2. 输入恢复码 + 绑定邮箱
3. POST /recovery/request → 系统发送验证邮件
4. 用户点击邮件链接
5. POST /recovery/verify → 验证通过
6. 进入 Passkey 注册流程，绑定新设备
7. 账号数据迁移完成
```

## 异常处理

| 场景 | 错误码 | 提示信息 |
|-----|--------|---------|
| Passkey 不支持 | 400 | 当前设备不支持通行密钥，请升级系统 |
| 生物识别失败 | 400 | 验证失败，请重试 |
| 注册 challenge 过期 | 400 | 注册超时，请重新尝试 |
| 登录 challenge 过期 | 400 | 登录超时，请重新尝试 |
| 恢复码无效 | 400 | 恢复码无效或已使用 |
| 邮箱未绑定 | 400 | 该账号未绑定邮箱，无法恢复 |
| 邮箱已被绑定 | 400 | 该邮箱已被其他账号绑定 |
| Token 过期 | 401 | 登录已过期，请重新登录 |
| Passkey 被删除 | 401 | 该通行密钥已被删除 |

## 安全考虑

| 要求 | 说明 |
|-----|------|
| Challenge 有效期 | 5 分钟 |
| Challenge 单次有效 | 使用后立即失效 |
| 签名计数器 | 检测克隆凭证 |
| 恢复码加密 | bcrypt 存储 |
| 邮箱验证 | 24 小时有效 |
| Passkey 数量限制 | 单账号最多 5 个 |
