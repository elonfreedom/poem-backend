# 数据统计

## 功能概述

数据统计模块为管理员提供用户增长、浏览量、活跃度、打卡率等数据的可视化展示，帮助运营决策。

## 功能详情

### 1. 总览数据

**功能说明**：展示系统核心指标。

**数据项**：
| 指标 | 说明 |
|-----|------|
| 总用户数 | 注册用户总数 |
| 总诗歌数 | 诗歌总数 |
| 总浏览量 | 诗歌浏览总次数 |
| 今日活跃 | 今日活跃用户数 |
| 今日打卡 | 今日打卡用户数 |

### 2. 每日统计

**功能说明**：展示指定日期范围的每日统计数据。

**数据项**：
| 指标 | 说明 |
|-----|------|
| 日期 | 统计日期 |
| 新增用户 | 当日新增用户数 |
| 活跃用户 | 当日活跃用户数 |
| 浏览量 | 当日浏览量 |
| 打卡数 | 当日打卡数 |

**参数**：
- start_date：开始日期
- end_date：结束日期（默认今天）

### 3. 热门诗歌

**功能说明**：展示浏览量最高的诗歌排行。

**业务规则**：
- 按浏览量降序排列
- 默认展示前 20 名
- 支持指定日期范围

### 4. 用户增长

**功能说明**：展示用户增长趋势。

**业务规则**：
- 按天统计新增用户
- 支持指定日期范围
- 展示累计用户数

## 数据模型

```go
// StatsOverview 总览数据
type StatsOverview struct {
    TotalUsers    int64 `json:"total_users"`
    TotalPoems    int64 `json:"total_poems"`
    TotalViews    int64 `json:"total_views"`
    TodayActive   int64 `json:"today_active"`
    TodayCheckIns int64 `json:"today_check_ins"`
}

// DailyStats 每日统计
type DailyStats struct {
    Date        time.Time `json:"date"`
    NewUsers    int64     `json:"new_users"`
    ActiveUsers int64     `json:"active_users"`
    Views       int64     `json:"views"`
    CheckIns    int64     `json:"check_ins"`
}

// DailyStatsResponse 每日统计响应
type DailyStatsResponse struct {
    Total int          `json:"total"`
    List  []DailyStats `json:"list"`
}

// HotPoem 热门诗歌
type HotPoem struct {
    PoemID    int64  `json:"poem_id"`
    Title     string `json:"title"`
    Author    string `json:"author"`
    ViewCount int64  `json:"view_count"`
}

// HotPoemsResponse 热门诗歌响应
type HotPoemsResponse struct {
    Total int       `json:"total"`
    List  []HotPoem `json:"list"`
}

// UserGrowth 用户增长
type UserGrowth struct {
    Date          time.Time `json:"date"`
    NewUsers      int64     `json:"new_users"`
    TotalUsers    int64     `json:"total_users"`
}

// UserGrowthResponse 用户增长响应
type UserGrowthResponse struct {
    Total int          `json:"total"`
    List  []UserGrowth `json:"list"`
}

// DateRangeRequest 日期范围请求
type DateRangeRequest struct {
    StartDate string `json:"start_date" validate:"required,datetime=2006-01-02"`
    EndDate   string `json:"end_date" validate:"required,datetime=2006-01-02"`
}
```

## API 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/admin/stats/overview | 总览数据 |
| GET | /api/admin/stats/daily | 每日统计（支持日期范围） |
| GET | /api/admin/stats/poems/hot | 热门诗歌 |
| GET | /api/admin/stats/users/growth | 用户增长 |

## 请求示例

### 每日统计

```
GET /api/admin/stats/daily?start_date=2026-08-01&end_date=2026-08-21
```

### 热门诗歌

```
GET /api/admin/stats/poems/hot?limit=20&start_date=2026-08-01&end_date=2026-08-21
```

## 异常处理

| 场景 | 错误码 | 提示信息 |
|-----|--------|---------|
| 无权限 | 403 | 无权限操作 |
| 日期格式错误 | 400 | 日期格式错误，请使用 YYYY-MM-DD |
| 开始日期晚于结束日期 | 400 | 开始日期不能晚于结束日期 |
