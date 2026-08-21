# 管理端 API 接口汇总

## 管理接口（需 Admin 认证）

### 诗歌管理

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/poems | 诗歌列表（分页、筛选） |
| POST | /api/admin/poems | 创建诗歌 |
| GET | /api/admin/poems/:id | 诗歌详情 |
| PUT | /api/admin/poems/:id | 更新诗歌 |
| DELETE | /api/admin/poems/:id | 删除诗歌 |
| PUT | /api/admin/poems/:id/status | 更新状态 |

### 分类管理

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/categories | 分类列表 |
| POST | /api/admin/categories | 创建分类 |
| PUT | /api/admin/categories/:id | 更新分类 |
| DELETE | /api/admin/categories/:id | 删除分类 |

### 标签管理

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/tags | 标签列表 |
| POST | /api/admin/tags | 创建标签 |
| DELETE | /api/admin/tags/:id | 删除标签 |

### 数据统计

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/stats/overview | 总览数据 |
| GET | /api/admin/stats/daily | 每日统计 |
| GET | /api/admin/stats/poems/hot | 热门诗歌 |
| GET | /api/admin/stats/users/growth | 用户增长 |

### Banner 管理

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/banners | Banner 列表 |
| POST | /api/admin/banners | 创建 Banner |
| PUT | /api/admin/banners/:id | 更新 Banner |
| DELETE | /api/admin/banners/:id | 删除 Banner |

### 公告管理

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/announcements | 公告列表 |
| POST | /api/admin/announcements | 创建公告 |
| PUT | /api/admin/announcements/:id | 更新公告 |
| DELETE | /api/admin/announcements/:id | 删除公告 |

### 系统配置

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/config | 获取配置列表 |
| GET | /api/admin/config/:key | 获取单个配置 |
| PUT | /api/admin/config | 更新配置 |

## 接口规范

### 请求头

```
Authorization: Bearer <admin_token>
Content-Type: application/json
```

### 分页参数

| 参数 | 类型 | 默认值 | 说明 |
|-----|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 20 | 每页数量（最大 100） |

### 响应格式

```json
{
    "code": 200,
    "message": "success",
    "data": {}
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限（非 Admin） |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
