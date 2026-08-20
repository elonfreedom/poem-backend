# 每日阅读计划

## 功能

- 用户可创建每日阅读计划
- 设置每日阅读诗歌数量目标
- 计划周期（连续N天）
- 查看计划进度

## 数据模型

```go
// ReadingPlan 阅读计划
type ReadingPlan struct {
    ID         int64
    UserID     int64
    DailyCount int       // 每日阅读数量目标
    StartDate  time.Time // 计划开始日期
    EndDate    time.Time // 计划结束日期
    Status     string    // active, completed, paused
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

// ReadingProgress 阅读进度
type ReadingProgress struct {
    ID        int64
    PlanID    int64
    UserID    int64
    Date      time.Time // 阅读日期
    ReadCount int       // 当日阅读数量
    PoemIDs   []int64   // 阅读的诗歌ID列表
    CreatedAt time.Time
}
```

## API 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/user/reading-plans | 创建阅读计划 |
| GET | /api/user/reading-plans/current | 当前计划 |
| PUT | /api/user/reading-plans/:id/pause | 暂停计划 |
| GET | /api/user/reading-plans/:id/progress | 计划进度 |
| POST | /api/user/reading-plans/log | 记录今日阅读 |
