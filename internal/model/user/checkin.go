package usermodel

import "time"

// CheckIn 打卡记录
type CheckIn struct {
	ID             int64     `json:"id" db:"id" description:"打卡记录ID"`
	UserID         string    `json:"user_id" db:"user_id" description:"用户ID (UUID v7)"`
	Nickname       string    `json:"nickname,omitempty" db:"-" description:"用户昵称（查询时 JOIN users 表填充，不存储）"`
	Date           time.Time `json:"date" db:"date" description:"打卡日期"`
	SubscriptionID *int64    `json:"subscription_id,omitempty" db:"subscription_id" description:"订阅ID（订阅打卡时非空）"`
	DayNumber      *int      `json:"day_number,omitempty" db:"day_number" description:"第几天（订阅打卡时非空）"`
	ConsecutiveDay int       `json:"consecutive_day" db:"consecutive_day" description:"连续打卡天数"`
	PoemID         *int64    `json:"poem_id" db:"poem_id" description:"关联诗歌ID"`
	PoemIDs        []int64   `json:"poem_ids,omitempty" db:"poem_ids" description:"关联诗歌ID列表（订阅打卡，支持多首）"`
	PoemTitle      *string   `json:"poem_title,omitempty" db:"-" description:"诗歌标题（查询时 JOIN poems 表填充，不存储）"`
	CreatedAt      time.Time `json:"created_at" db:"created_at" description:"创建时间"`
}

// PoemIDOrFirst 返回 poem_id 或 poem_ids[0]（兼容单诗/多诗）
func (c *CheckIn) PoemIDOrFirst() *int64 {
	if c.PoemID != nil {
		return c.PoemID
	}
	if len(c.PoemIDs) > 0 {
		return &c.PoemIDs[0]
	}
	return nil
}

// CheckInStats 打卡统计
type CheckInStats struct {
	UserID         string    `json:"user_id" db:"user_id" description:"用户ID (UUID v7)"`
	TotalDays      int       `json:"total_days" db:"total_days" description:"累计打卡天数"`
	ConsecutiveDay int       `json:"consecutive_day" db:"consecutive_day" description:"当前连续天数"`
	MaxConsecutive int       `json:"max_consecutive" db:"max_consecutive" description:"最大连续天数"`
	LastCheckIn    time.Time `json:"last_check_in" db:"last_check_in" description:"最后打卡时间"`
}

// CheckInResponse 打卡响应
type CheckInResponse struct {
	Date           time.Time `json:"date" description:"打卡日期"`
	ConsecutiveDay int       `json:"consecutive_day" description:"连续打卡天数"`
	PoemID         *int64    `json:"poem_id,omitempty" description:"关联诗歌ID"`
	PoemTitle      *string   `json:"poem_title,omitempty" description:"诗歌标题（热力图 hover 展示）"`
}

// CheckInStatsResponse 打卡统计响应
type CheckInStatsResponse struct {
	TotalDays      int       `json:"total_days" description:"累计打卡天数"`
	ConsecutiveDay int       `json:"consecutive_day" description:"当前连续天数"`
	MaxConsecutive int       `json:"max_consecutive" description:"最大连续天数"`
	LastCheckIn    time.Time `json:"last_check_in" description:"最后打卡时间"`
}

// CheckInCalendarResponse 打卡日历响应
type CheckInCalendarResponse struct {
	Year  int           `json:"year" description:"年"`
	Month int           `json:"month" description:"月"`
	Days  []CalendarDay `json:"days" description:"每日打卡情况"`
}

// CalendarDay 日历天
type CalendarDay struct {
	Day       int  `json:"day" description:"日"`
	IsChecked bool `json:"is_checked" description:"是否已打卡"`
}

// RankingItem 排行榜项
type RankingItem struct {
	Rank           int    `json:"rank" description:"排名"`
	Nickname       string `json:"nickname" description:"昵称"`
	ConsecutiveDay int    `json:"consecutive_day" description:"连续打卡天数"`
}

// RankingResponse 排行榜响应
type RankingResponse struct {
	Total            int           `json:"total" description:"总人数"`
	MyRank           int           `json:"my_rank" description:"我的排名"`
	MyConsecutiveDay int           `json:"my_consecutive_day" description:"我的连续天数"`
	List             []RankingItem `json:"list" description:"排行榜列表"`
}

// CheckInListResponse 打卡记录列表响应
type CheckInListResponse struct {
	Total int               `json:"total" description:"总数"`
	List  []CheckInResponse `json:"items" description:"打卡记录列表"`
}
