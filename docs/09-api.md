# API 接口汇总

## 公开接口（无需认证）

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/public/login/phone | 手机号登录 |
| POST | /api/public/login/wechat | 微信登录 |
| POST | /api/public/login/apple | Apple 登录 |
| POST | /api/public/sms/send | 发送验证码 |
| GET | /api/public/banners | Banner 列表 |
| GET | /api/public/announcements | 公告列表 |

## 用户接口（需 User 认证）

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/user/profile | 获取个人信息 |
| PUT | /api/user/profile | 更新个人信息 |
| GET | /api/user/poems | 诗歌列表 |
| GET | /api/user/poems/:id | 诗歌详情 |
| GET | /api/user/poems/search | 搜索诗歌 |
| GET | /api/user/poems/daily | 每日推荐 |
| GET | /api/user/categories | 分类列表 |
| POST | /api/user/favorites | 收藏诗歌 |
| DELETE | /api/user/favorites/:poem_id | 取消收藏 |
| GET | /api/user/favorites | 收藏列表 |
| POST | /api/user/reading-plans | 创建阅读计划 |
| GET | /api/user/reading-plans/current | 当前计划 |
| PUT | /api/user/reading-plans/:id/pause | 暂停计划 |
| GET | /api/user/reading-plans/:id/progress | 计划进度 |
| POST | /api/user/reading-plans/log | 记录今日阅读 |
| POST | /api/user/checkins | 每日打卡 |
| GET | /api/user/checkins | 打卡记录 |
| GET | /api/user/checkins/stats | 打卡统计 |
| GET | /api/user/checkins/calendar | 打卡日历 |
| GET | /api/user/checkins/ranking | 打卡排行榜 |

## 管理接口（需 Admin 认证）

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/poems | 诗歌列表 |
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
| GET | /api/admin/stats/overview | 总览数据 |
| GET | /api/admin/stats/daily | 每日统计 |
| GET | /api/admin/stats/poems/hot | 热门诗歌 |
| GET | /api/admin/stats/users/growth | 用户增长 |
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
