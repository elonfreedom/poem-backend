# 诗歌浏览

## 功能概述

诗歌浏览是用户端核心功能，提供诗歌列表、详情、搜索和每日推荐，帮助用户发现和阅读古诗词。

## 用户故事

- 作为用户，我想浏览诗歌列表，发现感兴趣的内容
- 作为用户，我想按分类筛选诗歌，找到特定类型的作品
- 作为用户，我想查看诗歌详情，阅读翻译和赏析
- 作为用户，我想搜索诗歌，通过标题、作者或内容查找
- 作为用户，我想看每日推荐，每天发现一首新诗

## 功能详情

### 1. 诗歌列表

**功能说明**：分页展示已发布的诗歌，支持按分类筛选。

**业务规则**：
- 仅展示状态为 `published` 的诗歌
- 默认按创建时间倒序排列
- 支持按分类筛选
- 分页参数：page（默认1）、page_size（默认10，最大50）

**列表项信息**：标题、作者、朝代、分类、封面图

### 2. 诗歌详情

**功能说明**：展示诗歌完整内容，包括原文、翻译、赏析。

**业务规则**：
- 浏览详情时记录浏览量（poem_views 表）
- 显示是否已收藏状态
- 显示该诗歌的标签

**详情信息**：标题、作者、朝代、原文、翻译、赏析、分类、标签、收藏状态

### 3. 搜索

**功能说明**：通过关键词搜索诗歌。

**业务规则**：
- 搜索范围：标题、作者、内容
- 使用 PostgreSQL 全文搜索
- 关键词最少 1 个字符
- 无结果时展示友好提示

### 4. 每日推荐

**功能说明**：每天推荐一首诗歌给用户。

**业务规则** |
- 每日固定一首推荐诗（可后台配置或算法选取）
- 同一用户当天看到的推荐相同
- 推荐诗从已发布诗歌中选取

### 5. 分类列表

**功能说明**：展示所有诗歌分类。

**业务规则**：
- 按 sort 字段升序排列
- 仅展示有诗歌关联的分类

## 数据模型

```go
// Poem 诗歌模型
type Poem struct {
    ID           int64     // 自增 ID
    Title        string
    Author       string
    Dynasty      string    // 朝代
    Content      string    // 诗歌内容
    Translation  string    // 翻译
    Appreciation string    // 赏析
    CategoryID   int64     // 自增 ID
    Tags         []string
    CoverURL     string
    Status       string    // draft, published, archived
    CreatedBy    string    // UUID v7
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// PoemResponse 诗歌响应
type PoemResponse struct {
    ID           int64    `json:"id"`
    Title        string   `json:"title"`
    Author       string   `json:"author"`
    Dynasty      string   `json:"dynasty"`
    Content      string   `json:"content"`
    Translation  string   `json:"translation"`
    Appreciation string   `json:"appreciation"`
    Category     string   `json:"category"`
    Tags         []string `json:"tags"`
    CoverURL     string   `json:"cover_url"`
    IsFavorited  bool     `json:"is_favorited"`
}

// PoemListItem 诗歌列表项
type PoemListItem struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    Author    string `json:"author"`
    Dynasty   string `json:"dynasty"`
    Category  string `json:"category"`
    CoverURL  string `json:"cover_url"`
}

// Category 分类
type Category struct {
    ID        int64     // 自增 ID
    Name      string
    Sort      int
    CreatedAt time.Time
    UpdatedAt time.Time
}

// CategoryResponse 分类响应
type CategoryResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
    Sort int    `json:"sort"`
}

// Tag 标签
type Tag struct {
    ID        int64     // 自增 ID
    Name      string
    CreatedAt time.Time
}

// PoemListResponse 诗歌列表响应
type PoemListResponse struct {
    Total int             `json:"total"`
    List  []PoemListItem  `json:"list"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
    Keyword  string `json:"keyword" validate:"required,min=1"`
    Page     int    `json:"page" validate:"min=1"`
    PageSize int    `json:"page_size" validate:"min=1,max=50"`
}
```

## API 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/user/poems | 诗歌列表（支持分类筛选、分页） |
| GET | /api/user/poems/:id | 诗歌详情 |
| GET | /api/user/poems/search | 搜索诗歌 |
| GET | /api/user/poems/daily | 每日推荐 |
| GET | /api/user/categories | 分类列表 |

## 请求示例

### 诗歌列表

```
GET /api/user/poems?page=1&page_size=10&category_id=1
```

### 搜索诗歌

```
GET /api/user/poems/search?keyword=李白&page=1&page_size=10
```

## 异常处理

| 场景 | 错误码 | 提示信息 |
|-----|--------|---------|
| 诗歌不存在 | 404 | 诗歌不存在或已下架 |
| 分类不存在 | 404 | 分类不存在 |
| 搜索关键词为空 | 400 | 请输入搜索关键词 |
| 分页参数非法 | 400 | 分页参数错误 |
