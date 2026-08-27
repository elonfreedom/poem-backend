package admin

import (
	"time"
)

// ========== 认证 ==========

// AdminLoginRequest 后台登录请求（适配 vben-admin，username 即邮箱）
type AdminLoginRequest struct {
	Username      string `json:"username" validate:"required" description:"管理员邮箱/用户名"`
	Password      string `json:"password" validate:"required,min=6,max=64" description:"密码（6-64个字符）"`
	SelectAccount string `json:"selectAccount,omitempty" description:"vben-admin 选择账户字段（忽略）"`
	Captcha       bool   `json:"captcha,omitempty" description:"vben-admin 验证码字段（忽略）"`
}

// AdminLoginResponse 后台登录响应（适配 vben-admin）
type AdminLoginResponse struct {
	AccessToken string            `json:"accessToken" description:"JWT 认证令牌"`
	User        AdminUserResponse `json:"user" description:"用户信息"`
}

// AdminUserResponse 后台用户响应
type AdminUserResponse struct {
	ID        string    `json:"id" description:"用户唯一标识"`
	Nickname  string    `json:"nickname" description:"用户昵称"`
	Email     string    `json:"email,omitempty" description:"邮箱地址（脱敏显示）"`
	Role      string    `json:"role" description:"用户角色"`
	CreatedAt time.Time `json:"created_at" description:"创建时间"`
}

// AdminUserInfoResponse 用户信息响应（适配 vben-admin /user/info）
type AdminUserInfoResponse struct {
	UserId   string          `json:"userId" description:"用户ID"`
	Username string          `json:"username" description:"用户名"`
	RealName string          `json:"realName" description:"真实姓名"`
	Avatar   string          `json:"avatar" description:"头像"`
	Desc     string          `json:"desc" description:"描述/角色"`
	HomePath string          `json:"homePath" description:"首页路径"`
	Roles    []AdminRoleInfo `json:"roles" description:"角色列表"`
}

type AdminRoleInfo struct {
	RoleName string `json:"roleName" description:"角色名称"`
	Value    string `json:"value" description:"角色值"`
}

// ========== 诗歌管理 ==========

// AdminPoemResponse 诗歌管理响应
type AdminPoemResponse struct {
	ID            int64     `json:"id" description:"诗歌ID"`
	Title         string    `json:"title" description:"诗歌标题"`
	Author        string    `json:"author" description:"作者"`
	Dynasty       string    `json:"dynasty,omitempty" description:"朝代"`
	Content       string    `json:"content" description:"原文内容"`
	Translation   string    `json:"translation,omitempty" description:"现代译文"`
	Appreciation  string    `json:"appreciation,omitempty" description:"赏析"`
	Source        string    `json:"source,omitempty" description:"来源（如《唐诗三百首》）"`
	CategoryID    *int64    `json:"category_id,omitempty" description:"分类ID"`
	CategoryName  string    `json:"category_name,omitempty" description:"分类名称"`
	Tags          []string  `json:"tags,omitempty" description:"标签列表"`
	CoverURL      string    `json:"cover_url,omitempty" description:"封面图片URL"`
	Status        string    `json:"status" description:"状态: draft, published, archived"`
	CreatedBy     *string   `json:"created_by,omitempty" description:"创建者ID"`
	CreatedAt     time.Time `json:"created_at" description:"创建时间"`
	UpdatedAt     time.Time `json:"updated_at" description:"更新时间"`
	// 拼音字段（可手动校正多音字）
	TitlePinyin   string `json:"title_pinyin" description:"标题拼音（带声调）"`
	ContentPinyin string `json:"content_pinyin" description:"内容拼音（带声调）"`
	AuthorPinyin  string `json:"author_pinyin" description:"作者拼音（带声调）"`
}

// AdminPoemCreateRequest 创建诗歌请求
type AdminPoemCreateRequest struct {
	Title         string   `json:"title" validate:"required,max=100" description:"诗歌标题"`
	Author        string   `json:"author" validate:"required,max=50" description:"作者"`
	Dynasty       string   `json:"dynasty,omitempty" description:"朝代"`
	Content       string   `json:"content" validate:"required" description:"原文内容"`
	Translation   string   `json:"translation,omitempty" description:"现代译文"`
	Appreciation  string   `json:"appreciation,omitempty" description:"赏析"`
	Source        string   `json:"source,omitempty" validate:"omitempty,max=200" description:"来源（如《唐诗三百首》）"`
	CategoryID    *int64   `json:"category_id,omitempty" description:"分类ID"`
	Tags          []string `json:"tags,omitempty" description:"标签列表"`
	CoverURL      string   `json:"cover_url,omitempty" description:"封面图片URL"`
	Status        string   `json:"status,omitempty" description:"状态: draft/published/archived"`
	// 拼音字段（可手动校正多音字，留空则自动生成）
	TitlePinyin   string `json:"title_pinyin,omitempty" description:"标题拼音（带声调）"`
	ContentPinyin string `json:"content_pinyin,omitempty" description:"内容拼音（带声调）"`
	AuthorPinyin  string `json:"author_pinyin,omitempty" description:"作者拼音（带声调）"`
}

