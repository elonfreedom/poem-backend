package usermodel

import "time"

// SharedPlan 共享计划（社区发布的计划模板）
type SharedPlan struct {
	ID             int64     `json:"id" db:"id" description:"计划ID"`
	CreatorID      string    `json:"creator_id" db:"creator_id" description:"创建者用户ID"`
	CreatorName    string    `json:"creator_name" db:"creator_name" description:"创建者昵称"`
	Title          string    `json:"title" db:"title" description:"计划标题"`
	Description    string    `json:"description" db:"description" description:"计划描述"`
	Tags           []string  `json:"tags" db:"tags" description:"标签列表"`
	PoemIDs        []int64   `json:"poem_ids" db:"poem_ids" description:"有序的诗文ID列表"`
	DailyCount     int       `json:"daily_count" db:"daily_count" description:"每日阅读数量"`
	TotalDays      int       `json:"total_days" db:"total_days" description:"总天数"`
	SubscribeCount int       `json:"subscribe_count" db:"subscribe_count" description:"订阅数"`
	IsPublished    bool      `json:"is_published" db:"is_published" description:"是否已发布"`
	CreatedAt      time.Time `json:"created_at" db:"created_at" description:"创建时间"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at" description:"更新时间"`
}

// PlanSubscription 用户订阅关系
type PlanSubscription struct {
	ID            int64     `json:"id" db:"id" description:"订阅ID"`
	UserID        string    `json:"user_id" db:"user_id" description:"用户ID"`
	SharedPlanID  int64     `json:"shared_plan_id" db:"shared_plan_id" description:"共享计划ID"`
	StartDate     time.Time `json:"start_date" db:"start_date" description:"开始日期"`
	Status        string    `json:"status" db:"status" description:"状态: active, completed, paused"`
	CreatedAt     time.Time `json:"created_at" db:"created_at" description:"订阅时间"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at" description:"更新时间"`
}

// PlanCheckin 打卡记录
type PlanCheckin struct {
	ID             int64     `json:"id" db:"id" description:"打卡ID"`
	SubscriptionID int64     `json:"subscription_id" db:"subscription_id" description:"订阅ID"`
	UserID         string    `json:"user_id" db:"user_id" description:"用户ID"`
	DayNumber      int       `json:"day_number" db:"day_number" description:"第几天"`
	CheckinDate    time.Time `json:"checkin_date" db:"checkin_date" description:"打卡日期"`
	PoemIDs        []int64   `json:"poem_ids" db:"poem_ids" description:"当天打卡的诗歌ID"`
	CreatedAt      time.Time `json:"created_at" db:"created_at" description:"打卡时间"`
}

// ==================== 请求/响应结构体 ====================

// CreateSharedPlanRequest 创建共享计划请求
type CreateSharedPlanRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=100" description:"计划标题"`
	Description string   `json:"description" description:"计划描述"`
	Tags        []string `json:"tags" description:"标签列表"`
	PoemIDs     []int64  `json:"poem_ids" validate:"required,min=2" description:"有序的诗文ID列表（至少2首）"`
	DailyCount  int      `json:"daily_count" validate:"required,min=1,max=10" description:"每日阅读数量"`
}

// UpdateSharedPlanRequest 更新共享计划请求
type UpdateSharedPlanRequest struct {
	Title       string   `json:"title" validate:"omitempty,min=1,max=100" description:"计划标题"`
	Description string   `json:"description" description:"计划描述"`
	Tags        []string `json:"tags" description:"标签列表"`
	PoemIDs     []int64  `json:"poem_ids" validate:"omitempty,min=2" description:"有序的诗文ID列表"`
	DailyCount  int      `json:"daily_count" validate:"omitempty,min=1,max=10" description:"每日阅读数量"`
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	StartDate string `json:"start_date" description:"开始日期（YYYY-MM-DD，可选，默认今天）"`
}

// SetStartDateRequest 设置开始日期请求
type SetStartDateRequest struct {
	StartDate string `json:"start_date" validate:"required" description:"开始日期（YYYY-MM-DD）"`
}

// CheckinRequest 打卡请求
type CheckinRequest struct {
	PoemIDs []int64 `json:"poem_ids" validate:"required,min=1" description:"当天打卡的诗歌ID"`
}

