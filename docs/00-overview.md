# 项目概述

## 产品定位

诗歌应用后端服务，为用户提供古诗词浏览、收藏、每日阅读计划和打卡功能，同时提供 Admin 后台管理系统。

## 目标用户

- **普通用户**：喜爱古诗词的阅读者
- **管理员**：内容运营人员

## 核心价值

- 优质古诗词内容聚合与呈现
- 个性化阅读计划与打卡激励
- 便捷的内容管理后台

## 技术栈

- **语言**: Go 1.26
- **框架**: Fuego
- **数据库**: PostgreSQL
- **认证**: JWT + 微信/Apple 登录

## 模块列表

| 模块 | 说明 | 文档 |
|-----|------|------|
| 用户模块 | 登录注册、个人信息、角色权限 | [01-auth.md](01-auth.md) |
| 诗歌模块 | Admin CRUD + User 浏览搜索 | [02-poem.md](02-poem.md) |
| 收藏模块 | 收藏/取消收藏、收藏列表 | [03-favorite.md](03-favorite.md) |
| 阅读计划 | 每日阅读目标、进度跟踪 | [04-reading-plan.md](04-reading-plan.md) |
| 打卡系统 | 每日打卡、统计、排行榜 | [05-checkin.md](05-checkin.md) |
| 数据统计 | 用户增长、浏览量、活跃度 | [06-stats.md](06-stats.md) |
| 系统配置 | Banner、公告、系统参数 | [07-system.md](07-system.md) |
| 数据库设计 | 表结构、索引 | [08-database.md](08-database.md) |
| API 汇总 | 接口列表 | [09-api.md](09-api.md) |
| 非功能需求 | 性能、安全、可用性 | [10-non-functional.md](10-non-functional.md) |
