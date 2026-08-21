# 打卡系统

## 功能概述

打卡系统是用户端的激励模块，用户完成每日阅读后可打卡，系统记录连续打卡天数，并提供排行榜激励用户坚持。

## 用户故事

- 作为用户，我想每天打卡，记录我的坚持
- 作为用户，我想查看我的打卡记录和统计
- 作为用户，我想看打卡日历，了解每月打卡情况
- 作为用户，我想看排行榜，与其他用户比较

## 功能详情

### 1. 每日打卡

**功能说明**：用户完成每日阅读后进行打卡。

**业务规则**：
- 每天只能打卡一次（同一日期重复打卡返回成功，幂等）
- 连续打卡天数 = 上次打卡日期 +1 天 = 今天，则 +1；否则重置为 1
- 打卡日期以服务器时间为准（按天计算）
- 打卡后更新用户的打卡统计表

### 2. 打卡记录

**功能说明**：分页查看用户的历史打卡记录。

**业务规则**：
- 按打卡日期倒序排列
- 展示日期和当日连续打卡天数

### 3. 打卡统计

**功能说明**：查看用户的打卡统计数据。

**统计项**：
- 总打卡天数
- 当前连续打卡天数
- 最长连续打卡天数
- 最近打卡日期

### 4. 打卡日历

**功能说明**：展示某月的打卡情况。

**业务规则**：
- 默认展示当前月
- 支持通过参数指定年月
- 每天标记是否打卡

### 5. 打卡排行榜

**功能说明**：展示用户打卡排名。

**业务规则**：
- 按连续打卡天数降序排列
- 连续天数相同则按最近打卡时间排序
- 展示前 100 名
- 展示用户自己的排名

## 数据模型

```go
// CheckIn 打卡记录（复合主键：user_id + date）
type CheckIn struct {
    UserID         string    // UUID v7
    Date           time.Time // 打卡日期
    ConsecutiveDay int       // 连续打卡天数
    CreatedAt      time.Time
}

// CheckInStats 打卡统计（主键：user_id）
type CheckInStats struct {
    UserID         string    // UUID v7
    TotalDays      int       // 总打卡天数
    ConsecutiveDay int       // 当前连续打卡天数
    MaxConsecutive int       // 最长连续打卡天数
    LastCheckIn    time.Time // 最近打卡日期
}

// CheckInResponse 打卡响应
type CheckInResponse struct {
    Date           time.Time `json:"date"`
    ConsecutiveDay int       `json:"consecutive_day"`
}

// CheckInStatsResponse 打卡统计响应
type CheckInStatsResponse struct {
    TotalDays      int       `json:"total_days"`
    ConsecutiveDay int       `json:"consecutive_day"`
    MaxConsecutive int       `json:"max_consecutive"`
    LastCheckIn    time.Time `json:"last_check_in"`
}

// CheckInCalendarResponse 打卡日历响应
type CheckInCalendarResponse struct {
    Year  int           `json:"year"`
    Month int           `json:"month"`
    Days  []CalendarDay `json:"days"`
}

// CalendarDay 日历天
type CalendarDay struct {
    Day       int  `json:"day"`
    IsChecked bool `json:"is_checked"`
}

// RankingItem 排行榜项
type RankingItem struct {
    Rank           int    `json:"rank"`
    Nickname       string `json:"nickname"`
    ConsecutiveDay int    `json:"consecutive_day"`
}

// RankingResponse 排行榜响应
type RankingResponse struct {
    Total            int           `json:"total"`
    MyRank           int           `json:"my_rank"`
    MyConsecutiveDay int           `json:"my_consecutive_day"`
    List             []RankingItem `json:"list"`
}

// CheckInListResponse 打卡记录列表响应
type CheckInListResponse struct {
    Total int               `json:"total"`
    List  []CheckInResponse `json:"list"`
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

## 请求示例

### 打卡

```
POST /api/user/checkins
```

### 打卡记录

```
GET /api/user/checkins?page=1&page_size=30
```

### 打卡日历

```
GET /api/user/checkins/calendar?year=2026&month=8
```

### 排行榜

```
GET /api/user/checkins/ranking
```

## 异常处理

| 场景 | 错误码 | 提示信息 |
|-----|--------|---------|
| 未登录 | 401 | 请先登录 |
| 重复打卡 | 200 | 已打卡（幂等） |
| 日期参数非法 | 400 | 日期参数错误 |