// AdminPoemUpdateRequest 更新诗歌请求
type AdminPoemUpdateRequest struct {
	Title         string   `json:"title" validate:"required,max=100" description:"诗歌标题"`
	Author        string   `json:"author" validate:"required,max=50" description:"作者"`
	Dynasty       string   `json:"dynasty,omitempty" description:"朝代"`
	Content       string   `json:"content" validate:"required" description:"原文内容"`
	Translation   string   `json:"translation,omitempty" description:"现代译文"`
	Appreciation  string   `json:"appreciation,omitempty" description:"赏析"`
	Source        string   `json:"source,omitempty" validate:"omitempty,max=200" description:"来源（如《唐诗三百首》）"`
	CategoryID    *int64   `json:"category_id,omitempty" description:"分类ID"`
	Tags          []string `json:"tags,omitempty" description:"标签列表"`
	CoverURL      string   `json:"cover_url,omitempty" description:"封面图片URL"`
	Status        string   `json:"status,omitempty" description:"状态: draft/published/archived"`
	// 拼音字段（可手动校正多音字，留空则自动生成）
	TitlePinyin   string `json:"title_pinyin,omitempty" description:"标题拼音（带声调）"`
	ContentPinyin string `json:"content_pinyin,omitempty" description:"内容拼音（带声调）"`
	AuthorPinyin  string `json:"author_pinyin,omitempty" description:"作者拼音（带声调）"`
}

// AdminPoemUpdateStatusRequest 更新诗歌状态请求
type AdminPoemUpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft published archived" description:"状态"`
}

// AdminPoemBatchUpdateStatusRequest 批量更新诗歌状态请求
type AdminPoemBatchUpdateStatusRequest struct {
	IDs    []int64 `json:"ids" validate:"required,min=1" description:"诗歌ID数组"`
	Status string  `json:"status" validate:"required,oneof=draft published archived" description:"目标状态"`
}

// ========== 分类管理 ==========

// AdminCategoryResponse 分类响应
type AdminCategoryResponse struct {
	ID        int64     `json:"id" description:"分类ID"`
	Name      string    `json:"name" description:"分类名称"`
	Sort      int       `json:"sort" description:"排序值"`
	PoemCount int64     `json:"poem_count" description:"诗歌数量"`
	CreatedAt time.Time `json:"created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" description:"更新时间"`
}

// AdminCategoryCreateRequest 创建分类请求
type AdminCategoryCreateRequest struct {
	Name string `json:"name" validate:"required,max=50" description:"分类名称"`
	Sort int    `json:"sort,omitempty" description:"排序值"`
}

// AdminCategoryUpdateRequest 更新分类请求
type AdminCategoryUpdateRequest struct {
	Name string `json:"name" validate:"required,max=50" description:"分类名称"`
	Sort int    `json:"sort,omitempty" description:"排序值"`
}

// ========== 标签管理 ==========

// AdminTagResponse 标签响应
type AdminTagResponse struct {
	ID        int64     `json:"id" description:"标签ID"`
	Name      string    `json:"name" description:"标签名称"`
	CreatedAt time.Time `json:"created_at" description:"创建时间"`
}

// AdminTagCreateRequest 创建标签请求
type AdminTagCreateRequest struct {
	Name string `json:"name" validate:"required,max=50" description:"标签名称"`
}

// ========== 数据统计 ==========

// AdminStatsOverview 总览统计
type AdminStatsOverview struct {
	TotalUsers   int64 `json:"total_users" description:"用户总数"`
	TotalPoems   int64 `json:"total_poems" description:"诗歌总数"`
	TotalViews   int64 `json:"total_views" description:"总浏览量"`
	TodayActive  int64 `json:"today_active" description:"今日活跃"`
	TodayCheckin int64 `json:"today_checkin" description:"今日打卡"`
}

