# 系统配置

## 功能概述

系统配置模块提供 Banner 轮播图、公告、系统参数的配置功能，支持前台展示内容的动态管理。

## 功能详情

### 1. Banner 管理

**功能说明**：管理首页轮播图。

**字段**：
| 字段 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| title | string | 是 | 标题 |
| image_url | string | 是 | 图片地址 |
| link_type | string | 是 | 链接类型：poem（诗歌）、url（外链） |
| link_value | string | 是 | 链接值（诗歌ID 或 URL） |
| sort | int | 否 | 排序值（升序） |
| status | string | 是 | 状态：active（启用）、inactive（禁用） |

**业务规则**：
- 最多启用 5 个 Banner
- 前台按 sort 升序展示启用的 Banner
- 点击 Banner 跳转到对应诗歌或外链

### 2. 公告管理

**功能说明**：管理首页公告。

**字段**：
| 字段 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| title | string | 是 | 公告标题 |
| content | string | 是 | 公告内容 |
| status | string | 是 | 状态：draft（草稿）、published（已发布） |

**业务规则**：
- 前台只展示已发布的公告
- 支持设置展示时间范围

### 3. 系统参数配置

**功能说明**：管理系统运行参数。

**配置项**：
| Key | 说明 | 默认值 |
|-----|------|--------|
| daily_poem_id | 每日推荐诗歌 ID | 无 |
| app_version | 当前 App 版本 | 1.0.0 |
| min_version | 最低支持版本 | 1.0.0 |
| force_update | 是否强制更新 | false |

**业务规则**：
- 配置变更实时生效
- 支持 key-value 形式存储

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

// BannerRequest Banner 请求
type BannerRequest struct {
    Title     string `json:"title" validate:"required,max=100"`
    ImageURL  string `json:"image_url" validate:"required,url"`
    LinkType  string `json:"link_type" validate:"required,oneof=poem url"`
    LinkValue string `json:"link_value" validate:"required"`
    Sort      int    `json:"sort" validate:"min=0"`
    Status    string `json:"status" validate:"required,oneof=active inactive"`
}

// BannerResponse Banner 响应
type BannerResponse struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    ImageURL  string `json:"image_url"`
    LinkType  string `json:"link_type"`
    LinkValue string `json:"link_value"`
    Sort      int    `json:"sort"`
    Status    string `json:"status"`
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

// AnnouncementRequest 公告请求
type AnnouncementRequest struct {
    Title   string `json:"title" validate:"required,max=100"`
    Content string `json:"content" validate:"required"`
    Status  string `json:"status" validate:"required,oneof=draft published"`
}

// AnnouncementResponse 公告响应
type AnnouncementResponse struct {
    ID        int64     `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

// SystemConfig 系统配置
type SystemConfig struct {
    ID        int64
    Key       string
    Value     string
    Remark    string
    UpdatedAt time.Time
}

// SystemConfigRequest 系统配置请求
type SystemConfigRequest struct {
    Key   string `json:"key" validate:"required"`
    Value string `json:"value" validate:"required"`
    Remark string `json:"remark"`
}

// SystemConfigResponse 系统配置响应
type SystemConfigResponse struct {
    Key       string `json:"key"`
    Value     string `json:"value"`
    Remark    string `json:"remark"`
}
```

## API 接口

### Admin 接口

#### Banner

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/banners | Banner 列表 |
| POST | /api/admin/banners | 创建 Banner |
| PUT | /api/admin/banners/:id | 更新 Banner |
| DELETE | /api/admin/banners/:id | 删除 Banner |

#### 公告

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/announcements | 公告列表 |
| POST | /api/admin/announcements | 创建公告 |
| PUT | /api/admin/announcements/:id | 更新公告 |
| DELETE | /api/admin/announcements/:id | 删除公告 |

#### 系统配置

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/config | 获取配置列表 |
| GET | /api/admin/config/:key | 获取单个配置 |
| PUT | /api/admin/config | 更新配置 |

### 公开接口（前台展示）

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/public/banners | Banner 列表（仅启用） |
| GET | /api/public/announcements | 公告列表（仅已发布） |

## 异常处理

| 场景 | 错误码 | 提示信息 |
|-----|--------|---------|
| 无权限 | 403 | 无权限操作 |
| Banner 不存在 | 404 | Banner 不存在 |
| 启用 Banner 超限 | 400 | 最多启用 5 个 Banner |
| 公告不存在 | 404 | 公告不存在 |
| 配置 Key 不存在 | 404 | 配置不存在 |
