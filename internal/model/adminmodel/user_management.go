package adminmodel

import "time"

// AdminUserListResponse 用户列表项
type AdminUserListItem struct {
	ID        string    `json:"id" description:"用户ID"`
	Nickname  string    `json:"nickname" description:"昵称"`
	Email     string    `json:"email,omitempty" description:"邮箱（脱敏）"`
	Role      string    `json:"role" description:"角色"`
	Status    string    `json:"status" description:"状态"`
	CreatedAt time.Time `json:"created_at" description:"注册时间"`
}

// AdminUserDetailResponse 用户详情（含统计数据）
type AdminUserDetailResponse struct {
	ID        string    `json:"id" description:"用户ID"`
	Nickname  string    `json:"nickname" description:"昵称"`
	Email     string    `json:"email,omitempty" description:"邮箱（脱敏）"`
	Role      string    `json:"role" description:"角色"`
	Status    string    `json:"status" description:"状态"`
	CreatedAt time.Time `json:"created_at" description:"注册时间"`
	UpdatedAt time.Time `json:"updated_at" description:"更新时间"`
	// 统计数据
	TotalCheckinDays int `json:"total_checkin_days" description:"累计打卡天数"`
	ConsecutiveDays  int `json:"consecutive_days" description:"当前连续天数"`
	FavoriteCount    int `json:"favorite_count" description:"收藏数量"`
	ReadingPlanCount int `json:"reading_plan_count" description:"阅读计划数量"`
	PasskeyCount     int `json:"passkey_count" description:"Passkey 数量"`
}

// AdminUserUpdateStatusRequest 更新用户状态请求
type AdminUserUpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active disabled" description:"用户状态: active, disabled"`
}

// AdminUserListRequest 用户列表请求参数
type AdminUserListRequest struct {
	Page     int    `json:"page" description:"页码"`
	PageSize int    `json:"page_size" description:"每页数量"`
	Keyword  string `json:"keyword" description:"搜索关键词（昵称/邮箱）"`
	Status   string `json:"status" description:"状态筛选"`
}
