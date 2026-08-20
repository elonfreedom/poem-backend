# 诗歌模块

## Admin 端

### 内容管理

- 诗歌 CRUD（创建、查看、编辑、删除）
- 草稿/发布/归档状态管理
- 批量操作

### 分类管理

- 分类 CRUD（如：唐诗、宋词、元曲）
- 分类排序

### 标签管理

- 标签 CRUD（如：思乡、爱国、爱情）
- 标签关联诗歌

## User 端

### 功能

- 诗歌列表（分页、分类筛选）
- 诗歌详情（含翻译、赏析）
- 关键词搜索（标题、作者、内容）
- 每日推荐

## 数据模型

```go
// Poem 诗歌模型
type Poem struct {
    ID           int64
    Title        string
    Author       string
    Dynasty      string    // 朝代
    Content      string    // 诗歌内容
    Translation  string    // 翻译
    Appreciation string    // 赏析
    CategoryID   int64
    Tags         []string
    CoverURL     string
    Status       string    // draft, published, archived
    CreatedBy    int64
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// Category 分类
type Category struct {
    ID        int64
    Name      string
    Sort      int
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Tag 标签
type Tag struct {
    ID        int64
    Name      string
    CreatedAt time.Time
}

// PoemTag 诗歌标签关联
type PoemTag struct {
    PoemID int64
    TagID  int64
}
```

## API 接口

### Admin 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/poems | 诗歌列表（分页、筛选） |
| POST | /api/admin/poems | 创建诗歌 |
| GET | /api/admin/poems/:id | 诗歌详情 |
| PUT | /api/admin/poems/:id | 更新诗歌 |
| DELETE | /api/admin/poems/:id | 删除诗歌 |
| PUT | /api/admin/poems/:id/status | 更新状态 |
| GET | /api/admin/categories | 分类列表 |
| POST | /api/admin/categories | 创建分类 |
| PUT | /api/admin/categories/:id | 更新分类 |
| DELETE | /api/admin/categories/:id | 删除分类 |
| GET | /api/admin/tags | 标签列表 |
| POST | /api/admin/tags | 创建标签 |
| DELETE | /api/admin/tags/:id | 删除标签 |

### User 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/user/poems | 诗歌列表 |
| GET | /api/user/poems/:id | 诗歌详情 |
| GET | /api/user/poems/search | 搜索诗歌 |
| GET | /api/user/poems/daily | 每日推荐 |
| GET | /api/user/categories | 分类列表 |
