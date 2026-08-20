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

**响应格式**: 统一使用 `pkg/response` 的 `Success[T]` 和 `Error` 函数，返回 `{code, message, data}` 结构。

**认证**: JWT token 通过 `Authorization: Bearer <token>` 传递。Admin 路由需 `role=admin`。用户信息存储在 context 中，通过 `middleware.GetUserIDFromContext(ctx)` 获取。

**Model 设计**: 每个实体有 Model（数据库字段）、Response（API 输出，隐藏敏感字段）、Request（入参验证）三类结构体。通过 `ToResponse()` 方法转换。

**Fuego 框架**: 路由定义使用 `fuego` 的类型安全方式，handler 签名包含 `*fuego.Context[T]` 泛型参数。

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

- `@code-style.md` - 代码规范：命名约定、错误处理、代码组织
- `@api-style.md` - API 规范：路由设计、请求/响应格式、错误码
- `@git-style.md` - Git 规范：Commit Message 格式、分支管理
- `@dev-flow.md` - 开发流程：新功能/Bug 修复流程、测试规范
- `@performance.md` - 性能规范：数据库优化、缓存策略
- `@security.md` - 安全规范：认证、数据校验、敏感信息
- `@documentation.md` - 文档规范：代码注释、API 文档
