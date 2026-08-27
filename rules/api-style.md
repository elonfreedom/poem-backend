# API 规范

## 路由设计
- RESTful 风格：`GET /poems`, `POST /poems`, `GET /poems/:id`
- Admin 路由前缀：`/api/admin/`
- User 路由前缀：`/api/user/`
- 公开路由：`/api/public/`

## 请求格式
- JSON body，Content-Type: `application/json`
- 查询参数通过 URL query 传递
- 分页参数：`page`（默认 1）, `page_size`（默认 10，最大 100）

## 响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

## 错误码

采用 **HTTP 状态码 + 业务错误码双轨制**。详细错误码体系见 `@error-code.md`。

| HTTP 状态码 | 含义 | 使用场景 |
|-------------|------|----------|
| `400` | 请求参数错误 | 参数缺失、格式错误、JSON 解析失败 |
| `401` | 未认证 | 缺少 Token、Token 过期/无效 |
| `403` | 无权限 | 无权访问他人资源、非管理员 |
| `404` | 资源不存在 | 数据库查不到数据（用户/资源/连接等） |
| `409` | 业务冲突 | 连接令牌过期、状态不匹配 |
| `422` | 验证失败 | WebAuthn 验证、凭证验证失败 |
| `500` | 服务器内部错误 | 数据库操作失败、未预期的错误 |

**核心规则**：
- 数据库查不到数据 → **必须返回 404**，不能返回 400
- 数据库操作失败 → **必须返回 500**，附带具体错误信息
- 参数错误 → **必须标识具体字段名和原因**，不能笼统返回"参数错误"

---

# Fuego 接口注释规范

## 1. Handler 签名规范

### 1.1 有请求体的接口
使用 `fuego.ContextWithBody[T]`，T 必须是**具体结构体类型**，不能用 `any`：

```go
// ✅ 正确：使用具体类型
func (h *Handler) CreatePlan(c fuego.ContextWithBody[model.CreatePlanRequest]) (*model.CreatePlanResponse, error)

// ❌ 错误：使用 any 会导致 Swagger 无法显示入参
func (h *Handler) CreatePlan(c fuego.ContextNoBody) (any, error)
```

### 1.2 无请求体的接口
使用 `fuego.ContextNoBody`，返回**具体类型**：

```go
// ✅ 正确
func (h *Handler) GetProfile(c fuego.ContextNoBody) (*model.UserResponse, error)

// ❌ 错误：返回 any 会导致 Swagger 显示 unknown-interface
func (h *Handler) GetProfile(c fuego.ContextNoBody) (any, error)
```

## 2. 请求/响应结构体规范

### 2.1 结构体定义
每个请求/响应必须有 `description` 标签：

```go
// CreatePlanRequest 创建计划请求
type CreatePlanRequest struct {
    DailyCount int `json:"daily_count" validate:"required,min=1,max=50" description:"每日阅读数量（1-50）"`
    Duration   int `json:"duration" validate:"required,oneof=7 14 30 90" description:"计划天数（7/14/30/90）"`
}
```

### 2.2 标签顺序
结构体字段标签按以下顺序排列：
1. `json` - JSON 字段名
2. `validate` - 验证规则（可选）
3. `description` - 字段说明（必须）

## 3. 路由注册规范

### 3.1 必须使用的选项
每条路由**必须**包含以下选项：

```go
fuego.Post(group, "/reading-plans", readingPlanHandler.CreatePlan,
    fuego.OptionSummary("创建阅读计划"),                    // 简短摘要（2-6字）
    fuego.OptionOverrideDescription("创建新的阅读计划"),    // 详细描述
    fuego.OptionTags("阅读计划"),                           // 模块分类
)
```

### 3.2 OptionSummary vs OptionOverrideDescription
| 选项 | 用途 | 示例 |
|------|------|------|
| `OptionSummary` | 接口标题，2-6字 | `"获取诗歌列表"` |
| `OptionOverrideDescription` | 接口详细说明 | `"分页获取诗歌列表，支持按分类筛选"` |
| `OptionTags` | 模块分类 | `"诗歌浏览"` |

**注意**：必须使用 `OptionOverrideDescription`，不能用 `OptionDescription`，否则会被追加 controller 信息。

### 3.3 模块分类标签
接口按模块分类，使用 `OptionTags`：

| 标签 | 说明 |
|------|------|
| `Passkey 认证` | 注册、登录相关 |
| `个人信息` | 用户资料、Passkey 管理 |
| `诗歌浏览` | 诗歌列表、详情、搜索、推荐 |
| `收藏管理` | 收藏添加、取消、列表 |
| `阅读计划` | 计划创建、暂停、恢复、记录 |
| `打卡系统` | 打卡、记录、统计、日历、排行榜 |

## 4. 错误处理规范

