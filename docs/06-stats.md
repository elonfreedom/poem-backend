# 数据统计（Admin）

## 功能

- 用户增长趋势
- 诗歌浏览量统计
- 活跃用户统计
- 打卡率统计
- 热门诗歌排行

## 数据模型

```go
// StatsOverview 总览数据
type StatsOverview struct {
    TotalUsers    int64
    TotalPoems    int64
    TotalViews    int64
    TodayActive   int64
    TodayCheckIns int64
}

// DailyStats 每日统计
type DailyStats struct {
    Date        time.Time
    NewUsers    int64
    ActiveUsers int64
    Views       int64
    CheckIns    int64
}
```

## API 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/stats/overview | 总览数据 |
| GET | /api/admin/stats/daily | 每日统计（支持日期范围） |
| GET | /api/admin/stats/poems/hot | 热门诗歌 |
| GET | /api/admin/stats/users/growth | 用户增长 |
