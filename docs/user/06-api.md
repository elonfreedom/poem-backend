# 用户端 API 接口汇总

## 公开接口（无需认证）

### Passkey 注册登录

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/public/passkey/register/begin | 开始 Passkey 注册 |
| POST | /api/public/passkey/register/finish | 完成 Passkey 注册 |
| POST | /api/public/passkey/login/begin | 开始 Passkey 登录 |
| POST | /api/public/passkey/login/finish | 完成 Passkey 登录 |

### 账号恢复

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/public/recovery/request | 请求账号恢复（发送邮件） |
| POST | /api/public/recovery/verify | 验证恢复，绑定新 Passkey |

### 系统

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/public/banners | Banner 列表 |
| GET | /api/public/announcements | 公告列表 |

## 用户接口（需 User 认证）

### 用户系统

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/user/profile | 获取个人信息 |
| PUT | /api/user/profile | 更新个人信息（昵称/头像） |
| POST | /api/user/email/bind | 绑定邮箱 |
| GET | /api/user/email/verify | 验证邮箱 |
| GET | /api/user/passkeys | 查看 Passkey 列表 |
| DELETE | /api/user/passkeys/:id | 删除 Passkey |
| GET | /api/user/recovery-code | 获取恢复码 |
| POST | /api/user/recovery-code/regenerate | 重新生成恢复码 |

### 诗歌浏览

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/user/poems | 诗歌列表（支持分类筛选、分页） |
| GET | /api/user/poems/:id | 诗歌详情 |
| GET | /api/user/poems/search | 搜索诗歌 |
| GET | /api/user/poems/daily | 每日推荐 |
| GET | /api/user/categories | 分类列表 |

### 收藏

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/user/favorites | 收藏诗歌 |
| DELETE | /api/user/favorites/:poem_id | 取消收藏 |
| GET | /api/user/favorites | 收藏列表（分页） |

### 阅读计划

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/user/reading-plans | 创建阅读计划 |
| GET | /api/user/reading-plans/current | 当前计划 |
| PUT | /api/user/reading-plans/:id/pause | 暂停计划 |
| PUT | /api/user/reading-plans/:id/resume | 恢复计划 |
| GET | /api/user/reading-plans/:id/progress | 计划进度 |
| POST | /api/user/reading-plans/log | 记录今日阅读 |

### 打卡系统

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/user/checkins | 每日打卡 |
| GET | /api/user/checkins | 打卡记录（分页） |
| GET | /api/user/checkins/stats | 打卡统计 |
| GET | /api/user/checkins/calendar | 打卡日历 |
| GET | /api/user/checkins/ranking | 打卡排行榜 |

## 接口规范

### 请求头

```
Authorization: Bearer <token>
Content-Type: application/json
```

### 分页参数

| 参数 | 类型 | 默认值 | 说明 |
|-----|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页数量（最大 50） |

### 响应格式

```json
{
    "code": 200,
    "message": "success",
    "data": {}
}
```

### 列表响应格式

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "total": 100,
        "list": []
    }
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |
