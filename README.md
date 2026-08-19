# Poem Backend

诗歌应用后端服务，基于 Go + Fuego 框架。

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
└── migrations/          # 数据库迁移
```

## 开发计划

- [ ] 项目基础架构
- [ ] 用户认证模块
- [ ] 诗歌管理模块
- [ ] 收藏功能
- [ ] 阅读计划
- [ ] 打卡系统
- [ ] 数据统计
