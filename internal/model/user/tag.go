package usermodel

// TagResponse 标签响应
type TagResponse struct {
	ID   int64  `json:"id" description:"标签ID"`
	Name string `json:"name" description:"标签名称"`
}
