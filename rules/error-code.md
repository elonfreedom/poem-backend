# 错误码规范

## 核心原则

1. **精确识别**：每个错误必须明确标识出哪个参数缺失、错误或哪个资源不存在
2. **不吞错误**：数据库查询失败、外部调用失败等必须通过接口返回具体信息
3. **HTTP 状态码 + 业务错误码双轨**：HTTP 状态码表示错误大类，body 中的 `code` 字段表示具体错误
4. **可前端解析**：错误结构包含 `error_code`，便于前端做国际化或条件判断
5. **集中管理**：所有错误码定义在 `pkg/errorcode/error_code.go`，禁止在 handler 中硬编码错误字符串

---

## 错误码枚举

所有错误码集中在 `pkg/errorcode/error_code.go` 中定义，使用方式：

```go
import "poem-backend/pkg/errorcode"

// handler 中使用
return nil, errorcode.ParamRequired("token").ToFuegoError()
return nil, errorcode.ConnectionNotFound(token).ToFuegoError()
return nil, errorcode.Internal("操作名称", err).ToFuegoError()
```

### 便捷构造函数

| 函数 | 用途 | 示例 |
|------|------|------|
| `ParamRequired(field)` | 必填参数缺失 | `errorcode.ParamRequired("token")` |
| `ParamInvalid(field, reason)` | 参数格式错误 | `errorcode.ParamInvalid("email", "不是有效邮箱")` |
| `BodyMalformed(err)` | JSON 解析失败 | `errorcode.BodyMalformed(err)` |
| `QueryRequired(field)` | 查询参数缺失 | `errorcode.QueryRequired("token")` |
| `Unauthorized()` | 未认证 | `errorcode.Unauthorized()` |
| `Forbidden(reason)` | 无权限 | `errorcode.Forbidden("无权访问")` |
| `UserNotFound(id)` | 用户不存在 | `errorcode.UserNotFound(userID)` |
| `ConnectionNotFound(token)` | 连接不存在 | `errorcode.ConnectionNotFound(token)` |
| `ConnectionExpired(token)` | 连接已过期 | `errorcode.ConnectionExpired(token)` |
| `ConnectionStatusInvalid(expected, actual)` | 状态错误 | `errorcode.ConnectionStatusInvalid("waiting", "connected")` |
| `NotConfirmed(status)` | 未确认 | `errorcode.NotConfirmed("waiting")` |
| `DatabaseError(op, err)` | 数据库错误 | `errorcode.DatabaseError("查询用户", err)` |
| `Internal(op, err)` | 内部错误 | `errorcode.Internal("生成令牌", err)` |
| `WebAuthnVerifyFailed(err)` | WebAuthn 验证失败 | `errorcode.WebAuthnVerifyFailed(err)` |

### 自定义错误

对于没有便捷构造函数的场景，使用 `New` 或 `Newf`：

```go
// 静态描述
return nil, errorcode.New(errorcode.ErrPasskeyNotFound, "passkey not found", "Passkey 不存在").ToFuegoError()

// 格式化描述
return nil, errorcode.Newf(errorcode.ErrParamOutOfRange, "page_size 超出范围", "允许 1-100, 实际 %d", pageSize).ToFuegoError()
```

---

## 错误响应格式

### 结构定义

