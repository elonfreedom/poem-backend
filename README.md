  # Poem Backend

诗歌应用后端服务，基于 Go + Fuego 框架。

## 📋 需求文档

| 文档 | 说明 |
|-----|------|
| [00-overview.md](docs/00-overview.md) | 项目概述 |
| [01-auth.md](docs/01-auth.md) | 用户模块 |
| [02-poem.md](docs/02-poem.md) | 诗歌模块 |
| [03-favorite.md](docs/03-favorite.md) | 收藏模块 |
| [04-reading-plan.md](docs/04-reading-plan.md) | 阅读计划 |
| [05-checkin.md](docs/05-checkin.md) | 打卡系统 |
| [06-stats.md](docs/06-stats.md) | 数据统计 |
| [07-system.md](docs/07-system.md) | 系统配置 |
| [08-database.md](docs/08-database.md) | 数据库设计 |
| [09-api.md](docs/09-api.md) | API 汇总 |
| [10-non-functional.md](docs/10-non-functional.md) | 非功能需求 |

## 功能特性

### Admin 后台管理系统
- 内容管理（诗歌 CRUD）
- 用户管理
- 数据统计
- 系统配置

### User 用户端
- 诗歌浏览、搜索
- 收藏功能
- 用户系统（登录、注册、个人中心）
- 每日阅读计划
- 每日阅读打卡
- 打卡记录

## 技术栈

- **语言**: Go 1.26
- **框架**: Fuego
- **数据库**: PostgreSQL
- **认证**: JWT + 微信/Apple 登录

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件配置数据库等信息
```

### 启动服务

```bash
go run cmd/server/main.go
```

服务默认运行在 `http://localhost:8080`

### API 文档

启动服务后访问：`http://localhost:8080/swagger`

## 项目结构

```
poem-backend/
├── cmd/server/          # 服务入口
├── internal/
│   ├── config/          # 配置管理
│   ├── middleware/       # 中间件
│   ├── handler/         # 路由处理
│   │   ├── admin/       # Admin 路由
│   │   └── user/        # User 路由
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑层
│   └── router/          # 路由注册
├── pkg/
│   ├── database/        # 数据库连接
│   └── response/        # 统一响应格式
├── migrations/          # 数据库迁移
├── rules/               # 开发规范
└── docs/                # 需求文档
    └── PRD.md           # 产品需求文档
```

## 开发计划

- [x] 项目基础架构（Config、中间件、Model）
- [ ] 用户认证模块（登录/注册）
- [ ] 诗歌管理模块（Admin CRUD + User 浏览）
- [ ] 收藏功能
- [ ] 阅读计划
- [ ] 打卡系统
- [ ] 数据统计
- [ ] 系统配置（Banner/公告）
