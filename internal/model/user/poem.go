package usermodel

// PoemResponse 诗歌响应
type PoemResponse struct {
	ID           int64    `json:"id" description:"诗歌ID"`
	Title        string   `json:"title" description:"诗歌标题"`
	Author       string   `json:"author" description:"作者"`
	Dynasty      string   `json:"dynasty" description:"朝代"`
	Content      string   `json:"content" description:"原文内容"`
	Translation  string   `json:"translation,omitempty" description:"现代译文"`
	Appreciation string   `json:"appreciation,omitempty" description:"赏析"`
	Category     string   `json:"category" description:"分类名称"`
	Tags         []string `json:"tags" description:"标签列表"`
	CoverURL     string   `json:"cover_url,omitempty" description:"封面图片URL"`
	IsFavorited  bool     `json:"is_favorited" description:"是否已收藏"`
	// 拼音字段
	TitlePinyin   string `json:"title_pinyin" description:"标题拼音（带声调）"`
	ContentPinyin string `json:"content_pinyin" description:"内容拼音（带声调）"`
	// 简体字段
	TitleSC        string `json:"title_sc" description:"标题（简体）"`
	AuthorSC       string `json:"author_sc" description:"作者（简体）"`
	ContentSC      string `json:"content_sc" description:"内容（简体）"`
	TranslationSC  string `json:"translation_sc,omitempty" description:"译文（简体）"`
	AppreciationSC string `json:"appreciation_sc,omitempty" description:"赏析（简体）"`
}

// PoemListItem 诗歌列表项
type PoemListItem struct {
	ID       int64  `json:"id" description:"诗歌ID"`
	Title    string `json:"title" description:"诗歌标题"`
	Author   string `json:"author" description:"作者"`
	Dynasty  string `json:"dynasty" description:"朝代"`
	Category string `json:"category" description:"分类名称"`
	CoverURL string `json:"cover_url" description:"封面图片URL"`
}

// PoemListResponse 诗歌列表响应
type PoemListResponse struct {
	Total int            `json:"total" description:"总数"`
	Items []PoemListItem `json:"items" description:"诗歌列表"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword  string `json:"keyword" validate:"required,min=1" description:"搜索关键词"`
	Page     int    `json:"page" validate:"min=1" description:"页码"`
	PageSize int    `json:"page_size" validate:"min=1,max=50" description:"每页数量"`
}

// SearchResponse 搜索响应（返回完整诗文数据）
type SearchResponse struct {
	Total int            `json:"total" description:"总数"`
	Items []PoemResponse `json:"items" description:"搜索结果列表"`
}
