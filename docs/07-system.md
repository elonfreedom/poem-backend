# 系统配置（Admin）

## 功能

- Banner 管理
- 公告管理
- 系统参数配置

## 数据模型

```go
// Banner 轮播图
type Banner struct {
    ID        int64
    Title     string
    ImageURL  string
    LinkType  string // poem, url
    LinkValue string
    Sort      int
    Status    string // active, inactive
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Announcement 公告
type Announcement struct {
    ID        int64
    Title     string
    Content   string
    Status    string // draft, published
    CreatedAt time.Time
    UpdatedAt time.Time
}

// SystemConfig 系统配置
type SystemConfig struct {
    ID        int64
    Key       string
    Value     string
    Remark    string
    UpdatedAt time.Time
}
```

## API 接口

### Admin 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/banners | Banner 列表 |
| POST | /api/admin/banners | 创建 Banner |
| PUT | /api/admin/banners/:id | 更新 Banner |
| DELETE | /api/admin/banners/:id | 删除 Banner |
| GET | /api/admin/announcements | 公告列表 |
| POST | /api/admin/announcements | 创建公告 |
| PUT | /api/admin/announcements/:id | 更新公告 |
| DELETE | /api/admin/announcements/:id | 删除公告 |
| GET | /api/admin/config | 获取配置 |
| PUT | /api/admin/config | 更新配置 |

### 公开接口（前台展示）

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/public/banners | Banner 列表 |
| GET | /api/public/announcements | 公告列表 |
