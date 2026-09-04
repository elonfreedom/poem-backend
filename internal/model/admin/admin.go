package admin

import (
	"time"
)

// ========== 认证 ==========

// AdminLoginRequest 后台登录请求（适配 vben-admin，username 即邮箱）
type AdminLoginRequest struct {
	Username      string `json:"username" validate:"required" description:"管理员邮箱/用户名"`
	Password      string `json:"password" validate:"required,min=6,max=64" description:"密码（6-64个字符）"`
	SelectAccount string `json:"select_account,omitempty" description:"vben-admin 选择账户字段（忽略）"`
	Captcha       bool   `json:"captcha,omitempty" description:"vben-admin 验证码字段（忽略）"`
}

// AdminLoginResponse 后台登录响应（适配 vben-admin）
type AdminLoginResponse struct {
	AccessToken string            `json:"access_token" description:"JWT 认证令牌"`
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
	UserId   string          `json:"user_id" description:"用户ID"`
	Username string          `json:"username" description:"用户名"`
	RealName string          `json:"real_name" description:"真实姓名"`
	Avatar   string          `json:"avatar" description:"头像"`
	Desc     string          `json:"desc" description:"描述/角色"`
	HomePath string          `json:"home_path" description:"首页路径"`
	Roles    []AdminRoleInfo `json:"roles" description:"角色列表"`
}

type AdminRoleInfo struct {
	RoleName string `json:"role_name" description:"角色名称"`
	Value    string `json:"value" description:"角色值"`
}

// ========== 诗歌管理 ==========

// AdminPoemResponse 诗歌管理响应
type AdminPoemResponse struct {
	ID           int64     `json:"id" description:"诗歌ID"`
	Title        string    `json:"title" description:"诗歌标题"`
	Author       string    `json:"author" description:"作者"`
	Dynasty      string    `json:"dynasty,omitempty" description:"朝代"`
	Content      string    `json:"content" description:"原文内容"`
	Translation  string    `json:"translation,omitempty" description:"现代译文"`
	Appreciation string    `json:"appreciation,omitempty" description:"赏析"`
	Source       string    `json:"source,omitempty" description:"来源（如《唐诗三百首》）"`
	CategoryID   *int64    `json:"category_id,omitempty" description:"分类ID"`
	AuthorID     *int64    `json:"author_id,omitempty" description:"作者ID"`
	CategoryName string    `json:"category_name,omitempty" description:"分类名称"`
	Tags         []string  `json:"tags,omitempty" description:"标签列表"`
	CoverURL     string    `json:"cover_url,omitempty" description:"封面图片URL"`
	Status       string    `json:"status" description:"状态: draft, published, archived"`
	CreatedBy    *string   `json:"created_by,omitempty" description:"创建者ID"`
	CreatedAt    time.Time `json:"created_at" description:"创建时间"`
	UpdatedAt    time.Time `json:"updated_at" description:"更新时间"`
	// 拼音字段（可手动校正多音字）
	TitlePinyin   string `json:"title_pinyin" description:"标题拼音（带声调）"`
	ContentPinyin string `json:"content_pinyin" description:"内容拼音（带声调）"`
	// 简体字段（由繁体自动生成，可手动校正）
	TitleSC        string `json:"title_sc" description:"标题（简体）"`
	AuthorSC       string `json:"author_sc" description:"作者（简体）"`
	ContentSC      string `json:"content_sc" description:"内容（简体）"`
	TranslationSC  string `json:"translation_sc,omitempty" description:"译文（简体）"`
	AppreciationSC string `json:"appreciation_sc,omitempty" description:"赏析（简体）"`
}

