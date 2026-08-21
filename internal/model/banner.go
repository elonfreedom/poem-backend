package model

import "time"

// Banner Banner 轮播图模型
type Banner struct {
	ID        int64     `json:"id" db:"id" description:"Banner ID（自增）"`
	Title     string    `json:"title" db:"title" description:"标题"`
	ImageURL  string    `json:"image_url" db:"image_url" description:"图片URL"`
	LinkType  string    `json:"link_type" db:"link_type" description:"链接类型: poem, url"`
	LinkValue string    `json:"link_value" db:"link_value" description:"链接值"`
	Sort      int       `json:"sort" db:"sort" description:"排序值"`
	Status    string    `json:"status" db:"status" description:"状态: active, inactive"`
	CreatedAt time.Time `json:"created_at" db:"created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" description:"更新时间"`
}