// SharedPlanListItem 共享计划列表项
type SharedPlanListItem struct {
	ID             int64     `json:"id" description:"计划ID"`
	Title          string    `json:"title" description:"计划标题"`
	Description    string    `json:"description" description:"计划描述"`
	Tags           []string  `json:"tags" description:"标签列表"`
	DailyCount     int       `json:"daily_count" description:"每日阅读数量"`
	TotalDays      int       `json:"total_days" description:"总天数"`
	SubscribeCount int       `json:"subscribe_count" description:"订阅数"`
	CreatorName    string    `json:"creator_name" description:"创建者昵称"`
	CreatedAt      time.Time `json:"created_at" description:"创建时间"`
}

// SharedPlanDetail 共享计划详情
type SharedPlanDetail struct {
	SharedPlanListItem
	PoemIDs []int64     `json:"poem_ids" description:"有序的诗文ID列表"`
	Poems   []PoemBrief `json:"poems" description:"诗文详情列表"`
}

// PoemBrief 诗歌简要信息（用于计划详情）
type PoemBrief struct {
	ID      int64  `json:"id" description:"诗歌ID"`
	Title   string `json:"title" description:"标题"`
	Author  string `json:"author" description:"作者"`
	Dynasty string `json:"dynasty" description:"朝代"`
}

// SubscribeListResponse 订阅列表响应
type SubscribeListResponse struct {
	ID           int64     `json:"id" description:"订阅ID"`
	SharedPlanID int64     `json:"shared_plan_id" description:"共享计划ID"`
	Title        string    `json:"title" description:"计划标题"`
	Tags         []string  `json:"tags" description:"标签列表"`
	DailyCount   int       `json:"daily_count" description:"每日阅读数量"`
	TotalDays    int       `json:"total_days" description:"总天数"`
	StartDate    time.Time `json:"start_date" description:"开始日期"`
	Status       string    `json:"status" description:"状态"`
	CreatedAt    time.Time `json:"created_at" description:"订阅时间"`
}

// TodayPoemResponse 今日诗文响应
type TodayPoemResponse struct {
	DayNumber    int       `json:"day_number" description:"第几天"`
	Date         string    `json:"date" description:"日期"`
	Poems        []Poem    `json:"poems" description:"今日诗文列表"`
	IsCheckedIn  bool      `json:"is_checked_in" description:"今日是否已打卡"`
	CheckedAt    *string   `json:"checked_at" description:"今日打卡时间（未打卡为null）"`
	TotalDays    int       `json:"total_days" description:"总天数"`
	ProgressRate float64   `json:"progress_rate" description:"完成率"`
}

// Poem 诗歌简要信息（用于今日诗文）
type Poem struct {
	ID           int64     `json:"id" description:"诗歌ID"`
	Title        string    `json:"title" description:"标题"`
	Author       string    `json:"author" description:"作者"`
	Dynasty      string    `json:"dynasty" description:"朝代"`
	Content      string    `json:"content" description:"内容"`
	Translation  string    `json:"translation" description:"译文"`
	Appreciation string    `json:"appreciation" description:"赏析"`
	Tags         []string  `json:"tags" description:"标签"`
	LastCheckin  *CheckinInfo `json:"last_checkin" description:"上次打卡信息（从未打卡为null）"`
}

// CheckinInfo 打卡信息
type CheckinInfo struct {
	Date      string `json:"date" description:"上次打卡日期"`
	PlanTitle string `json:"plan_title" description:"在哪个计划打卡的"`
	DaysAgo   int    `json:"days_ago" description:"多少天前"`
}

// CheckinRecord 打卡记录（用于热力图）
type CheckinRecord struct {
	Date      string `json:"date" description:"打卡日期"`
	DayNumber int    `json:"day_number" description:"第几天"`
	PoemTitle string `json:"poem_title" description:"当天打卡的诗文标题"`
}

// CheckinsResponse 打卡记录列表响应
type CheckinsResponse struct {
	Total    int             `json:"total" description:"总打卡天数"`
	Items    []CheckinRecord `json:"items" description:"打卡记录列表"`
}

// SkipDayRequest 跳过天数请求
type SkipDayRequest struct {
	CurrentDay int `json:"current_day" validate:"required,min=1" description:"当前天数"`
}

// SkipDayResponse 跳过天数响应
type SkipDayResponse struct {
	NextDay int  `json:"next_day" description:"下一天数"`
	Poem    Poem `json:"poem" description:"下一首诗文"`
}

// CheckinResponse 打卡响应
type CheckinResponse struct {
	DayNumber    int     `json:"day_number" description:"第几天"`
	IsTodayFinish bool   `json:"is_today_finish" description:"今日是否完成"`
	CompletedDays int    `json:"completed_days" description:"已完成天数"`
	TotalDays     int    `json:"total_days" description:"总天数"`
	ProgressRate  float64 `json:"progress_rate" description:"完成率"`
}