// AdminPoemCreateRequest 创建诗歌请求
type AdminPoemCreateRequest struct {
	Title        string   `json:"title" validate:"required,max=200" description:"诗歌标题"`
	Author       string   `json:"author" validate:"required,max=50" description:"作者"`
	Dynasty      string   `json:"dynasty,omitempty" description:"朝代"`
	Content      string   `json:"content" validate:"required" description:"原文内容"`
	Translation  string   `json:"translation,omitempty" description:"现代译文"`
	Appreciation string   `json:"appreciation,omitempty" description:"赏析"`
	Source       string   `json:"source,omitempty" validate:"omitempty,max=200" description:"来源（如《唐诗三百首》）"`
	CategoryID   *int64   `json:"category_id,omitempty" description:"分类ID"`
	AuthorID     *int64   `json:"author_id,omitempty" description:"作者ID"`
	Tags         []string `json:"tags,omitempty" description:"标签列表"`
	CoverURL     string   `json:"cover_url,omitempty" description:"封面图片URL"`
	Status       string   `json:"status,omitempty" description:"状态: draft/published/archived"`
	// 拼音字段（可手动校正多音字，留空则自动生成）
	TitlePinyin   string `json:"title_pinyin,omitempty" description:"标题拼音（带声调）"`
	ContentPinyin string `json:"content_pinyin,omitempty" description:"内容拼音（带声调）"`
	// 简体字段（由繁体自动生成，可手动校正，留空则自动生成）
	TitleSC        string `json:"title_sc,omitempty" description:"标题（简体）"`
	AuthorSC       string `json:"author_sc,omitempty" description:"作者（简体）"`
	ContentSC      string `json:"content_sc,omitempty" description:"内容（简体）"`
	TranslationSC  string `json:"translation_sc,omitempty" description:"译文（简体）"`
	AppreciationSC string `json:"appreciation_sc,omitempty" description:"赏析（简体）"`
}

// AdminPoemUpdateRequest 更新诗歌请求
type AdminPoemUpdateRequest struct {
	Title        string   `json:"title" validate:"required,max=200" description:"诗歌标题"`
	Author       string   `json:"author" validate:"required,max=50" description:"作者"`
	Dynasty      string   `json:"dynasty,omitempty" description:"朝代"`
	Content      string   `json:"content" validate:"required" description:"原文内容"`
	Translation  string   `json:"translation,omitempty" description:"现代译文"`
	Appreciation string   `json:"appreciation,omitempty" description:"赏析"`
	Source       string   `json:"source,omitempty" validate:"omitempty,max=200" description:"来源（如《唐诗三百首》）"`
	CategoryID   *int64   `json:"category_id,omitempty" description:"分类ID"`
	AuthorID     *int64   `json:"author_id,omitempty" description:"作者ID"`
	Tags         []string `json:"tags,omitempty" description:"标签列表"`
	CoverURL     string   `json:"cover_url,omitempty" description:"封面图片URL"`
	Status       string   `json:"status,omitempty" description:"状态: draft/published/archived"`
	// 拼音字段（可手动校正多音字，留空则自动生成）
	TitlePinyin   string `json:"title_pinyin,omitempty" description:"标题拼音（带声调）"`
	ContentPinyin string `json:"content_pinyin,omitempty" description:"内容拼音（带声调）"`
	// 简体字段（由繁体自动生成，可手动校正，留空则自动生成）
	TitleSC        string `json:"title_sc,omitempty" description:"标题（简体）"`
	AuthorSC       string `json:"author_sc,omitempty" description:"作者（简体）"`
	ContentSC      string `json:"content_sc,omitempty" description:"内容（简体）"`
	TranslationSC  string `json:"translation_sc,omitempty" description:"译文（简体）"`
	AppreciationSC string `json:"appreciation_sc,omitempty" description:"赏析（简体）"`
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
	PoemID    int64  `json:"poem_id" description:"诗歌ID"`
	Title     string `json:"title" description:"诗歌标题"`
	Author    string `json:"author" description:"作者"`
	ViewCount int64  `json:"view_count" description:"浏览次数"`
}

