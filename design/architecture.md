# 技术架构

> 本文档定义诗歌后端服务的技术架构设计，包含技术栈、目录结构、数据流、API 规范和部署架构。

---

## 1. 技术栈

| 类别 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 语言 | Go | 1.26 | 后端主语言 |
| HTTP 框架 | Fuego | - | 类型安全路由 + OpenAPI 自动生成 |
| 数据库 | PostgreSQL | - | 主数据库 |
| 驱动 | pgx/v5 | - | PostgreSQL 连接池 |
| 认证 | JWT | - | Token 认证 |
| 登录 | WeChat / Apple | - | 第三方登录 |
| 简繁转换 | gocc | - | 纯 Go 实现，无 C 依赖 |
| 拼音 | go-pinyin | - | 汉字转拼音（带声调） |

---

## 2. 系统架构

```
┌─────────────────────────────────────────────────┐
│                    Client                        │
│         (Web / iOS / Android / Admin)            │
└─────────────────────┬───────────────────────────┘
                      │ HTTP / JSON
┌─────────────────────▼───────────────────────────┐
│                  Fuego Router                    │
│         (路由注册 + OpenAPI 文档生成)              │
├─────────────────────────────────────────────────┤
│                  Middleware                      │
│         (Auth / CORS / Logger)                   │
├──────────────────────┬──────────────────────────┤
│    Admin Handler     │      User Handler         │
├──────────────────────┼──────────────────────────┤
│   Admin Service      │     User Service          │
├──────────────────────┴──────────────────────────┤
│                  Repository                      │
├─────────────────────────────────────────────────┤
│              PostgreSQL (pgx/v5)                 │
└─────────────────────────────────────────────────┘
```

---

## 3. 目录结构

```
cmd/server/              → 服务入口
internal/
├── config/              → 环境变量配置 (SERVER_PORT, DB_*, JWT_*)
├── middleware/          → HTTP 中间件 (auth, cors, logger)
├── handler/             → 路由处理
│   ├── admin/           → 管理后台处理器（诗歌、作者、分类、工具）
│   └── user/            → 用户端处理器（认证、诗歌、收藏、打卡）
├── model/               → 数据模型
│   ├── admin/           → Admin 请求/响应结构体
│   └── user/            → User 请求/响应结构体
├── repository/          → 数据访问层（SQL 查询）
├── service/             → 业务逻辑层
│   ├── admin/           → 管理后台业务逻辑
│   └── user/            → 用户端业务逻辑
└── router/              → 路由注册
    ├── admin_router.go  → 管理后台路由
    └── user_router.go   → 用户端路由
pkg/
├── database/            → PostgreSQL 连接池
├── response/            → 统一 JSON 响应格式
├── errorcode/           → 错误码定义
├── convert/             → 简繁体转换
├── pinyin/              → 汉字转拼音
└── migrate/             → 数据库迁移
```

---

## 4. 分层职责与数据流

### 4.1 分层职责

| 层级 | 目录 | 职责 | 禁止 |
|------|------|------|------|
| Router | `internal/router/` | 注册路由、配置中间件、初始化依赖 | 业务逻辑 |
| Handler | `internal/handler/` | 解析请求、参数校验、调用 Service | 数据库操作 |
| Service | `internal/service/` | 业务逻辑、事务协调、数据转换 | 直接操作 HTTP |
| Repository | `internal/repository/` | 数据库 CRUD、SQL 查询 | 业务逻辑 |
| Model | `internal/model/` | 数据结构定义 | - |

### 4.2 依赖方向

```
Router → Handler → Service → Repository → Database
                         ↓
                      Model (DTO)
```

**规则**：单向依赖，禁止反向依赖。

### 4.3 数据流

```
Request → Router → Middleware(Auth) → Handler → Service → Repository → DB
                                                                    ↓
Response ← Router ← Handler ← Service ← Repository ←─────── Data
```

---

## 5. API 路由规范

### 5.1 RESTful 设计原则

API 严格遵循 **RESTful 风格**：

