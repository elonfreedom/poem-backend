# 晓诗产品需求文档 · 后端篇（poem-backend）

> **版本**：v1.0 · **日期**：2026-08-29 · **状态**：待评审
> **项目**：poem-backend（Go + Fuego + PostgreSQL）
> **用户**：为 poem-front 和 poem-admin 提供 API 服务

---

## 1. 产品定位

poem-backend 是晓诗的 **后端 API 服务**，基于 Go + Fuego 框架，采用分层架构（Handler → Service → Repository → Database），为前端（poem-front）和后台管理（poem-admin）提供 RESTful API 接口。

### 1.1 设计原则

1. **RESTful**：严格遵循 REST 设计规范，资源导向
2. **类型安全**：Go 强类型 + Fuego 泛型，编译期捕获错误
3. **统一响应**：全局统一 JSON 响应格式 `{ code: 0, message: "ok", data: ... }`
4. **分层解耦**：单向依赖，禁止反向调用
5. **可观测**：结构化日志 + OpenAPI 文档自动生成

### 1.2 技术约束

| 维度 | 约束 |
|------|------|
| 语言 | Go 1.26 |
| HTTP 框架 | Fuego（类型安全路由 + OpenAPI） |
| 数据库 | PostgreSQL + pgx/v5 连接池 |
| 认证 | JWT + WebAuthn/Passkey |
| 简繁转换 | gocc（纯 Go） |
| 拼音 | go-pinyin |
| 部署 | Docker + 阿里云 ECS（上海） |

---

## 2. 架构设计

### 2.1 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    poem-front │ poem-admin                       │
└──────────────────────────────┬──────────────────────────────────┘
                               │ HTTP / JSON
┌──────────────────────────────▼──────────────────────────────────┐
│                        Fuego Router                             │
│              (路由注册 + OpenAPI 文档自动生成)                     │
├─────────────────────────────────────────────────────────────────┤
│                        Middleware                                │
│              (CORS / Auth / Logger / RateLimit)                 │
├───────────────────────────┬─────────────────────────────────────┤
│      Admin Handler        │           User Handler              │
│  (poems/authors/cats/     │  (auth/profile/poems/               │
│   users/tools/banners)    │   plans/checkins/favorites)         │
├───────────────────────────┼─────────────────────────────────────┤
│      Admin Service        │           User Service              │
├───────────────────────────┴─────────────────────────────────────┤
│                        Repository                               │
├─────────────────────────────────────────────────────────────────┤
│                    PostgreSQL (pgx/v5)                          │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 目录结构

```
cmd/server/                  → 服务入口
internal/
├── config/                  → 环境变量配置
├── middleware/              → HTTP 中间件
├── handler/                 → 路由处理
│   ├── admin/               → 管理后台处理器
│   └── user/                → 用户端处理器
├── model/                   → 数据模型
│   ├── admin/               → Admin 请求/响应结构体
│   └── user/                → User 请求/响应结构体
├── repository/              → 数据访问层
├── service/                 → 业务逻辑层
│   ├── admin/               → 管理后台业务逻辑
│   └── user/                → 用户端业务逻辑
└── router/                  → 路由注册
    ├── admin_router.go      → 管理后台路由
    └── user_router.go       → 用户端路由
pkg/
├── database/                → PostgreSQL 连接池
├── response/                → 统一 JSON 响应
├── errorcode/               → 错误码定义
├── convert/                 → 简繁体转换
├── pinyin/                  → 汉字转拼音
└── migrate/                 → 数据库迁移
```

### 2.3 分层职责

| 层级 | 目录 | 职责 | 禁止 |
|------|------|------|------|
| Router | `router/` | 注册路由、配置中间件 | 业务逻辑 |
| Handler | `handler/` | 解析请求、参数校验、调用 Service | 数据库操作 |
| Service | `service/` | 业务逻辑、事务协调、数据转换 | 直接操作 HTTP |
| Repository | `repository/` | 数据库 CRUD、SQL 查询 | 业务逻辑 |
| Model | `model/` | 数据结构定义 | - |

