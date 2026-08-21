package model

import "time"

// Announcement 公告模型
type Announcement struct {
	ID        int64     `json:"id" db:"id" description:"公告ID（自增）"`
	Title     string    `json:"title" db:"title" description:"标题"`
	Content   string    `json:"content" db:"content" description:"内容"`
	Status    string    `json:"status" db:"status" description:"状态: draft, published"`
	CreatedAt time.Time `json:"created_at" db:"created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" description:"更新时间"`
}