1. **资源导向**：URL 表示资源（名词），不表示动作（动词）
2. **HTTP 方法语义**：
   - `GET` — 简单查询（参数在 URL，幂等、安全）
   - `POST` — 创建资源、触发操作、复杂查询（Body 传参）
   - `PUT` — 全量替换（幂等）
   - `PATCH` — 部分更新
   - `DELETE` — 删除（幂等）
3. **复数名词**：资源集合用复数（`/poems`，非 `/poem`）
4. **嵌套资源**：用路径表达从属关系（`/users/{id}/favorites`）
5. **无动词 URL**：例外仅限非 CRUD 的操作型端点（如 `/auth/login`、`/tools/generate-pinyin`）

### 5.2 前缀规范

| 模块 | 前缀 | 认证 | 说明 |
|------|------|------|------|
| 公开接口 | `/api/public/` | 无需 | 登录、注册、健康检查、公开浏览 |
| 用户端 | `/api/user/` | JWT | 普通用户使用 |
| 管理后台 | `/api/admin/` | JWT + Admin | 管理员使用 |

### 5.3 HTTP 方法使用规则

| 方法 | 用途 | 幂等 | 安全 | 示例 |
|------|------|------|------|------|
| `GET` | 简单查询（参数在 URL） | ✅ | ✅ | `GET /poems`, `GET /poems/{id}` |
| `POST` | 创建资源 / 触发操作 / 复杂查询 | ❌ | ❌ | `POST /poems`, `POST /auth/login`, `POST /poems/search` |
| `PUT` | 全量替换 | ✅ | ❌ | `PUT /poems/{id}` |
| `PATCH` | 部分更新 | ❌ | ❌ | `PATCH /poems/{id}/status` |
| `DELETE` | 删除资源 | ✅ | ❌ | `DELETE /poems/{id}` |

> **关于复杂查询**：当前使用 `POST` 处理需要 Body 的复杂查询（如全文搜索、多条件筛选）。未来若框架支持 `QUERY` 方法（HTTP 扩展方法，语义=带 Body 的 GET），可迁移至 `QUERY` 以保持语义准确。

### 5.4 URL 命名约定

```
✅ GET    /poems              # 简单查询（参数在 URL）
✅ GET    /poems/{id}         # 获取单首诗歌
✅ POST   /poems/search       # 复杂搜索（Body 传结构化条件）
✅ POST   /poems/filter       # 复杂过滤（Body 传多条件组合）
✅ POST   /poems              # 创建诗歌
✅ PUT    /poems/{id}         # 全量更新诗歌
✅ DELETE /poems/{id}         # 删除诗歌

✅ GET    /users/{id}/favorites       # 获取用户收藏列表
✅ POST   /favorites                  # 添加收藏（poem_id 在 body）
✅ DELETE /favorites/{poem_id}        # 取消收藏

✅ POST   /auth/login                 # 登录（操作型，允许动词）
✅ POST   /tools/generate-pinyin      # 工具（操作型，允许动词）
```

**`GET` vs `POST`（复杂查询）选择**：
```
✅ GET    /poems?page=1&dynasty=唐          # 简单查询 → URL 参数足够
✅ POST   /poems/search {keyword, tags}     # 复杂查询 → Body 传递结构化条件
```

**禁止**：
```
❌ GET    /getPoems           # 动词开头
❌ POST   /createPoem         # 动词开头
❌ GET    /poem               # 单数名词
❌ POST   /poems/delete       # 动词混合
```

### 5.5 完整路由清单

