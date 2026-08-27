package usermodel

import "time"

// ReadingPlan 阅读计划
type ReadingPlan struct {
	UserID     string    `json:"user_id" db:"user_id" description:"用户ID (UUID v7)"`
	PlanID     int       `json:"plan_id" db:"plan_id" description:"计划ID（用户级自增）"`
	Title      string    `json:"title" db:"title" description:"计划标题"`
	DailyCount int       `json:"daily_count" db:"daily_count" description:"每日阅读数量"`
	StartDate  time.Time `json:"start_date" db:"start_date" description:"开始日期"`
	EndDate    time.Time `json:"end_date" db:"end_date" description:"结束日期"`
	Status     string    `json:"status" db:"status" description:"状态: active, completed, paused"`
	CreatedAt  time.Time `json:"created_at" db:"created_at" description:"创建时间"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at" description:"更新时间"`
}

// ReadingProgress 阅读进度
type ReadingProgress struct {
	UserID    string    `json:"user_id" db:"user_id" description:"用户ID (UUID v7)"`
	Date      time.Time `json:"date" db:"date" description:"日期"`
	ReadCount int       `json:"read_count" db:"read_count" description:"已读数量"`
	PoemIDs   []int64   `json:"poem_ids" db:"poem_ids" description:"已读诗歌ID列表"`
	CreatedAt time.Time `json:"created_at" db:"created_at" description:"创建时间"`
}

// CreatePlanRequest 创建计划请求
type CreatePlanRequest struct {
	Title      string `json:"title" description:"计划标题"`
	DailyCount int    `json:"daily_count" validate:"required,min=1,max=50" description:"每日阅读数量（1-50）"`
	Duration   int    `json:"duration" validate:"required,oneof=7 14 30 90" description:"计划天数（7/14/30/90）"`
	StartDate  string `json:"start_date" description:"开始日期（YYYY-MM-DD，可选，默认今天）"`
}

// CreatePlanResponse 创建计划响应
type CreatePlanResponse struct {
	PlanID     int       `json:"plan_id" description:"计划ID"`
	Title      string    `json:"title" description:"计划标题"`
	DailyCount int       `json:"daily_count" description:"每日阅读数量"`
	StartDate  time.Time `json:"start_date" description:"开始日期"`
	EndDate    time.Time `json:"end_date" description:"结束日期"`
	Status     string    `json:"status" description:"状态"`
}

// LogReadingRequest 记录阅读请求
type LogReadingRequest struct {
	PoemIDs []int64 `json:"poem_ids" validate:"required,min=1" description:"已读诗歌ID列表"`
}

// LogReadingResponse 记录阅读响应
type LogReadingResponse struct {
	TodayCount    int  `json:"today_count" description:"今日已读数量"`
	TargetCount   int  `json:"target_count" description:"目标数量"`
	IsTodayFinish bool `json:"is_today_finish" description:"今日是否完成"`
}

// PlanProgressResponse 计划进度响应
type PlanProgressResponse struct {
	PlanID         int             `json:"plan_id" description:"计划ID"`
	DailyCount     int             `json:"daily_count" description:"每日阅读数量"`
	StartDate      time.Time       `json:"start_date" description:"开始日期"`
	EndDate        time.Time       `json:"end_date" description:"结束日期"`
	Status         string          `json:"status" description:"状态"`
	TotalDays      int             `json:"total_days" description:"总天数"`
	CompletedDays  int             `json:"completed_days" description:"已完成天数"`
	CompletionRate float64         `json:"completion_rate" description:"完成率（百分比）"`
	DailyProgress  []DailyProgress `json:"daily_progress" description:"每日进度"`
}

// DailyProgress 每日进度
type DailyProgress struct {
	Date      time.Time `json:"date" description:"日期"`
	ReadCount int       `json:"read_count" description:"已读数量"`
	Target    int       `json:"target" description:"目标数量"`
	IsReached bool      `json:"is_reached" description:"是否达标"`
	PoemTitle string    `json:"poem_title" description:"当天打卡的诗文标题（未打卡为空）"`
}

// CurrentPlanResponse 当前计划响应
type CurrentPlanResponse struct {
	Plan           CreatePlanResponse `json:"plan" description:"计划信息"`
	TodayCount     int                `json:"today_count" description:"今日已读数量"`
	IsTodayFinish  bool               `json:"is_today_finish" description:"今日是否完成"`
	CompletionRate float64            `json:"completion_rate" description:"完成率（百分比）"`
}
