# 诗歌管理

## 功能概述

诗歌管理是后台核心模块，提供诗歌、分类、标签的 CRUD 功能，支持内容的草稿、发布、归档状态流转。

## 功能详情

### 1. 诗歌管理

#### 诗歌列表

**功能说明**：分页展示所有诗歌，支持多条件筛选。

**筛选条件**：
- 关键词（标题、作者）
- 分类
- 状态（草稿/发布/归档）
- 创建时间范围

**列表信息**：ID、标题、作者、分类、状态、创建人、创建时间

#### 创建诗歌

**字段**：
| 字段 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| title | string | 是 | 标题 |
| author | string | 是 | 作者 |
| dynasty | string | 否 | 朝代 |
| content | string | 是 | 诗歌内容 |
| translation | string | 否 | 翻译 |
| appreciation | string | 否 | 赏析 |
| category_id | int | 否 | 分类 ID |
| tags | []string | 否 | 标签列表 |
| cover_url | string | 否 | 封面图 |
| status | string | 是 | 初始状态 |

#### 编辑诗歌

**业务规则**：
- 可修改所有字段
- 可修改状态（草稿↔发布↔归档）

#### 删除诗歌

**业务规则**：
- 软删除，标记删除状态
- 已发布诗歌删除后用户端不可见

### 2. 分类管理

**功能说明**：管理诗歌分类（如唐诗、宋词、元曲）。

**业务规则**：
- 分类名称唯一
- 可设置排序值（sort）
- 删除分类前需确认无诗歌关联

### 3. 标签管理

**功能说明**：管理诗歌标签（如思乡、爱国、爱情）。

**业务规则**：
- 标签名称唯一
- 可关联多个诗歌
- 删除标签不影响诗歌

## 状态流转

```
草稿(draft) → 发布(published) → 归档(archived)
    ↑              ↓
    └──────────────┘
```

| 状态 | 说明 | 用户端可见 |
|-----|------|-----------|
| draft | 草稿 | 否 |
| published | 已发布 | 是 |
| archived | 已归档 | 否 |

## 数据模型

```go
// 诗歌管理 - 创建请求
type CreatePoemRequest struct {
    Title        string   `json:"title" validate:"required,max=100"`
    Author       string   `json:"author" validate:"required,max=50"`
    Dynasty      string   `json:"dynasty" validate:"omitempty,max=20"`
    Content      string   `json:"content" validate:"required"`
    Translation  string   `json:"translation"`
    Appreciation string   `json:"appreciation"`
    CategoryID   int64    `json:"category_id"`
    Tags         []string `json:"tags"`
    CoverURL     string   `json:"cover_url" validate:"omitempty,url"`
    Status       string   `json:"status" validate:"required,oneof=draft published archived"`
}

// 诗歌管理 - 更新请求
type UpdatePoemRequest struct {
    Title        string   `json:"title" validate:"omitempty,max=100"`
    Author       string   `json:"author" validate:"omitempty,max=50"`
    Dynasty      string   `json:"dynasty" validate:"omitempty,max=20"`
    Content      string   `json:"content"`
    Translation  string   `json:"translation"`
    Appreciation string   `json:"appreciation"`
    CategoryID   int64    `json:"category_id"`
    Tags         []string `json:"tags"`
    CoverURL     string   `json:"cover_url" validate:"omitempty,url"`
}

// 诗歌管理 - 状态更新请求
type UpdatePoemStatusRequest struct {
    Status string `json:"status" validate:"required,oneof=draft published archived"`
}

// 诗歌管理 - 列表请求
type PoemListRequest struct {
    Page       int    `json:"page" validate:"min=1"`
    PageSize   int    `json:"page_size" validate:"min=1,max=100"`
    Keyword    string `json:"keyword"`
    CategoryID int64  `json:"category_id"`
    Status     string `json:"status" validate:"omitempty,oneof=draft published archived"`
}

// 分类 - 创建请求
type CreateCategoryRequest struct {
    Name string `json:"name" validate:"required,max=50"`
    Sort int    `json:"sort" validate:"min=0"`
}

// 分类 - 更新请求
type UpdateCategoryRequest struct {
    Name string `json:"name" validate:"omitempty,max=50"`
    Sort int    `json:"sort" validate:"min=0"`
}

// 标签 - 创建请求
type CreateTagRequest struct {
    Name string `json:"name" validate:"required,max=50"`
}
```

## API 接口

### 诗歌

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/poems | 诗歌列表（分页、筛选） |
| POST | /api/admin/poems | 创建诗歌 |
| GET | /api/admin/poems/:id | 诗歌详情 |
| PUT | /api/admin/poems/:id | 更新诗歌 |
| DELETE | /api/admin/poems/:id | 删除诗歌 |
| PUT | /api/admin/poems/:id/status | 更新状态 |

### 分类

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/categories | 分类列表 |
| POST | /api/admin/categories | 创建分类 |
| PUT | /api/admin/categories/:id | 更新分类 |
| DELETE | /api/admin/categories/:id | 删除分类 |

### 标签

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/tags | 标签列表 |
| POST | /api/admin/tags | 创建标签 |
| DELETE | /api/admin/tags/:id | 删除标签 |

## 异常处理

| 场景 | 错误码 | 提示信息 |
|-----|--------|---------|
| 无权限 | 403 | 无权限操作 |
| 诗歌不存在 | 404 | 诗歌不存在 |
| 分类名称已存在 | 400 | 分类名称已存在 |
| 分类下有诗歌 | 400 | 该分类下有诗歌，无法删除 |
| 标签名称已存在 | 400 | 标签名称已存在 |
