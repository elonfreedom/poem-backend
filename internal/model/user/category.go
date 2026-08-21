package usermodel

// CategoryResponse 分类响应
type CategoryResponse struct {
	ID   int64  `json:"id" description:"分类ID"`
	Name string `json:"name" description:"分类名称"`
	Sort int    `json:"sort" description:"排序值"`
}