// AdminStatsUserGrowth 用户增长
type AdminStatsUserGrowth struct {
	Date       string `json:"date" description:"日期"`
	NewUsers   int64  `json:"new_users" description:"新增用户"`
	TotalUsers int64  `json:"total_users" description:"累计用户"`
}

// ========== Banner 管理 ==========

// AdminBannerResponse Banner 响应
type AdminBannerResponse struct {
	ID        int64     `json:"id" description:"Banner ID"`
	Title     string    `json:"title" description:"标题"`
	ImageURL  string    `json:"image_url" description:"图片URL"`
	LinkType  string    `json:"link_type" description:"链接类型: poem, url"`
	LinkValue string    `json:"link_value" description:"链接值"`
	Sort      int       `json:"sort" description:"排序值"`
	Status    string    `json:"status" description:"状态: active, inactive"`
	CreatedAt time.Time `json:"created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" description:"更新时间"`
}

// AdminBannerCreateRequest 创建 Banner 请求
type AdminBannerCreateRequest struct {
	Title     string `json:"title" validate:"required,max=200" description:"标题"`
	ImageURL  string `json:"image_url" validate:"required,max=500" description:"图片URL"`
	LinkType  string `json:"link_type" validate:"required,oneof=poem url" description:"链接类型"`
	LinkValue string `json:"link_value" validate:"required,max=500" description:"链接值"`
	Sort      int    `json:"sort,omitempty" description:"排序值"`
	Status    string `json:"status,omitempty" validate:"omitempty,oneof=active inactive" description:"状态"`
}

