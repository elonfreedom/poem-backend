package model

import (
	"time"
)

// Tag 标签模型
type Tag struct {
	ID        int64     `json:"id" db:"id" description:"标签ID（自增）"`
	Name      string    `json:"name" db:"name" description:"标签名称"`
	CreatedAt time.Time `json:"created_at" db:"created_at" description:"创建时间"`
}