```json
{
  "code": 400,
  "message": "请求参数错误",
  "error": "token 不能为空",
  "error_code": "ERR_TOKEN_REQUIRED",
  "detail": {
    "field": "token",
    "reason": "required"
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | HTTP 状态码（400/401/403/404/409/422/500） |
| `message` | string | 人类可读的错误摘要 |
| `error` | string | 具体错误描述（精确到字段和原因） |
| `error_code` | string | 机器可读的错误标识（大写蛇形命名） |
| `detail` | object | 可选，附加信息（如字段名、验证规则等） |

---

## 错误码体系

### 1xxx - 参数错误（HTTP 400）

| error_code | 说明 | 触发场景 |
|------------|------|----------|
| `ERR_PARAM_REQUIRED` | 必填参数缺失 | 请求体中缺少必填字段 |
| `ERR_PARAM_INVALID` | 参数格式错误 | 字段类型不匹配、格式非法 |
| `ERR_PARAM_OUT_OF_RANGE` | 参数超出范围 | 数值超出允许范围 |
| `ERR_PARAM_TOO_LONG` | 参数长度超限 | 字符串超过最大长度 |
| `ERR_BODY_MALFORMED` | 请求体格式错误 | JSON 解析失败 |
| `ERR_QUERY_REQUIRED` | 必填查询参数缺失 | URL query 中缺少必填参数 |

### 2xxx - 认证授权错误（HTTP 401/403）

| error_code | 说明 | 触发场景 |
|------------|------|----------|
| `ERR_UNAUTHORIZED` | 未认证 | 缺少 Token 或 Token 无效 |
| `ERR_TOKEN_EXPIRED` | Token 已过期 | JWT 超过有效期 |
| `ERR_TOKEN_INVALID` | Token 格式错误 | JWT 签名验证失败 |
| `ERR_FORBIDDEN` | 无权限 | 无权访问他人资源 |
| `ERR_ADMIN_REQUIRED` | 需要管理员权限 | 非 admin 角色访问管理接口 |

### 3xxx - 资源不存在（HTTP 404）

| error_code | 说明 | 触发场景 |
|------------|------|----------|
| `ERR_USER_NOT_FOUND` | 用户不存在 | 根据 ID 查不到用户 |
| `ERR_PASSKEY_NOT_FOUND` | Passkey 不存在 | 根据 ID 查不到 Passkey |
| `ERR_POEM_NOT_FOUND` | 诗歌不存在 | 根据 ID 查不到诗歌 |
| `ERR_PLAN_NOT_FOUND` | 阅读计划不存在 | 根据 ID 查不到计划 |
| `ERR_CONNECTION_NOT_FOUND` | 连接不存在 | 连接令牌无效或已过期 |
| `ERR_CREDENTIAL_NOT_FOUND` | 凭证不存在 | WebAuthn 凭证查不到 |

### 4xxx - 业务冲突（HTTP 409）

| error_code | 说明 | 触发场景 |
|------------|------|----------|
| `ERR_TOKEN_EXPIRED` | 连接令牌已过期 | 跨设备连接超过有效期 |
| `ERR_TOKEN_INVALID` | 连接令牌无效 | 连接令牌状态不匹配 |
| `ERR_DEVICE_NAME_EXISTS` | 设备名称已存在 | 重复的设备名 |
| `ERR_PASSKEY_ALREADY_EXISTS` | Passkey 已存在 | 重复注册 |

### 5xxx - 验证失败（HTTP 422）

| error_code | 说明 | 触发场景 |
|------------|------|----------|
| `ERR_WEBAUTHN_VERIFY_FAILED` | WebAuthn 验证失败 | 签名或挑战验证不通过 |
| `ERR_CREDENTIAL_VERIFY_FAILED` | 凭证验证失败 | Passkey 登录时凭证不匹配 |
| `ERR_SIGNATURE_INVALID` | 签名无效 | 数据签名验证失败 |

### 9xxx - 服务器错误（HTTP 500）

| error_code | 说明 | 触发场景 |
|------------|------|----------|
| `ERR_INTERNAL` | 服务器内部错误 | 未预期的错误 |
| `ERR_DATABASE` | 数据库错误 | 数据库操作失败 |
| `ERR_DATABASE_QUERY_FAILED` | 数据库查询失败 | SELECT 执行失败 |
| `ERR_DATABASE_INSERT_FAILED` | 数据库写入失败 | INSERT 执行失败 |
| `ERR_DATABASE_UPDATE_FAILED` | 数据库更新失败 | UPDATE 执行失败 |
| `ERR_DATABASE_DELETE_FAILED` | 数据库删除失败 | DELETE 执行失败 |
| `ERR_EXTERNAL_SERVICE` | 外部服务错误 | 第三方服务调用失败 |

---

## 数据库错误处理规范

### 查询不到数据

**禁止**：将"查不到数据"静默处理为通用 400 错误。

```go
// ❌ 错误：笼统返回 400，调用方无法区分是参数错误还是数据不存在
user, err := userRepo.GetByID(ctx, id)
if err != nil {
    return nil, fuego.BadRequestError{Title: "bad request", Detail: err.Error()}
}

