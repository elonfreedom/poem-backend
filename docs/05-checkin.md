# 打卡系统

## 功能

- 每日打卡（完成阅读后打卡）
- 打卡记录查询
- 连续打卡天数统计
- 打卡排行榜

## 数据模型

```go
// CheckIn 打卡记录
type CheckIn struct {
    ID             int64
    UserID         int64
    Date           time.Time // 打卡日期（唯一）
    ConsecutiveDay int       // 连续打卡天数
    CreatedAt      time.Time
}

// CheckInStats 打卡统计
type CheckInStats struct {
    UserID         int64
    TotalDays      int       // 总打卡天数
    ConsecutiveDay int       // 当前连续打卡天数
    MaxConsecutive int       // 最长连续打卡天数
    LastCheckIn    time.Time
}
```

## API 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/user/checkins | 每日打卡 |
| GET | /api/user/checkins | 打卡记录（分页） |
| GET | /api/user/checkins/stats | 打卡统计 |
| GET | /api/user/checkins/calendar | 打卡日历（某月打卡情况） |
| GET | /api/user/checkins/ranking | 打卡排行榜 |
