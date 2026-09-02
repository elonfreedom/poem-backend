package admin

import "time"

// ========== 导入记录 ==========

// ImportRecord 导入记录实体（数据库模型）
type ImportRecord struct {
	ID        int64           `json:"id" db:"id"`
	FileName  string          `json:"file_name" db:"file_name"`
	Source    string          `json:"source" db:"source"`
	Total     int             `json:"total" db:"total"`
	Processed int             `json:"processed" db:"processed"`
	Success   int             `json:"success" db:"success"`
	Failed    int             `json:"failed" db:"failed"`
	Status    string          `json:"status" db:"status"`
	Errors    []ImportError   `json:"errors" db:"errors"`
	CreatedBy *string         `json:"created_by" db:"created_by"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// ImportError 导入错误详情
type ImportError struct {
	Index int    `json:"index" description:"行号"`
	Title string `json:"title" description:"诗歌标题"`
	Error string `json:"error" description:"错误原因"`
}

// ========== 请求/响应 ==========

// ImportRecordListRequest 导入记录列表查询
type ImportRecordListRequest struct {
	Page      int    `query:"page" description:"页码"`
	PageSize  int    `query:"page_size" description:"每页数量"`
	Status    string `query:"status" description:"状态筛选: success/partial/failed"`
	StartDate string `query:"start_date" description:"起始日期 YYYY-MM-DD"`
	EndDate   string `query:"end_date" description:"结束日期 YYYY-MM-DD"`
}

// ImportRecordStatsResponse 统计响应
type ImportRecordStatsResponse struct {
	TotalImports int     `json:"total_imports" description:"总导入次数"`
	TotalPoems   int     `json:"total_poems" description:"总导入诗文数"`
	TotalSuccess int     `json:"total_success" description:"总成功数"`
	TotalFailed  int     `json:"total_failed" description:"总失败数"`
	SuccessRate  float64 `json:"success_rate" description:"成功率(%)"`
}