> **完整错误码体系请参考 `@error-code.md`**

### 4.1 使用 Fuego 错误类型 + 精确描述

每个错误必须明确标识**哪个参数**、**什么问题**、**期望值是什么**：

```go
// 400 - 参数缺失（标识字段名）
return nil, fuego.BadRequestError{Title: "missing parameter", Detail: "token 不能为空"}

// 400 - 参数格式错误（标识字段名 + 期望格式）
return nil, fuego.BadRequestError{Title: "invalid parameter", Detail: fmt.Sprintf("user_id 格式错误: %s 不是有效 UUID", id)}

// 404 - 资源不存在（标识资源类型 + ID）
return nil, fuego.NotFoundError{Title: "user not found", Detail: fmt.Sprintf("用户不存在: id=%s", userID)}

// 401 - 未认证
return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录或 Token 已过期"}

// 403 - 无权限
return nil, fuego.ForbiddenError{Title: "forbidden", Detail: "无权访问此资源"}

// 500 - 数据库错误（标识操作 + 错误原因）
return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: id=%s, error=%v", id, err)}
```

### 4.2 禁止使用旧版 response 包
```go
// ❌ 错误：旧版 response 包不会被 Fuego 识别
return nil, response.BadRequest(c, "invalid request")

// ✅ 正确：使用 Fuego 错误类型
return nil, fuego.BadRequestError{Title: "missing parameter", Detail: "token 不能为空"}
```

### 4.3 禁止笼统返回 400

```go
// ❌ 错误：调用方无法区分是参数错误还是数据不存在
if err != nil {
    return nil, fuego.BadRequestError{Title: "bad request", Detail: err.Error()}
}

// ✅ 精确：区分"查不到"和"查询失败"
if errors.Is(err, pgx.ErrNoRows) {
    return nil, fuego.NotFoundError{Title: "user not found", Detail: fmt.Sprintf("用户不存在: id=%s", id)}
}
return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询失败: %v", err)}
```

## 5. 完整示例

### 5.1 Handler 文件示例
```go
package user

import (
    "github.com/go-fuego/fuego"
    "poem-backend/internal/middleware"
    "poem-backend/internal/model"
    "poem-backend/internal/service"
)

type PoemHandler struct {
    poemService *service.PoemService
}

func NewPoemHandler(poemService *service.PoemService) *PoemHandler {
    return &PoemHandler{poemService: poemService}
}

// List 获取诗歌列表
func (h *PoemHandler) List(c fuego.ContextNoBody) (*model.PoemListResponse, error) {
    // 从 context 获取用户 ID
    userID := middleware.GetUserIDFromContext(c.Context())
    if userID == "" {
        return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
    }

    // 解析查询参数
    page, _ := strconv.Atoi(c.QueryParam("page"))
    if page < 1 {
        page = 1
    }

    // 调用 service
    result, err := h.poemService.List(c.Context(), page, 10, nil)
    if err != nil {
        return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
    }

    return result, nil
}

// CreatePlan 创建阅读计划（有请求体）
func (h *ReadingPlanHandler) CreatePlan(c fuego.ContextWithBody[model.CreatePlanRequest]) (*model.CreatePlanResponse, error) {
    userID := middleware.GetUserIDFromContext(c.Context())
    if userID == "" {
        return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
    }

    body, err := c.Body()
    if err != nil {
        return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
    }

    result, err := h.planService.CreatePlan(c.Context(), userID, &body)
    if err != nil {
        return nil, fuego.BadRequestError{Title: "bad request", Detail: err.Error()}
    }

    return result, nil
}
```

### 5.2 Router 文件示例
```go
// [诗歌浏览] 诗歌相关
fuego.Get(userGroup, "/poems", poemHandler.List,
    fuego.OptionSummary("获取诗歌列表"),
    fuego.OptionOverrideDescription("分页获取诗歌列表"),
    fuego.OptionTags("诗歌浏览"),
)
fuego.Get(userGroup, "/poems/{id}", poemHandler.GetByID,
    fuego.OptionSummary("获取诗歌详情"),
    fuego.OptionOverrideDescription("根据 ID 获取诗歌详情"),
    fuego.OptionTags("诗歌浏览"),
)
```

## 6. 检查清单

新增接口时，确认以下事项：

- [ ] Handler 使用 `ContextWithBody[T]`（有入参）或 `ContextNoBody`（无入参）
- [ ] 请求结构体有 `description` 标签
- [ ] 返回类型为具体类型，不是 `any`
- [ ] 路由注册包含 `OptionSummary`
- [ ] 路由注册包含 `OptionOverrideDescription`（不是 `OptionDescription`）
- [ ] 路由注册包含 `OptionTags` 用于模块分类
- [ ] 错误使用 `fuego.*Error` 类型
- [ ] Swagger 文档正确显示入参出参
