package model

import "time"

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID        int64     `json:"id" db:"id" description:"配置ID（自增）"`
	Key       string    `json:"key" db:"key" description:"配置键"`
	Value     string    `json:"value" db:"value" description:"配置值"`
	Remark    string    `json:"remark" db:"remark" description:"备注"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" description:"更新时间"`
}
