# Poem Backend

诗歌应用后端服务，基于 Go + Fuego 框架。

## 需求文档

### 用户端（C端）

| 文档 | 说明 |
|-----|------|
| [user/00-overview.md](docs/user/00-overview.md) | 用户端产品概述 |
| [user/01-auth.md](docs/user/01-auth.md) | 用户系统（登录注册、个人中心） |
| [user/02-poem.md](docs/user/02-poem.md) | 诗歌浏览（列表、详情、搜索、每日推荐） |
| [user/03-favorite.md](docs/user/03-favorite.md) | 收藏模块 |
| [user/04-reading-plan.md](docs/user/04-reading-plan.md) | 阅读计划 |
| [user/05-checkin.md](docs/user/05-checkin.md) | 打卡系统 |
| [user/06-api.md](docs/user/06-api.md) | 用户端接口汇总 |

### 后台管理（Admin）

| 文档 | 说明 |
|-----|------|
| [admin/00-overview.md](docs/admin/00-overview.md) | 后台管理产品概述 |
| [admin/01-poem-manage.md](docs/admin/01-poem-manage.md) | 诗歌管理（诗歌/分类/标签） |
| [admin/02-stats.md](docs/admin/02-stats.md) | 数据统计 |
| [admin/03-system.md](docs/admin/03-system.md) | 系统配置（Banner/公告/参数） |
| [admin/04-api.md](docs/admin/04-api.md) | 管理端接口汇总 |

### 共用文档

| 文档 | 说明 |
|-----|------|
| [shared/database.md](docs/shared/database.md) | 数据库设计 |
| [shared/non-functional.md](docs/shared/non-functional.md) | 非功能需求 |

## 功能特性

### 用户端（C端）
- 诗歌浏览、搜索、每日推荐
- 收藏功能
- 用户系统（登录、注册、个人中心）
- 每日阅读计划
- 每日阅读打卡
- 打卡记录与排行榜

### 后台管理（Admin）
- 内容管理（诗歌/分类/标签 CRUD）
- 数据统计（用户增长、浏览量、活跃度）
- 系统配置（Banner、公告、参数）

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
    ├── user/            # 用户端文档
    ├── admin/           # 后台管理文档
    └── shared/          # 共用文档
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
