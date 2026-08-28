package model

import "time"

// Author 作者模型
type Author struct {
	ID              int64     `json:"id" db:"id" description:"作者ID"`
	Name            string    `json:"name" db:"name" description:"作者名（简体）"`
	NameTraditional string    `json:"name_traditional" db:"name_traditional" description:"作者名（繁体）"`
	Dynasty         string    `json:"dynasty" db:"dynasty" description:"朝代"`
	Biography       string    `json:"biography" db:"biography" description:"作者简介"`
	CreatedAt       time.Time `json:"created_at" db:"created_at" description:"创建时间"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at" description:"更新时间"`
}
