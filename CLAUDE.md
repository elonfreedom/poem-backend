# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

诗歌应用后端服务，基于 Go + Fuego 框架。采用分层架构：handler → service → repository → database。

## Tech Stack

- **Language**: Go 1.26
- **Framework**: Fuego (HTTP framework with OpenAPI support)
- **Database**: PostgreSQL via pgx/v5 connection pool
- **Auth**: JWT + WeChat/Apple login

## Common Commands

```bash
# Install dependencies
go mod tidy

# Run the server (default port 8080)
go run cmd/server/main.go

# Run tests
go test ./...

# Run a single test
go test -v ./internal/handler/...

# Build
go build -o bin/server cmd/server/main.go

# Check code
go vet ./...

# 部署到阿里云 ECS（上海）
./deploy.sh all    # 构建 + 部署全流程
./deploy.sh build  # 仅构建镜像
./deploy.sh deploy # 仅部署（需先 build）
```

## Architecture

```
cmd/server/          → 服务入口
internal/
├── config/          → 环境变量配置 (SERVER_PORT, DB_*, JWT_*)
├── middleware/       → HTTP 中间件 (auth, cors, logger)
├── handler/         → 路由处理 (admin/, user/ 两个模块)
├── model/           → 数据模型 + 请求/响应结构体
├── repository/      → 数据访问层
├── service/         → 业务逻辑层
└── router/          → 路由注册
pkg/
├── database/        → PostgreSQL 连接池
└── response/        → 统一 JSON 响应格式 {code, message, data}
```

## Key Patterns

**响应格式**: 统一使用 `pkg/response` 的 `Success()` 函数，返回 `{code: 200, message: "success", data: ...}` 结构。

**认证**: JWT token 通过 `Authorization: Bearer <token>` 传递。Admin 路由需 `role=admin`。用户信息存储在 context 中，通过 `middleware.GetUserIDFromContext(ctx)` 获取。

**Model 设计**: 每个实体有 Model（数据库字段）、Response（API 输出，隐藏敏感字段）、Request（入参验证）三类结构体。通过 `ToResponse()` 方法转换。

**Fuego 框架**: 路由定义使用 `fuego` 的类型安全方式，handler 签名包含 `*fuego.Context[T]` 泛型参数。

### Fuego 接口文档规范（必须遵守）

**Handler 签名**：
- 有入参：`func (h *Handler) Method(c fuego.ContextWithBody[RequestType]) (ResponseType, error)`
- 无入参：`func (h *Handler) Method(c fuego.ContextNoBody) (ResponseType, error)`
- **禁止**使用 `any` 作为请求/响应类型

**路由注册**（每条路由必须包含）：
```go
fuego.Post(group, "/path", handler.Method,
    fuego.OptionSummary("简短摘要"),           // 2-6 字
    fuego.OptionOverrideDescription("详细描述"), // 必须用 Override
    fuego.OptionTags("模块分类"),
)
```

**错误处理**：使用 `fuego.*Error` 类型，不要用 `response.BadRequest` 等

**结构体标签**：每个字段必须有 `description` 标签

### User 端统一响应格式（必须遵守）

**所有 user 端接口**必须返回统一格式：`{code: 200, message: "success", data: ...}`

使用 `response.Success()` 包装返回值：
```go
// 正确
return response.Success(data), nil

// 错误 - 不要手动构造 map
return map[string]any{"code": 200, "message": "success", data: ...}, nil
```

**规则**：
- 成功返回：`return response.Success(数据), nil`
- 失败返回：使用 `errorcode.XXX().ToFuegoError()` 或 `fuego.*Error` 类型
- **禁止**手动构造 `map[string]any{"code": ...}` 返回格式
- `response.Success()` 定义在 `pkg/response/response.go`

**所有现有接口已统一**，以后新增的 user 端接口也必须遵守此规范。

## Environment Variables

```env
SERVER_PORT=8080
SERVER_MODE=debug
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=poem
DB_SSLMODE=disable
JWT_SECRET=your-secret-key
JWT_EXPIRE_HOUR=72
```

## Database

使用 PostgreSQL，通过 `pgxpool` 连接池。迁移文件位于 `migrations/` 目录。

---

## @规范文件

规范已拆分到 `rules/` 目录：

- `@error-code.md` - **错误码规范**：精确错误码体系、参数校验、数据库错误处理
- `@code-style.md` - 代码规范：命名约定、错误处理、代码组织
- `@api-style.md` - API 规范：路由设计、Fuego 接口注释规范、模块分类
- `@git-style.md` - Git 规范：Commit Message 格式、分支管理
- `@dev-flow.md` - 开发流程：新功能/Bug 修复流程、测试规范
- `@performance.md` - 性能规范：数据库优化、缓存策略
- `@security.md` - 安全规范：认证、数据校验、敏感信息
- `@documentation.md` - 文档规范：Fuego 接口文档规范、结构体标签、检查清单
