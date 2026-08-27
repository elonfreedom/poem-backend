# 文档规范

## 代码注释
- 导出的函数/类型必须有 godoc 注释
- 复杂逻辑添加行内注释
- TODO 使用 `// TODO: 说明` 格式

## API 文档
- 使用 Fuego 内置 OpenAPI 3.1 支持
- 文档由代码自动生成，**不要手写 Swagger 注解**
- 每个接口通过 Fuego Option 配置文档

---

# Fuego 接口文档规范

## 1. 核心原则

Fuego 框架通过**代码即文档**的方式自动生成 OpenAPI 规范：
- Handler 签名 → 决定入参/出参类型
- 结构体标签 → 决定字段说明
- Option 配置 → 决定接口摘要/描述/分类

**禁止**：
- ❌ 手写 Swagger 注解（如 `@Summary`、`@Param`）
- ❌ 使用 `any` 作为请求/响应类型
- ❌ 使用 `OptionDescription`（会被覆盖）

## 2. 结构体文档规范

### 2.1 请求结构体
```go
// CreatePlanRequest 创建计划请求（结构体注释必须有）
type CreatePlanRequest struct {
    // 每个字段必须有 description 标签
    DailyCount int `json:"daily_count" validate:"required,min=1,max=50" description:"每日阅读数量（1-50）"`
    Duration   int `json:"duration" validate:"required,oneof=7 14 30 90" description:"计划天数（7/14/30/90）"`
}
```

### 2.2 响应结构体
```go
// CreatePlanResponse 创建计划响应（结构体注释必须有）
type CreatePlanResponse struct {
    PlanID     int       `json:"plan_id" description:"计划ID"`
    DailyCount int       `json:"daily_count" description:"每日阅读数量"`
    StartDate  time.Time `json:"start_date" description:"开始日期"`
    EndDate    time.Time `json:"end_date" description:"结束日期"`
    Status     string    `json:"status" description:"状态"`
}
```

### 2.3 字段标签规范
| 标签 | 必填 | 说明 |
|------|------|------|
| `json` | ✅ | JSON 字段名，snake_case |
| `validate` | 可选 | 验证规则 |
| `description` | ✅ | 字段说明，中文 |

## 3. Handler 文档规范

### 3.1 注释格式
```go
// MethodName 中文说明（必须有）
func (h *Handler) MethodName(c fuego.ContextWithBody[RequestType]) (ResponseType, error)
```

### 3.2 命名约定
- Handler 方法名：PascalCase
- 方法注释：`// 方法名 中文说明`，如 `// List 获取诗歌列表`

## 4. 路由文档规范

每条路由**必须**配置三个 Option：

```go
fuego.Post(group, "/path", handler.Method,
    fuego.OptionSummary("简短摘要"),           // 2-6 字
    fuego.OptionOverrideDescription("详细描述"), // 不会被覆盖
    fuego.OptionTags("模块分类"),               // 用于分组
)
```

### 4.1 OptionSummary
- 长度：2-6 个汉字
- 格式：动词 + 名词
- 示例：`"获取诗歌列表"`、`"创建阅读计划"`、`"用户登录"`

### 4.2 OptionOverrideDescription
- 长度：10-50 个汉字
- 说明接口功能和行为
- 示例：`"分页获取诗歌列表，支持按分类筛选"`

### 4.3 OptionTags
- 用于模块分组
- 可选值见下方表格

## 5. 模块分类

| 标签 | 说明 | 路由前缀 |
|------|------|----------|
| `Passkey 认证` | 注册、登录 | `/api/public/passkey/*` |
| `个人信息` | 用户资料、Passkey 管理 | `/api/user/profile`, `/api/user/passkeys/*` |
| `诗歌浏览` | 诗歌列表、详情、搜索、推荐 | `/api/user/poems/*` |
| `收藏管理` | 收藏添加、取消、列表 | `/api/user/favorites/*` |
| `阅读计划` | 计划创建、暂停、恢复、记录 | `/api/user/reading-plans/*` |
| `打卡系统` | 打卡、记录、统计、日历、排行榜 | `/api/user/checkins/*` |

## 6. 验证方式

启动服务后，访问以下地址验证文档：
- Swagger UI: http://localhost:8080/swagger
- OpenAPI JSON: http://localhost:8080/swagger/openapi.json

### 6.1 检查项
- [ ] 接口摘要显示中文
- [ ] 接口描述正确（无 controller 信息）
- [ ] 入参结构体字段有说明
- [ ] 出参结构体字段有说明
- [ ] 接口按模块分组
- [ ] 错误码 400/401/404/500 都有定义
- [ ] 错误响应包含精确字段名（参数缺失/格式错误时标识具体字段）
- [ ] 数据库查不到数据时返回 404（非笼统 400）
- [ ] 数据库操作失败时返回 500 + 具体错误信息

## 7. 常见错误

| 错误 | 原因 | 修复 |
|------|------|------|
| 入参不显示 | 使用 `ContextNoBody` 或返回 `any` | 改用 `ContextWithBody[具体类型]` |
| 出参显示 `unknown-interface` | 返回类型为 `any` | 改用具体响应类型 |
| 描述被追加 controller 信息 | 使用 `OptionDescription` | 改用 `OptionOverrideDescription` |
| 字段无说明 | 缺少 `description` 标签 | 添加 `description:"说明"` |
