# 阅读计划

## 功能概述

阅读计划帮助用户设定每日阅读目标，记录每日阅读进度，培养持续阅读的习惯。

## 用户故事

- 作为用户，我想创建阅读计划，设定每日阅读目标
- 作为用户，我想记录今天的阅读，跟踪进度
- 作为用户，我想查看当前计划进度，了解完成情况
- 作为用户，我想暂停计划，暂时中断后恢复

## 功能详情

### 1. 创建阅读计划

**功能说明**：用户创建一个新的阅读计划。

**业务规则**：
- 同一用户只能有一个进行中的计划（active）
- 每日阅读数量目标：1-50 首
- 计划周期：7天、14天、30天、90天
- 开始日期默认为当天
- 结束日期 = 开始日期 + 周期天数 - 1

### 2. 记录阅读

**功能说明**：记录用户今天阅读了哪些诗歌。

**业务规则**：
- 每天可多次记录，累计计算当日阅读量
- 阅读量达到目标后，当日状态为"已完成"
- 同一诗歌当天多次阅读只计一次
- 记录阅读时自动更新进度

### 3. 查看当前计划

**功能说明**：获取用户当前进行中的计划。

**业务规则**：
- 返回计划详情 + 今日进度 + 整体完成率
- 无进行中计划时返回空

### 4. 查看计划进度

**功能说明**：查看指定计划的详细进度。

**业务规则**：
- 展示每日阅读情况（日期、阅读量、是否达标）
- 展示整体完成率
- 展示连续达标天数

### 5. 暂停/恢复计划

**功能说明**：暂停或恢复进行中的计划。

**业务规则**：
- 暂停后计划状态变为 paused
- 恢复后状态变回 active
- 暂停期间不计入进度

## 数据模型

```go
// ReadingPlan 阅读计划（复合主键：user_id + plan_id）
// plan_id 为用户级自增，支持未来计划分享场景
type ReadingPlan struct {
    UserID     string    // UUID v7
    PlanID     int       // 用户级自增 ID
    DailyCount int       // 每日阅读数量目标
    StartDate  time.Time // 计划开始日期
    EndDate    time.Time // 计划结束日期
    Status     string    // active, completed, paused
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

// ReadingProgress 阅读进度（复合主键：user_id + date）
type ReadingProgress struct {
    UserID    string    // UUID v7
    Date      time.Time // 阅读日期
    ReadCount int       // 当日阅读数量
    PoemIDs   []int64   // 阅读的诗歌ID列表
    CreatedAt time.Time
}

// CreatePlanRequest 创建计划请求
type CreatePlanRequest struct {
    DailyCount int `json:"daily_count" validate:"required,min=1,max=50"`
    Duration   int `json:"duration" validate:"required,oneof=7 14 30 90"`
}

// CreatePlanResponse 创建计划响应
type CreatePlanResponse struct {
    PlanID     int       `json:"plan_id"`
    DailyCount int       `json:"daily_count"`
    StartDate  time.Time `json:"start_date"`
    EndDate    time.Time `json:"end_date"`
    Status     string    `json:"status"`
}

// LogReadingRequest 记录阅读请求
type LogReadingRequest struct {
    PoemIDs []int64 `json:"poem_ids" validate:"required,min=1"`
}

// LogReadingResponse 记录阅读响应
type LogReadingResponse struct {
    TodayCount    int  `json:"today_count"`
    TargetCount   int  `json:"target_count"`
    IsTodayFinish bool `json:"is_today_finish"`
}

// PlanProgressResponse 计划进度响应
type PlanProgressResponse struct {
    PlanID          int              `json:"plan_id"`
    DailyCount      int              `json:"daily_count"`
    StartDate       time.Time        `json:"start_date"`
    EndDate         time.Time        `json:"end_date"`
    Status          string           `json:"status"`
    TotalDays       int              `json:"total_days"`
    CompletedDays   int              `json:"completed_days"`
    CompletionRate  float64          `json:"completion_rate"`
    DailyProgress   []DailyProgress  `json:"daily_progress"`
}

// DailyProgress 每日进度
type DailyProgress struct {
    Date      time.Time `json:"date"`
    ReadCount int       `json:"read_count"`
    Target    int       `json:"target"`
    IsReached bool      `json:"is_reached"`
}

// CurrentPlanResponse 当前计划响应
type CurrentPlanResponse struct {
    Plan           CreatePlanResponse `json:"plan"`
    TodayCount     int                `json:"today_count"`
    IsTodayFinish  bool               `json:"is_today_finish"`
    CompletionRate float64            `json:"completion_rate"`
}
```

## API 接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/user/reading-plans | 创建阅读计划 |
| GET | /api/user/reading-plans/current | 当前计划 |
| PUT | /api/user/reading-plans/:id/pause | 暂停计划 |
| PUT | /api/user/reading-plans/:id/resume | 恢复计划 |
| GET | /api/user/reading-plans/:id/progress | 计划进度 |
| POST | /api/user/reading-plans/log | 记录今日阅读 |

## 请求示例

### 创建计划

```json
POST /api/user/reading-plans
{
    "daily_count": 3,
    "duration": 30
}
```

### 记录阅读

```json
POST /api/user/reading-plans/log
{
    "poem_ids": [1, 2, 3]
}
```

## 异常处理

| 场景 | 错误码 | 提示信息 |
|-----|--------|---------|
| 未登录 | 401 | 请先登录 |
| 已有进行中的计划 | 400 | 请先完成或暂停当前计划 |
| 每日目标超出范围 | 400 | 每日目标需在 1-50 之间 |
| 计划不存在 | 404 | 计划不存在 |
| 计划不属于当前用户 | 403 | 无权操作此计划 |
| 诗歌不存在 | 400 | 部分诗歌不存在 |
