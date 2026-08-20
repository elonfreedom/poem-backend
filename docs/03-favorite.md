# 收藏模块

## 功能

- 收藏/取消收藏诗歌
- 查看收藏列表

## 数据模型

```go
// Favorite 收藏模型
type Favorite struct {
    ID        int64
    UserID    int64
    PoemID    int64
    CreatedAt time.Time
}

// FavoriteResponse 收藏响应
type FavoriteResponse struct {
    ID        int64
    Poem      PoemResponse
    CreatedAt time.Time
}

// FavoriteRequest 收藏请求
type FavoriteRequest struct {
    PoemID int64 `json:"poem_id" validate:"required"`
}
```

## API 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/user/favorites | 收藏诗歌 |
| DELETE | /api/user/favorites/:poem_id | 取消收藏 |
| GET | /api/user/favorites | 收藏列表（分页） |
