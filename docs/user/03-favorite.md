# 收藏模块

## 功能概述

收藏模块允许用户收藏感兴趣的诗歌，方便后续查阅。支持收藏/取消收藏和收藏列表查看。

## 用户故事

- 作为用户，我想收藏喜欢的诗歌，方便以后再次阅读
- 作为用户，我想取消不再感兴趣的收藏
- 作为用户，我想查看我的收藏列表，快速找到已收藏的诗歌

## 功能详情

### 1. 收藏诗歌

**功能说明**：将指定诗歌添加到收藏列表。

**业务规则**：
- 同一诗歌不可重复收藏（幂等操作）
- 收藏状态在诗歌详情接口返回
- 需登录后才能收藏

### 2. 取消收藏

**功能说明**：从收藏列表中移除指定诗歌。

**业务规则**：
- 取消不存在的收藏返回成功（幂等操作）
- 支持通过诗歌 ID 取消收藏

### 3. 收藏列表

**功能说明**：分页查看用户的所有收藏。

**业务规则** |
- 按收藏时间倒序排列
- 展示诗歌完整信息（标题、作者、分类等）
- 分页参数：page（默认1）、page_size（默认10，最大50）

## 数据模型

```go
// Favorite 收藏模型（复合主键：user_id + poem_id）
type Favorite struct {
    UserID    string    // UUID v7
    PoemID    int64     // 自增 ID
    CreatedAt time.Time
}

// FavoriteResponse 收藏响应
type FavoriteResponse struct {
    Poem      PoemListItem `json:"poem"`
    CreatedAt time.Time    `json:"created_at"`
}

// FavoriteRequest 收藏请求
type FavoriteRequest struct {
    PoemID int64 `json:"poem_id" validate:"required,min=1"`
}

// FavoriteListResponse 收藏列表响应
type FavoriteListResponse struct {
    Total int                `json:"total"`
    List  []FavoriteResponse `json:"list"`
}
```

## API 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/user/favorites | 收藏诗歌 |
| DELETE | /api/user/favorites/:poem_id | 取消收藏 |
| GET | /api/user/favorites | 收藏列表（分页） |

## 请求示例

### 收藏诗歌

```json
POST /api/user/favorites
{
    "poem_id": 1
}
```

### 取消收藏

```
DELETE /api/user/favorites/1
```

### 收藏列表

```
GET /api/user/favorites?page=1&page_size=10
```

## 异常处理

| 场景 | 错误码 | 提示信息 |
|-----|--------|---------|
| 未登录 | 401 | 请先登录 |
| 诗歌不存在 | 404 | 诗歌不存在 |
| 已收藏 | 200 | 收藏成功（幂等） |
| poem_id 非法 | 400 | 参数错误 |