#### 管理后台 `/api/admin/*`

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/auth/login` | 管理员登录（公开） |
| `GET` | `/user/info` | 获取管理员信息 |
| `GET` | `/auth/codes` | 获取权限码 |
| `POST` | `/auth/logout` | 退出登录 |
| **诗歌管理** | | |
| `GET` | `/poems` | 获取诗歌列表（支持筛选、分页） |
| `POST` | `/poems` | 创建诗歌 |
| `POST` | `/poems/import` | 批量导入诗歌 |
| `GET` | `/poems/{id}` | 获取诗歌详情 |
| `PUT` | `/poems/{id}` | 更新诗歌 |
| `DELETE` | `/poems/{id}` | 删除诗歌 |
| `PUT` | `/poems/{id}/status` | 更新诗歌状态 |
| `PUT` | `/poems/batch/status` | 批量更新状态 |
| **作者管理** | | |
| `GET` | `/authors` | 获取作者列表 |
| `POST` | `/authors` | 创建作者 |
| `GET` | `/authors/{id}` | 获取作者详情 |
| `GET` | `/authors/options` | 作者下拉搜索 |
| `PUT` | `/authors/{id}` | 更新作者 |
| `DELETE` | `/authors/{id}` | 删除作者 |
| `POST` | `/authors/batch/match` | 批量匹配诗歌关联 |
| **分类管理** | | |
| `GET` | `/categories` | 获取分类列表 |
| `POST` | `/categories` | 创建分类 |
| `PUT` | `/categories/{id}` | 更新分类 |
| `DELETE` | `/categories/{id}` | 删除分类 |
| **标签管理** | | |
| `GET` | `/tags` | 获取标签列表 |
| `POST` | `/tags` | 创建标签 |
| `DELETE` | `/tags/{id}` | 删除标签 |
| **数据统计** | | |
| `GET` | `/stats/overview` | 总览统计 |
| `GET` | `/stats/daily` | 每日统计 |
| `GET` | `/stats/poems/hot` | 热门诗歌 |
| `GET` | `/stats/users/growth` | 用户增长 |
| **Banner 管理** | | |
| `GET` | `/banners` | 获取 Banner 列表 |
| `POST` | `/banners` | 创建 Banner |
| `PUT` | `/banners/{id}` | 更新 Banner |
| `DELETE` | `/banners/{id}` | 删除 Banner |
| **公告管理** | | |
| `GET` | `/announcements` | 获取公告列表 |
| `POST` | `/announcements` | 创建公告 |
| `PUT` | `/announcements/{id}` | 更新公告 |
| `DELETE` | `/announcements/{id}` | 删除公告 |
| **系统配置** | | |
| `GET` | `/config` | 获取配置列表 |
| `GET` | `/config/{key}` | 获取单个配置 |
| `PUT` | `/config` | 更新配置 |
| **用户管理** | | |
| `GET` | `/users` | 获取前端用户列表 |
| `GET` | `/users/{id}` | 获取用户详情 |
| `PUT` | `/users/{id}/status` | 更新用户状态 |
| **工具模块** | | |
| `POST` | `/tools/convert-simplified` | 批量生成简体 |
| `POST` | `/tools/generate-pinyin` | 批量生成拼音 |
| `POST` | `/tools/generate-authors` | 提取作者 |

#### 用户端 `/api/user/*`（需 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| **个人信息** | | |
| `GET` | `/profile` | 获取个人信息 |
| `PUT` | `/profile` | 更新个人信息 |
| `GET` | `/passkeys` | 获取 Passkey 列表 |
| `DELETE` | `/passkeys/{id}` | 删除 Passkey |
| **诗歌浏览** | | |
| `GET` | `/poems` | 获取诗歌列表 |
| `GET` | `/poems/{id}` | 获取诗歌详情 |
| `GET` | `/poems/search` | 搜索诗歌 |
| **收藏管理** | | |
| `GET` | `/favorites` | 获取收藏列表 |
| `POST` | `/favorites` | 添加收藏 |
| `DELETE` | `/favorites/{poem_id}` | 取消收藏 |
| **阅读计划** | | |
| `POST` | `/reading-plans` | 创建阅读计划 |
| `GET` | `/reading-plans/current` | 获取当前计划 |
| `PUT` | `/reading-plans/{id}/pause` | 暂停计划 |
| `PUT` | `/reading-plans/{id}/resume` | 恢复计划 |
| `GET` | `/reading-plans/{id}/progress` | 获取计划进度 |
| `POST` | `/reading-plans/log` | 记录阅读 |
| **共享计划** | | |
| `GET` | `/shared-plans` | 浏览共享库（公开） |
| `GET` | `/shared-plans/{id}` | 计划详情（公开） |
| `POST` | `/shared-plans` | 创建共享计划 |
| `GET` | `/shared-plans/mine` | 我的计划 |
| `PUT` | `/shared-plans/{id}` | 更新计划 |
| `PUT` | `/shared-plans/{id}/publish` | 发布计划 |
| `PUT` | `/shared-plans/{id}/unpublish` | 取消发布 |
| `DELETE` | `/shared-plans/{id}` | 删除计划 |
| `POST` | `/shared-plans/{id}/subscribe` | 订阅计划 |
| `DELETE` | `/shared-plans/{id}/subscribe` | 取消订阅 |
| `PUT` | `/subscriptions/{id}/start-date` | 设置开始日期 |
| `GET` | `/subscriptions` | 我的订阅 |
| `GET` | `/subscriptions/{id}/today` | 今日诗文 |
| `POST` | `/subscriptions/{id}/checkin` | 打卡 |
| `POST` | `/subscriptions/{id}/skip` | 跳过天数 |
| `GET` | `/subscriptions/{id}/progress` | 订阅进度 |
| `GET` | `/subscriptions/{id}/checkins` | 打卡记录 |
| `PUT` | `/subscriptions/{id}/pause` | 暂停订阅 |
| `PUT` | `/subscriptions/{id}/resume` | 恢复订阅 |
| **打卡系统** | | |
| `POST` | `/checkins` | 打卡 |
| `GET` | `/checkins` | 打卡记录 |
| `GET` | `/checkins/stats` | 打卡统计 |
| `GET` | `/checkins/calendar` | 打卡日历 |
| `GET` | `/checkins/ranking` | 排行榜 |

#### 公开接口 `/api/public/*`

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/passkey/register/begin` | 开始 Passkey 注册（device_name 由前端自动检测） |
| `POST` | `/passkey/register/finish` | 完成 Passkey 注册 |
| `POST` | `/passkey/login/begin` | 开始 Passkey 登录 |
| `POST` | `/passkey/login/finish` | 完成 Passkey 登录 |
| `GET` | `/poems/daily` | 每日推荐诗歌 |
| `GET` | `/shared-plans` | 浏览共享计划库 |
| `GET` | `/shared-plans/{id}` | 共享计划详情 |
| `POST` | `/passkeys/add/connect` | 新设备连接 |
| `GET` | `/passkeys/add/status` | 查询连接状态（公开） |
| `POST` | `/passkeys/add/finish` | 完成设备注册 |
| `POST` | `/passkeys/add/reject` | 放弃绑定 |

---

## 6. 请求/响应规范

### 6.1 请求格式
- Content-Type: `application/json`
- Body 使用 JSON 格式
- 字段命名：snake_case

### 6.2 分页参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码（≥1） |
| `per_page` / `page_size` | int | 10 | 每页数量（1-50） |

### 6.3 统一响应结构

> **规范**：所有 API 使用统一的 JSON 响应格式。成功响应使用 `code: 0`，业务错误使用业务错误码（见 §11），HTTP 错误由 Fuego 框架自动处理。

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

### 6.4 分页响应结构

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [],
    "total": 100
  }
}
```

### 6.5 响应函数使用规范

| 函数 | 使用场景 | 返回类型 | code 值 |
|------|----------|----------|---------|
| `response.OK[T](data)` | 类型安全的成功响应 | `*APIResponse[T]` | 0 |
| `response.PageOK[T](items, total)` | 分页响应 | `*APIResponse[PageData[T]]` | 0 |
| `response.Success(data)` | 通用成功响应（User 端） | `*APIResponse[any]` | 0 |

> **规范**：所有响应统一使用 `code: 0` 表示成功，业务错误使用 §11 定义的错误码。

### 6.6 操作状态响应

对于删除、更新等无数据返回的操作，使用统一的 `SimpleResponse`：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "success": true
  }
}
```

或状态响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "status": "updated"
  }
}
```

---

## 7. Fuego 框架规范

### 7.1 Handler 签名

```go
// 有请求体 — 使用具体类型 T，禁止 any
func (h *Handler) CreatePoem(c fuego.ContextWithBody[model.CreatePoemRequest]) (*model.CreatePoemResponse, error)

// 无请求体
func (h *Handler) GetPoem(c fuego.ContextNoBody) (*model.PoemResponse, error)
```

### 7.2 结构体标签

```go
// 标签顺序：json → validate → description（必须）
type CreatePoemRequest struct {
    Title   string `json:"title" validate:"required" description:"诗歌标题"`
    Author  string `json:"author" validate:"required" description:"作者"`
    Status  string `json:"status" validate:"oneof=draft published" description:"状态"`
}
```

### 7.3 路由注册

```go
// 每条路由必须包含三个 Option
fuego.Post(group, "/poems", poemHandler.Create,
    fuego.OptionSummary("创建诗歌"),                    // 2-6 字
    fuego.OptionOverrideDescription("创建新的诗歌作品"),  // 详细描述
    fuego.OptionTags("诗歌管理"),                        // 模块分类
)
```

### 7.4 Model 设计模式

| 类型 | 用途 | 示例 |
|------|------|------|
| Model | 数据库实体 | `Poem` |
| Request | 入参验证 | `CreatePoemRequest`, `UpdatePoemRequest` |
| Response | API 输出 | `PoemResponse` |

通过 `ToResponse()` 方法转换，隐藏敏感字段。

### 7.5 响应包装

```go
// Admin 端
return response.OK(CreatePoemResponse{ID: poemID}), nil

// User 端
return response.Success(result), nil
```

---

## 8. 认证架构

### 8.1 JWT 认证流程
```
Client → [Login] → Server 发放 JWT
Client → [Request + Bearer Token] → Middleware 验证 → 注入 Context → Handler
```

### 8.2 角色控制
- `role=user` — 普通用户，访问 `/api/user/*`
- `role=admin` — 管理员，访问 `/api/admin/*`

### 8.3 Passkey 零输入方案
```
注册：前端自动从 User-Agent 检测设备名 → POST register/begin { device_name } → WebAuthn → POST register/finish
登录：POST login/begin → WebAuthn → POST login/finish

⚠️ device_name 由前端自动检测（如 "Chrome on Mac"），后端不做非空校验
```

### 8.3 Context 传递
```go
// 中间件注入
ctx := context.WithValue(ctx, middleware.UserIDKey, userID)

// Handler 获取
userID := middleware.GetUserIDFromContext(ctx)
```

---

## 9. 数据库设计原则

### 9.1 命名约定
- 表名：复数形式，snake_case（`poems`, `authors`, `reading_plans`）
- 字段：snake_case（`created_at`, `user_id`, `daily_count`）
- 主键：`id`（BIGSERIAL）
- 外键：`xxx_id`（BIGINT）
- 时间戳：`created_at`, `updated_at`（TIMESTAMPTZ）

### 9.2 迁移规范
- 文件命名：`NNN_description.up.sql` / `NNN_description.down.sql`
- 必须包含 `IF NOT EXISTS` / `IF EXISTS` 确保幂等

---

## 10. 部署架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Docker    │────▶│  Aliyun ECS │────▶│  PostgreSQL │
│   Build     │     │  (Shanghai) │     │  (RDS)      │
└─────────────┘     └─────────────┘     └─────────────┘
```

- 部署脚本：`deploy.sh`
- 容器化：Docker + Docker Compose
- 目标环境：阿里云 ECS（上海区域）

---

## 11. 错误码规范

### 11.1 错误码格式

业务错误码为 **4 位数字**，按类别分段：

| 码段 | 类别 | HTTP 状态码 | 说明 |
|------|------|-------------|------|
| 1xxx | 参数错误 | 400 | 请求参数缺失、格式错误、超出范围 |
| 2xxx | 认证授权 | 401/403 | 未认证、Token 过期、无权限 |
| 3xxx | 资源不存在 | 404 | 用户、诗歌、计划等资源不存在 |
| 4xxx | 业务冲突 | 400 | 重复订阅、状态冲突等 |
| 5xxx | 验证失败 | 422 | WebAuthn、凭证验证失败 |
| 9xxx | 服务器错误 | 500 | 数据库、内部服务错误 |

### 11.2 错误码清单

#### 参数错误 (1xxx)

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 1001 | `ErrParamRequired` | 必填参数缺失 |
| 1002 | `ErrParamInvalid` | 参数格式错误 |
| 1003 | `ErrParamOutOfRange` | 参数超出范围 |
| 1004 | `ErrParamTooLong` | 参数长度超限 |
| 1005 | `ErrBodyMalformed` | 请求体格式错误 |
| 1006 | `ErrQueryRequired` | 必填查询参数缺失 |

#### 认证授权错误 (2xxx)

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 2001 | `ErrUnauthorized` | 未认证 |
| 2002 | `ErrTokenExpired` | Token 已过期 |
| 2003 | `ErrTokenInvalid` | Token 格式错误 |
| 2004 | `ErrForbidden` | 无权限 |
| 2005 | `ErrAdminRequired` | 需要管理员权限 |

#### 资源不存在 (3xxx)

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 3001 | `ErrUserNotFound` | 用户不存在 |
| 3002 | `ErrPasskeyNotFound` | Passkey 不存在 |
| 3003 | `ErrPoemNotFound` | 诗歌不存在 |
| 3004 | `ErrPlanNotFound` | 阅读计划不存在 |
| 3005 | `ErrConnectionNotFound` | 连接不存在 |
| 3006 | `ErrCredentialNotFound` | 凭证不存在 |

#### 业务冲突 (4xxx)

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 4001 | `ErrConnectionExpired` | 连接令牌已过期 |
| 4002 | `ErrConnectionInvalid` | 连接令牌无效 |
| 4003 | `ErrConnectionStatus` | 连接状态错误 |
| 4004 | `ErrNotConfirmed` | 设备 A 尚未确认 |
| 4005 | `ErrDeviceNameExists` | 设备名称已存在 |
| 4006 | `ErrPasskeyExists` | Passkey 已存在 |

#### 验证失败 (5xxx)

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 5001 | `ErrWebAuthnVerify` | WebAuthn 验证失败 |
| 5002 | `ErrCredentialVerify` | 凭证验证失败 |
| 5003 | `ErrSignatureInvalid` | 签名无效 |

#### 服务器错误 (9xxx)

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 9001 | `ErrInternal` | 服务器内部错误 |
| 9002 | `ErrDatabase` | 数据库错误 |
| 9003 | `ErrDatabaseQuery` | 数据库查询失败 |
| 9004 | `ErrDatabaseInsert` | 数据库写入失败 |
| 9005 | `ErrDatabaseUpdate` | 数据库更新失败 |
| 9006 | `ErrDatabaseDelete` | 数据库删除失败 |
| 9007 | `ErrExternalService` | 外部服务错误 |

### 11.3 错误响应格式

```json
{
  "code": 1001,
  "message": "missing parameter",
  "error": "field_name 不能为空",
  "data": null
}
```

### 11.4 使用规范

- **Service 层**：使用 `fuego.*Error` 类型返回错误
- **Handler 层**：透传 Service 错误，不做二次包装
- **错误码包**：`pkg/errorcode` 提供业务错误码定义和 `ToFuegoError()` 转换
- **错误标题**：使用英文 snake_case，如 `"invalid_param"`, `"not_found"`

---

## 12. 文档验证

启动服务后访问：
- Swagger UI: http://localhost:8080/swagger
- OpenAPI JSON: http://localhost:8080/swagger/openapi.json

### 检查清单

- [ ] Handler 使用 `ContextWithBody[T]` 或 `ContextNoBody`
- [ ] 请求结构体有 `description` 标签
- [ ] 返回类型为具体类型，不是 `any`
- [ ] 路由注册包含 `OptionSummary` + `OptionOverrideDescription` + `OptionTags`
- [ ] 错误使用 `fuego.*Error` 类型
- [ ] Swagger 文档正确显示入参出参
- [ ] 错误码符合 §11 规范