---

## 3. API 设计

### 3.1 路由前缀

| 模块 | 前缀 | 认证 | 说明 |
|------|------|------|------|
| 公开接口 | `/api/public/` | 无需 | 登录、注册、健康检查、公开浏览 |
| 用户端 | `/api/user/` | JWT | 普通用户使用 |
| 管理后台 | `/api/admin/` | JWT + Admin | 管理员使用 |

### 3.2 HTTP 方法语义

| 方法 | 用途 | 幂等 | 安全 |
|------|------|------|------|
| `GET` | 简单查询（参数在 URL） | ✅ | ✅ |
| `POST` | 创建资源 / 触发操作 / 复杂查询 | ❌ | ❌ |
| `PUT` | 全量替换 | ✅ | ❌ |
| `PATCH` | 部分更新 | ❌ | ❌ |
| `DELETE` | 删除资源 | ✅ | ❌ |

### 3.3 完整路由清单

#### 公开接口 `/api/public/*`

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/passkey/register/begin` | 开始 Passkey 注册 |
| `POST` | `/passkey/register/finish` | 完成 Passkey 注册 |
| `POST` | `/passkey/login/begin` | 开始 Passkey 登录 |
| `POST` | `/passkey/login/finish` | 完成 Passkey 登录 |
| `GET` | `/poems/daily` | 每日推荐诗歌 |
| `GET` | `/poems/search` | 诗歌搜索（v1.1） |
| `GET` | `/poems/search/hot` | 热门搜索词（v1.1） |
| `GET` | `/shared-plans` | 浏览共享计划库 |
| `GET` | `/shared-plans/{id}` | 共享计划详情 |
| `POST` | `/passkeys/add/connect` | 新设备连接 |
| `GET` | `/passkeys/add/status` | 查询连接状态 |
| `POST` | `/passkeys/add/finish` | 完成设备注册 |
| `POST` | `/passkeys/add/reject` | 放弃绑定 |

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
| `GET` | `/poems/daily` | 每日推荐（登录后） |
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
| `GET` | `/shared-plans` | 浏览共享库 |
| `GET` | `/shared-plans/{id}` | 计划详情 |
| `POST` | `/shared-plans` | 创建共享计划 |
| `GET` | `/shared-plans/mine` | 我的计划 |
| `PUT` | `/shared-plans/{id}` | 更新计划 |
| `PUT` | `/shared-plans/{id}/publish` | 发布计划 |
| `PUT` | `/shared-plans/{id}/unpublish` | 取消发布 |
| `DELETE` | `/shared-plans/{id}` | 删除计划 |
| `POST` | `/shared-plans/{id}/subscribe` | 订阅计划 |
| `DELETE` | `/shared-plans/{id}/subscribe` | 取消订阅 |
| **订阅管理** | | |
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

#### 管理后台 `/api/admin/*`

| 方法 | 路径 | 说明 |
|------|------|------|
| **认证** | | |
| `POST` | `/auth/login` | 管理员登录（公开） |
| `GET` | `/user/info` | 获取管理员信息 |
| `GET` | `/auth/codes` | 获取权限码 |
| `POST` | `/auth/logout` | 退出登录 |
| **诗歌管理** | | |
| `GET` | `/poems` | 获取诗歌列表 |
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

---

## 4. 请求/响应规范

### 4.1 统一响应结构

**成功响应**：
```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

**分页响应**：
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

**错误响应**：
```json
{
  "code": 1001,
  "message": "missing parameter",
  "error": "title 不能为空",
  "data": null
}
```

### 4.2 分页参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码（≥1） |
| `page_size` | int | 10 | 每页数量（1-50） |

### 4.3 错误码体系

| 码段 | 类别 | HTTP 状态码 |
|------|------|-------------|
| 1xxx | 参数错误 | 400 |
| 2xxx | 认证授权 | 401/403 |
| 3xxx | 资源不存在 | 404 |
| 4xxx | 业务冲突 | 400 |
| 5xxx | 验证失败 | 422 |
| 9xxx | 服务器错误 | 500 |

---

## 5. 核心业务逻辑

### 5.1 诗歌管理

#### 状态机
```
draft → published → archived
  ↑         ↓
  └─────────┘ (编辑)
```

#### 业务规则

| 规则 | 说明 |
|------|------|
| 标题+作者+正文首句唯一 | 导入时自动去重 |
| 默认状态为 draft | 创建时未指定状态自动设为草稿 |
| 拼音可手动校正 | 自动生成后，admin 可编辑覆盖 |
| 简繁体双向转换 | 填一端自动生成另一端 |
| 作者不参与拼音 | 拼音仅针对标题和正文 |

### 5.2 拼音与简繁体

| 字段 | 是否生成 | 说明 |
|------|----------|------|
| `title_pinyin` | ✅ | 标题拼音（带声调） |
| `content_pinyin` | ✅ | 正文拼音（带声调，保留换行） |
| `author_pinyin` | ❌ | 作者不参与拼音 |
| `*_sc` | ✅ | 简体字段（从繁体转换） |

**转换优先级**：
1. 用户同时提供繁体和简体 → 以用户输入为准
2. 仅提供繁体 → 自动生成简体
3. 仅提供简体 → 保留原样

### 5.3 收藏管理

| 规则 | 说明 |
|------|------|
| 同一诗歌不可重复收藏 | 重复调用返回成功（幂等） |
| 收藏不存在的诗歌 | 返回 404 |
| 取消不存在的收藏 | 返回成功（幂等） |
| 诗歌已被删除 | 收藏记录保留，显示时标记「已下架」 |

### 5.4 阅读计划

| 规则 | 说明 |
|------|------|
| 每日阅读量 | 1-50 篇 |
| 同时只能有一个活跃计划 | 创建新计划需先完成或取消当前计划 |
| 完成条件 | 连续打卡达到计划天数 |

### 5.5 打卡系统

| 规则 | 说明 |
|------|------|
| 每日只能打卡一次 | 重复打卡返回成功（幂等） |
| 打卡需关联诗文 | 每次打卡关联一首诗文（poem_id） |
| 连续打卡中断 | 连续天数归零，计划进度保留 |
| 不支持补卡 | 跨天以服务器时间为准 |

#### 5.5.1 打卡热力图 API

> **设计约束**：每日限 1 篇打卡，热力图为**二色阶**（已打卡/未打卡），hover 显示日期 + 诗文标题。

| 接口 | 方法 | 说明 | 请求参数 | 响应字段 |
|------|------|------|----------|----------|
| `/checkins` | GET | 获取打卡记录（日期范围） | `start_date`, `end_date` | `date`, `consecutive_day`, `poem_id`, `poem_title` |
| `/checkins/stats` | GET | 打卡统计 | - | `total_days`, `consecutive_day`, `max_consecutive`, `last_check_in` |
| `/checkins/calendar` | GET | 月度打卡日历 | `year`, `month` | `days[].{day, is_checked}` |

**热力图数据流**：
```
前端请求：GET /checkins?start_date=2025-08-29&end_date=2026-08-29
后端处理：JOIN poems 表获取 poem_title
响应格式（统一响应格式）：
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "date": "2026-08-29",
        "consecutive_day": 30,
        "poem_id": 42,
        "poem_title": "静夜思"
      }
    ],
    "total": 180
  }
}
```

### 5.6 接口限流

| 维度 | 规则 |
|------|------|
| IP 级限流 | 100 req/min/IP（已有） |
| 用户级限流 | 登录用户 30 req/min/user（新增） |
| 注册/登录 | 5 req/min/IP（防暴力破解） |
| 打卡接口 | 10 req/min/user（防刷打卡） |
| 超限响应 | 429 Too Many Requests + `Retry-After` 头 |

### 5.7 批量工具

| 工具 | 逻辑 | 返回 |
|------|------|------|
| 生成拼音 | 扫描 title_pinyin 为空的记录，自动生成 | 处理数量 |
| 生成简体 | 扫描简体字段为空的记录，从繁体转换 | 处理数量 |
| 提取作者 | 从 poems 表 author 字段去重插入 authors 表 | 提取数量 |

### 5.8 Piper TTS 朗读服务

| 项目 | 说明 |
|------|------|
| **服务** | Piper TTS（自部署，Docker 容器） |
| **接口** | `GET /api/public/tts?text={诗文内容}&speed=1.0` |
| **输入** | 纯文本（后端自动清洗标点、处理多音字） |
| **输出** | OGG 音频流（`Content-Type: audio/ogg`） |
| **缓存** | 基于 text hash 缓存，24h TTL |
| **限流** | 20 req/min/user（生成成本较高） |
| **降级** | Piper 不可用时返回 503，前端降级 Web Speech API |
| **部署** | 与 poem-backend 同服务器，通过 Unix Socket 通信 |

**数据流**：
```
前端请求：GET /public/tts?text=床前明月光...
  → 检查缓存（Redis/SQLite）
  → 缓存命中：直接返回音频
  → 缓存未命中：调用 Piper → 缓存 → 返回音频
```

### 5.9 诗文状态过滤

| 规则 | 说明 |
|------|------|
| 用户端搜索 | 仅返回 `status='published'` 的诗文 |
| 推荐接口 | 仅返回 `published` 诗文 |
| 诗林浏览 | 仅返回 `published` 诗文 |
| 每日推荐 | 仅返回 `published` 诗文 |
| 收藏列表 | 仅返回 `published` 诗文（已下架的标记为「已下架」） |

---

## 6. 认证架构

### 6.1 JWT 认证流程

```
Client → [Login] → Server 发放 JWT（有效期 72h）
Client → [Request + Bearer Token] → Middleware 验证 → 注入 Context → Handler
```

### 6.2 角色控制

| 角色 | 访问范围 |
|------|----------|
| `user` | `/api/user/*` |
| `admin` | `/api/admin/*` + `/api/user/*` |

### 6.3 Passkey 认证流程（零输入方案）

```
注册（前端自动检测设备名，用户零输入）：
  POST register/begin { device_name: "Chrome on Mac" } ← 前端自动填充
    → 返回 options → 浏览器 WebAuthn → POST register/finish → 返回 JWT

登录（用户零输入）：
  POST login/begin → 返回 options → 浏览器 WebAuthn → POST login/finish → 返回 JWT

⚠️ 注意：device_name 由前端从 navigator.userAgent 自动检测，后端不做非空校验
```

### 6.4 跨设备添加 Passkey

```
设备 A：POST add/begin → 获得 connection_token → 生成 QR 码
设备 B：POST add/connect（token）→ 等待确认
设备 A：确认连接 → POST add/confirm
设备 B：浏览器 WebAuthn → POST add/finish → 完成绑定
```

---

## 7. 数据库设计

### 7.1 核心表

| 表名 | 说明 |
|------|------|
| `users` | 用户表 |
| `passkeys` | Passkey 凭证表 |
| `poems` | 诗歌表 |
| `authors` | 作者表 |
| `categories` | 分类表 |
| `tags` | 标签表 |
| `poem_tags` | 诗歌-标签关联表 |
| `favorites` | 收藏表 |
| `reading_plans` | 阅读计划表 |
| `plan_poems` | 计划-诗歌关联表 |
| `shared_plans` | 共享计划表 |
| `subscriptions` | 订阅表 |
| `checkins` | 打卡记录表（含 poem_id 关联诗文） |
| `banners` | Banner 表 |
| `announcements` | 公告表 |
| `configs` | 系统配置表 |

### 7.2 索引优化

```sql
-- 诗歌搜索索引（v1.1）
CREATE INDEX idx_poems_title ON poems USING gin (title gin_trgm_ops);
CREATE INDEX idx_poems_title_sc ON poems USING gin (title_sc gin_trgm_ops);
CREATE INDEX idx_poems_author ON poems USING gin (author gin_trgm_ops);
CREATE INDEX idx_poems_author_sc ON poems USING gin (author_sc gin_trgm_ops);

-- 常用查询索引
CREATE INDEX idx_checkins_user_date ON checkins(user_id, date);
CREATE INDEX idx_checkins_poem_id ON checkins(poem_id);

-- checkins 表结构（含 poem_id）
CREATE TABLE checkins (
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date                DATE NOT NULL,
    consecutive_day     INT NOT NULL DEFAULT 1,
    poem_id             BIGINT REFERENCES poems(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, date)
);
CREATE INDEX idx_favorites_user ON favorites(user_id);
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id, status);
CREATE INDEX idx_poems_status ON poems(status);
```

---

## 8. 非功能需求

### 8.1 性能要求

| 指标 | 目标 |
|------|------|
| API 响应时间（P95） | < 200ms |
| 搜索响应时间 | < 500ms（1 万条数据内） |
| 数据库连接池 | 最大 20 连接 |
| 并发用户数（MVP） | 1000 |

### 8.2 安全要求

| 维度 | 措施 |
|------|------|
| 认证 | JWT + Passkey 双体系 |
| 密码 | bcrypt 加密（管理员） |
| SQL 注入 | 参数化查询（pgx） |
| XSS | 输入校验 + 输出编码 |
| CSRF | CORS 白名单 |
| 限流 | IP 级 100 req/min + 用户级 30 req/min + 关键接口独立限流 |
| 日志 | 结构化日志，敏感信息脱敏 |

### 8.3 可观测性

| 维度 | 方案 |
|------|------|
| 日志 | 结构化 JSON 日志（请求 ID 追踪） |
| 文档 | Swagger UI 自动生成（/swagger） |
| 健康检查 | GET /health |
| 指标 | Prometheus metrics（预留） |

---

## 9. 部署架构

### 9.1 部署流程

```
本地构建 → Docker 镜像（linux/amd64）→ 上传 OSS → 服务器加载 → 容器运行
```

### 9.2 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_PORT` | 服务端口 | 8080 |
| `SERVER_MODE` | 运行模式 | debug |
| `DB_HOST` | 数据库主机 | localhost |
| `DB_PORT` | 数据库端口 | 5432 |
| `DB_USER` | 数据库用户 | postgres |
| `DB_PASSWORD` | 数据库密码 | - |
| `DB_NAME` | 数据库名 | poem |
| `JWT_SECRET` | JWT 密钥 | - |
| `JWT_EXPIRE_HOUR` | Token 有效期 | 72 |

### 9.3 Docker 配置

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

---

## 10. 验收标准

### 10.1 功能验收

- [ ] 全部 API 接口通过 Postman/curl 测试
- [ ] Passkey 注册/登录全流程通过
- [ ] JWT 认证 + 角色权限控制正确
- [ ] 诗歌 CRUD + 批量导入正确
- [ ] 拼音/简繁自动生成正确
- [ ] 收藏/计划/打卡业务逻辑正确
- [ ] 打卡记录返回 poem_title（JOIN poems 表）
- [ ] 热力图 API 支持日期范围查询（start_date, end_date）
- [ ] 用户管理 + 状态变更正确
- [ ] 数据统计接口返回准确
- [ ] 批量工具执行正确
- [ ] 接口限流生效（用户级/IP级）
- [ ] 用户端搜索/推荐仅返回 published 诗文
- [ ] 用户信息编辑接口正常（昵称/头像）
- [ ] Piper TTS 接口正常（生成/缓存/流式返回）
- [ ] TTS 限流生效（20 req/min/user）

### 10.2 性能验收

- [ ] API P95 响应时间 < 200ms
- [ ] 100 并发请求无错误
- [ ] 数据库连接池无泄漏

### 10.3 安全验收

- [ ] SQL 注入测试通过
- [ ] XSS 测试通过
- [ ] 未授权访问返回 401/403
- [ ] 敏感信息不泄露

### 10.4 文档验收

- [ ] Swagger 文档完整（所有接口 + 请求/响应示例）
- [ ] README 部署文档更新
- [ ] 数据库迁移脚本完整