// AdminBannerUpdateRequest 更新 Banner 请求
type AdminBannerUpdateRequest struct {
	Title     string `json:"title" validate:"required,max=200" description:"标题"`
	ImageURL  string `json:"image_url" validate:"required,max=500" description:"图片URL"`
	LinkType  string `json:"link_type" validate:"required,oneof=poem url" description:"链接类型"`
	LinkValue string `json:"link_value" validate:"required,max=500" description:"链接值"`
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

// ========== 作者管理 ==========

// AdminAuthorResponse 作者响应
type AdminAuthorResponse struct {
	ID              int64     `json:"id" description:"作者ID"`
	Name            string    `json:"name" description:"作者名（简体）"`
	NameTraditional string    `json:"name_traditional" description:"作者名（繁体）"`
	Dynasty         string    `json:"dynasty" description:"朝代"`
	Biography       string    `json:"biography" description:"作者简介"`
	PoemCount       int64     `json:"poem_count" description:"关联诗歌数量"`
	CreatedAt       time.Time `json:"created_at" description:"创建时间"`
	UpdatedAt       time.Time `json:"updated_at" description:"更新时间"`
}

// AdminAuthorCreateRequest 创建作者请求
type AdminAuthorCreateRequest struct {
	Name            string `json:"name" validate:"required,max=100" description:"作者名（简体）"`
	NameTraditional string `json:"name_traditional,omitempty" description:"作者名（繁体）"`
	Dynasty         string `json:"dynasty,omitempty" description:"朝代"`
	Biography       string `json:"biography,omitempty" description:"作者简介"`
}

// AdminAuthorUpdateRequest 更新作者请求
type AdminAuthorUpdateRequest struct {
	Name            string `json:"name" validate:"required,max=100" description:"作者名（简体）"`
	NameTraditional string `json:"name_traditional,omitempty" description:"作者名（繁体）"`
	Dynasty         string `json:"dynasty,omitempty" description:"朝代"`
	Biography       string `json:"biography,omitempty" description:"作者简介"`
}

// AdminAuthorOptionResponse 作者下拉选项（简化结构，用于搜索框）
type AdminAuthorOptionResponse struct {
	ID      int64  `json:"id" description:"作者ID"`
	Name    string `json:"name" description:"作者名"`
	Dynasty string `json:"dynasty" description:"朝代"`
}

// AdminAuthorBatchMatchRequest 批量匹配诗歌请求
// poetry_ids 为空数组时处理全部诗歌，非空时只处理指定 ID
type AdminAuthorBatchMatchRequest struct {
	PoetryIDs []int64 `json:"poetry_ids" description:"诗歌ID数组（空数组=全部诗歌）"`
}

// AdminAuthorBatchMatchResponse 批量匹配结果
type AdminAuthorBatchMatchResponse struct {
	Total     int64 `json:"total" description:"总条数"`
	Matched   int64 `json:"matched" description:"成功匹配数"`
	Unmatched int64 `json:"unmatched" description:"未匹配数"`
}

// ========== 打卡管理 ==========

// AdminCheckinListItem 打卡记录列表项
type AdminCheckinListItem struct {
	ID              int64     `json:"id" description:"记录序号"`
	UserID          string    `json:"user_id" description:"用户ID"`
	Nickname        string    `json:"nickname" description:"用户昵称"`
	CheckinDate     string    `json:"checkin_date" description:"打卡日期"`
	PoemID          *int64    `json:"poem_id,omitempty" description:"关联诗歌ID"`
	PoemTitle       *string   `json:"poem_title,omitempty" description:"诗歌标题"`
	ConsecutiveDays int       `json:"consecutive_days" description:"连续打卡天数"`
	CreatedAt       time.Time `json:"created_at" description:"打卡时间"`
}

// AdminCheckinListResponse 打卡记录列表响应
type AdminCheckinListResponse struct {
	Items []AdminCheckinListItem `json:"items" description:"打卡记录列表"`
	Total int64                  `json:"total" description:"总条数"`
}

// AdminCheckinHotPoem 打卡热门诗歌
type AdminCheckinHotPoem struct {
	PoemID       int64  `json:"poem_id" description:"诗歌ID"`
	PoemTitle    string `json:"poem_title" description:"诗歌标题"`
	CheckinCount int64  `json:"checkin_count" description:"打卡次数"`
}

// AdminCheckinStats 打卡数据统计响应
type AdminCheckinStats struct {
	DailyAvgRate  float64             `json:"daily_avg_rate" description:"日均打卡率"`
	Retention7d   float64             `json:"retention_7d" description:"7日留存率"`
	TotalCheckins int64               `json:"total_checkins" description:"总打卡次数"`
	TotalUsers    int64               `json:"total_users" description:"总打卡用户数"`
	HotPoems      []AdminCheckinHotPoem `json:"hot_poems" description:"热门诗歌TOP10"`
}

// ========== 工具模块 ==========

// AdminToolGenerateAuthorsResponse 从诗歌提取作者工具响应
type AdminToolGenerateAuthorsResponse struct {
	TotalUnique  int `json:"total_unique" description:"诗歌中唯一作者数"`
	Created      int `json:"created" description:"新建作者数"`
	Skipped      int `json:"skipped" description:"已存在跳过数"`
	WithDynasty  int `json:"with_dynasty" description:"提取时附带朝代信息的作者数"`
	Backfilled   int `json:"backfilled" description:"回填朝代信息的已有作者数"`
}

// ========== 作者查重工具 ==========

// AdminToolAuthorDedupScanRequest 扫描重复作者请求
type AdminToolAuthorDedupScanRequest struct {
	MatchBy string `query:"match_by" description:"匹配方式: name(仅姓名) 或 name_dynasty(姓名+朝代)"`
}

// AdminToolAuthorDedupScanResponse 扫描重复作者响应
type AdminToolAuthorDedupScanResponse struct {
	TotalScanned int                       `json:"total_scanned" description:"扫描的作者总数"`
	TotalGroups  int                       `json:"total_groups" description:"重复组总数"`
	Groups       []AdminToolAuthorDedupGroup `json:"groups" description:"重复组列表"`
}

// AdminToolAuthorDedupGroup 重复作者组
type AdminToolAuthorDedupGroup struct {
	GroupKey     string                       `json:"group_key" description:"组标识（匹配键）"`
	MatchReason  string                       `json:"match_reason" description:"匹配原因"`
	AuthorCount  int                          `json:"author_count" description:"组内作者数"`
	Authors      []AdminToolAuthorDedupItem   `json:"authors" description:"组内作者列表"`
}

// AdminToolAuthorDedupItem 重复组内的作者
type AdminToolAuthorDedupItem struct {
	ID              int64  `json:"id" description:"作者ID"`
	Name            string `json:"name" description:"作者名"`
	Dynasty         string `json:"dynasty" description:"朝代"`
	Biography       string `json:"biography" description:"作者简介"`
	PoemCount       int64  `json:"poem_count" description:"关联诗歌数量"`
}

// AdminToolAuthorDedupMergeRequest 合并重复作者请求
type AdminToolAuthorDedupMergeRequest struct {
	KeepID   int64   `json:"keep_id" validate:"required" description:"保留的作者ID"`
	MergeIDs []int64 `json:"merge_ids" validate:"required,min=1" description:"要合并的作者ID（合并后删除）"`
}

// AdminToolAuthorDedupMergeResponse 合并重复作者响应
type AdminToolAuthorDedupMergeResponse struct {
	KeepID          int64  `json:"keep_id" description:"保留的作者ID"`
	Merged          int    `json:"merged" description:"合并的作者数"`
	ReassignedPoems int64  `json:"reassigned_poems" description:"重新关联的诗歌数"`
	Message         string `json:"message" description:"处理结果描述"`
}

// ========== 简繁体工具 ==========

// AdminToolDetectCharsTypeRequest 检测字符类型请求
type AdminToolDetectCharsTypeRequest struct {
	Text string `json:"text" validate:"required" description:"待检测的文本"`
}

// AdminToolDetectCharsTypeResponse 检测字符类型响应
type AdminToolDetectCharsTypeResponse struct {
	Type string `json:"type" description:"字符类型: simplified, traditional, mixed, unknown"`
}

// AdminToolConvertCharsRequest 字符转换请求
type AdminToolConvertCharsRequest struct {
	Text   string `json:"text" validate:"required" description:"待转换的文本"`
	Target string `json:"target" validate:"required,oneof=simplified traditional" description="目标类型: simplified 或 traditional"`
}

// AdminToolConvertCharsResponse 字符转换响应
type AdminToolConvertCharsResponse struct {
	Text string `json:"text" description:"转换后的文本"`
}

// AdminToolBatchConvertCharsRequest 批量转换字符请求
type AdminToolBatchConvertCharsRequest struct {
	PoetryIDs []int64 `json:"poetry_ids" validate:"required,min=1" description:"诗歌ID数组"`
	Target    string  `json:"target" validate:"required,oneof=simplified traditional" description="目标类型: simplified 或 traditional"`
}

// AdminToolBatchConvertCharsResponse 批量转换字符响应
type AdminToolBatchConvertCharsResponse struct {
	Total     int    `json:"total" description:"请求处理的诗歌数"`
	Converted int    `json:"converted" description:"成功转换的诗歌数"`
	Message   string `json:"message" description:"处理结果描述"`
}

// ========== 诗文去重工具 ==========

// AdminToolDedupScanRequest 扫描重复组请求
type AdminToolDedupScanRequest struct {
	MatchFields   []string `json:"match_fields" validate:"required,min=1,dive,oneof=title author content" description:"匹配维度 title/author/content"`
	StatusFilter  string   `json:"status_filter,omitempty" description:"按状态筛选 published/draft/archived/non_archived(排除已归档)"`
	DynastyFilter string   `json:"dynasty_filter,omitempty" description:"按朝代筛选 如 唐"`
	Page          int      `json:"page" description:"页码 默认1"`
	PageSize      int      `json:"page_size" description:"每页组数 默认20 最大500"`
}

// AdminToolDedupScanResponse 扫描重复组响应
type AdminToolDedupScanResponse struct {
	TotalScanned    int                   `json:"total_scanned" description:"扫描的诗文总数"`
	TotalGroups     int                   `json:"total_groups" description:"重复组总数"`
	TotalDuplicates int                   `json:"total_duplicates" description:"重复诗文总数（不含每组保留的 1 首）"`
	Page            int                   `json:"page" description:"当前页码"`
	PageSize        int                   `json:"page_size" description:"每页组数"`
	Groups          []AdminToolDedupGroup `json:"groups" description:"当前页重复组列表"`
}

// AdminToolDedupGroup 重复组
type AdminToolDedupGroup struct {
	GroupID           string               `json:"group_id" description:"组标识（hash）"`
	MatchReason       string               `json:"match_reason" description:"匹配原因（如：标题+作者相同）"`
	MatchKey          string               `json:"match_key" description:"匹配键（如：静夜思 - 李白）"`
	PoemCount         int64                `json:"poem_count" description:"组内诗文数量"`
	Poems             []AdminToolDedupPoem `json:"poems" description:"组内诗文列表"`
	RecommendedKeepID int64                `json:"recommended_keep_id" description:"推荐保留的诗文 ID"`
}

// AdminToolDedupPoem 去重工具中的诗文信息
type AdminToolDedupPoem struct {
	ID           int64     `json:"id" description:"诗歌ID"`
	Title        string    `json:"title" description:"标题"`
	TitleSC      string    `json:"title_sc" description:"标题（简体）"`
	Author       string    `json:"author" description:"作者"`
	AuthorSC     string    `json:"author_sc" description:"作者（简体）"`
	Dynasty      string    `json:"dynasty" description:"朝代"`
	Content      string    `json:"content" description:"内容"`
	ContentSC    string    `json:"content_sc" description:"内容（简体）"`
	Translation  string    `json:"translation" description:"译文"`
	Appreciation string    `json:"appreciation" description:"赏析"`
	CategoryID   *int64    `json:"category_id,omitempty" description:"分类ID"`
	CategoryName string    `json:"category_name,omitempty" description:"分类名称"`
	Tags         []string  `json:"tags" description:"标签列表"`
	Status       string    `json:"status" description:"状态"`
	CreatedAt    time.Time `json:"created_at" description:"创建时间"`
	UpdatedAt    time.Time `json:"updated_at" description:"更新时间"`
}

// AdminToolDedupExecuteRequest 执行去重请求
type AdminToolDedupExecuteRequest struct {
	ArchiveIDs []int64 `json:"archive_ids" description:"需要归档的诗文 ID 数组"`
	DeleteIDs  []int64 `json:"delete_ids" description:"需要删除的诗文 ID 数组"`
}

// AdminToolDedupExecuteResponse 执行去重响应
type AdminToolDedupExecuteResponse struct {
	Archived int    `json:"archived" description:"归档数量"`
	Deleted  int    `json:"deleted" description:"删除数量"`
	Message  string `json:"message" description:"处理结果描述"`
}

// AdminToolDedupMergeRequest 合并重复诗文请求
type AdminToolDedupMergeRequest struct {
	KeepID    int64   `json:"keep_id" validate:"required" description:"保留的诗 ID"`
	MergeIDs  []int64 `json:"merge_ids" validate:"required,min=1" description:"要合并的诗 ID（合并后会被归档）"`
}

// AdminToolDedupMergeResponse 合并重复诗文响应
type AdminToolDedupMergeResponse struct {
	KeepID       int64    `json:"keep_id" description:"保留的诗 ID"`
	MergedFields []string `json:"merged_fields" description:"实际合并的字段列表"`
	Archived     int      `json:"archived" description:"归档的诗文数量"`
	Message      string   `json:"message" description:"处理结果描述"`
}