// AdminStatsDaily 每日统计
type AdminStatsDaily struct {
	Date  string `json:"date" description:"日期"`
	Views int64  `json:"views" description:"浏览量"`
	Users int64  `json:"users" description:"新增用户"`
}

// AdminStatsHotPoem 热门诗歌
type AdminStatsHotPoem struct {
	PoemID    int64  `json:"poemId" description:"诗歌ID"`
	Title     string `json:"title" description:"诗歌标题"`
	Author    string `json:"author" description:"作者"`
	ViewCount int64  `json:"viewCount" description:"浏览次数"`
}

// AdminStatsUserGrowth 用户增长
type AdminStatsUserGrowth struct {
	Date       string `json:"date" description:"日期"`
	NewUsers   int64  `json:"newUsers" description:"新增用户"`
	TotalUsers int64  `json:"totalUsers" description:"累计用户"`
}

// ========== Banner 管理 ==========

// AdminBannerResponse Banner 响应
type AdminBannerResponse struct {
	ID        int64     `json:"id" description:"Banner ID"`
	Title     string    `json:"title" description:"标题"`
	ImageURL  string    `json:"imageUrl" description:"图片URL"`
	LinkType  string    `json:"linkType" description:"链接类型: poem, url"`
	LinkValue string    `json:"linkValue" description:"链接值"`
	Sort      int       `json:"sort" description:"排序值"`
	Status    string    `json:"status" description:"状态: active, inactive"`
	CreatedAt time.Time `json:"created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" description:"更新时间"`
}

// AdminBannerCreateRequest 创建 Banner 请求
type AdminBannerCreateRequest struct {
	Title     string `json:"title" validate:"required,max=100" description:"标题"`
	ImageURL  string `json:"imageUrl" validate:"required,max=500" description:"图片URL"`
	LinkType  string `json:"linkType" validate:"required,oneof=poem url" description:"链接类型"`
	LinkValue string `json:"linkValue" validate:"required,max=500" description:"链接值"`
	Sort      int    `json:"sort,omitempty" description:"排序值"`
	Status    string `json:"status,omitempty" validate:"omitempty,oneof=active inactive" description:"状态"`
}

// AdminBannerUpdateRequest 更新 Banner 请求
type AdminBannerUpdateRequest struct {
	Title     string `json:"title" validate:"required,max=100" description:"标题"`
	ImageURL  string `json:"imageUrl" validate:"required,max=500" description:"图片URL"`
	LinkType  string `json:"linkType" validate:"required,oneof=poem url" description:"链接类型"`
	LinkValue string `json:"linkValue" validate:"required,max=500" description:"链接值"`
	Sort      int    `json:"sort,omitempty" description:"排序值"`
	Status    string `json:"status,omitempty" validate:"omitempty,oneof=active inactive" description:"状态"`
}

// ========== 公告管理 ==========

// AdminAnnouncementResponse 公告响应
type AdminAnnouncementResponse struct {
	ID        int64     `json:"id" description:"公告ID"`
	Title     string    `json:"title" description:"标题"`
	Content   string    `json:"content" description:"内容"`
	Status    string    `json:"status" description:"状态: draft, published"`
	CreatedAt time.Time `json:"created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" description:"更新时间"`
}

// AdminAnnouncementCreateRequest 创建公告请求
type AdminAnnouncementCreateRequest struct {
	Title   string `json:"title" validate:"required,max=200" description:"标题"`
	Content string `json:"content" validate:"required" description:"内容"`
	Status  string `json:"status,omitempty" validate:"omitempty,oneof=draft published" description:"状态"`
}

// AdminAnnouncementUpdateRequest 更新公告请求
type AdminAnnouncementUpdateRequest struct {
	Title   string `json:"title" validate:"required,max=200" description:"标题"`
	Content string `json:"content" validate:"required" description:"内容"`
	Status  string `json:"status,omitempty" validate:"omitempty,oneof=draft published" description:"状态"`
}

// ========== 系统配置 ==========

// AdminConfigResponse 系统配置响应
type AdminConfigResponse struct {
	Key       string    `json:"key" description:"配置键"`
	Value     string    `json:"value" description:"配置值"`
	Remark    string    `json:"remark,omitempty" description:"备注"`
	UpdatedAt time.Time `json:"updated_at" description:"更新时间"`
}

// AdminConfigUpdateRequest 更新配置请求
type AdminConfigUpdateRequest struct {
	Value  string `json:"value" validate:"required" description:"配置值"`
	Remark string `json:"remark,omitempty" description:"备注"`
}