// ✅ 精确：区分"查不到"和"查询失败"
user, err := userRepo.GetByID(ctx, id)
if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, fuego.NotFoundError{
            Title: "user not found",
            Detail: fmt.Sprintf("用户不存在: id=%s", id),
        }
    }
    return nil, fuego.InternalServerError{
        Title: "database error",
        Detail: fmt.Sprintf("查询用户失败: id=%s, error=%v", id, err),
    }
}
```

### 数据库操作失败

```go
// ✅ 写入失败
if err := userRepo.Create(ctx, user); err != nil {
    return nil, fuego.InternalServerError{
        Title: "database error",
        Detail: fmt.Sprintf("创建用户失败: error=%v", err),
    }
}

// ✅ 更新失败
if err := passkeyRepo.Update(ctx, passkey); err != nil {
    return nil, fuego.InternalServerError{
        Title: "database error",
        Detail: fmt.Sprintf("更新 Passkey 失败: id=%s, error=%v", passkey.ID, err),
    }
}

// ✅ 删除失败
if err := passkeyRepo.Delete(ctx, id); err != nil {
    return nil, fuego.InternalServerError{
        Title: "database error",
        Detail: fmt.Sprintf("删除 Passkey 失败: id=%s, error=%v", id, err),
    }
}
```

---

## 参数校验规范

### 必填参数缺失

```go
// ❌ 错误：不告诉调用方缺了什么
if body.Token == "" {
    return nil, fuego.BadRequestError{Title: "bad request", Detail: "参数错误"}
}

// ✅ 精确：明确标识缺失字段
if body.Token == "" {
    return nil, fuego.BadRequestError{
        Title: "missing parameter",
        Detail: "token 不能为空",
    }
}
```

### 参数格式错误

```go
// ✅ 精确：标识字段名和期望格式
if _, err := uuid.Parse(body.UserID); err != nil {
    return nil, fuego.BadRequestError{
        Title: "invalid parameter",
        Detail: fmt.Sprintf("user_id 格式错误: %s 不是有效的 UUID", body.UserID),
    }
}
```

### 参数超出范围

```go
// ✅ 精确：标识字段名和允许范围
if body.PageSize < 1 || body.PageSize > 100 {
    return nil, fuego.BadRequestError{
        Title: "parameter out of range",
        Detail: fmt.Sprintf("page_size=%d 超出范围: 允许 1-100", body.PageSize),
    }
}
```

---

## Handler 层错误处理模板

```go
import "poem-backend/pkg/errorcode"

func (h *AuthHandler) AddDeviceConnect(c fuego.ContextWithBody[AddDeviceConnectRequest]) (map[string]any, error) {
    body, err := c.Body()
    if err != nil {
        return nil, errorcode.BodyMalformed(err).ToFuegoError()
    }

    if body.Token == "" {
        return nil, errorcode.ParamRequired("token").ToFuegoError()
    }

    conn, ok := h.ConnectionStore.Get(body.Token)
    if !ok {
        return nil, errorcode.ConnectionNotFound(body.Token).ToFuegoError()
    }

    if time.Now().After(conn.ExpiresAt) {
        return nil, errorcode.ConnectionExpired(body.Token).ToFuegoError()
    }

    if conn.Status != ConnectionStatusWaiting {
        return nil, errorcode.ConnectionStatusInvalid("waiting", string(conn.Status)).ToFuegoError()
    }

    // ... 业务逻辑
}
```

---

## Service 层错误处理模板

```go
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            // 资源不存在 → 404
            return nil, fuego.NotFoundError{
                Title:  "user not found",
                Detail: fmt.Sprintf("用户不存在: id=%s", userID),
            }
        }
        // 数据库查询失败 → 500
        return nil, fuego.InternalServerError{
            Title:  "database error",
            Detail: fmt.Sprintf("查询用户失败: id=%s, error=%v", userID, err),
        }
    }
    return user, nil
}
```

---

## 错误码使用检查清单

新增接口或修改错误处理时，确认：

- [ ] 使用 `errorcode.*` 构造函数，不硬编码错误字符串
- [ ] 参数缺失时返回具体字段名（如 `errorcode.ParamRequired("token")`）
- [ ] 参数格式错误时返回字段名和期望格式
- [ ] 数据库查不到数据时返回 404 + 具体资源标识
- [ ] 数据库操作失败时返回 500 + 具体错误信息
- [ ] 权限不足时返回 403 + 具体原因
- [ ] 资源冲突时返回 409 + 冲突详情
- [ ] 所有错误都包含人类可读的 Detail 字段
- [ ] 错误信息中不包含敏感数据（密码、密钥等）
- [ ] 新增错误码在 `pkg/errorcode/error_code.go` 中定义
