package model

import (
	"time"
)

// Category 分类模型
type Category struct {
	ID        int64     `json:"id" db:"id" description:"分类ID（自增）"`
	Name      string    `json:"name" db:"name" description:"分类名称"`
	Sort      int       `json:"sort" db:"sort" description:"排序值"`
	CreatedAt time.Time `json:"created_at" db:"created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" description:"更新时间"`
}
