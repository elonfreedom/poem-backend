package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/model"
)

type ConfigRepository struct {
	db *pgxpool.Pool
}

func NewConfigRepository(db *pgxpool.Pool) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// List 获取配置列表
func (r *ConfigRepository) List(ctx context.Context) ([]model.SystemConfig, error) {
	query := `SELECT id, key, value, remark, updated_at FROM system_configs ORDER BY key`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []model.SystemConfig
	for rows.Next() {
		var c model.SystemConfig
		if err := rows.Scan(&c.ID, &c.Key, &c.Value, &c.Remark, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

// GetByKey 根据 key 获取配置
func (r *ConfigRepository) GetByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	query := `SELECT id, key, value, remark, updated_at FROM system_configs WHERE key = $1`
	row := r.db.QueryRow(ctx, query, key)
	var c model.SystemConfig
	err := row.Scan(&c.ID, &c.Key, &c.Value, &c.Remark, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Update 更新配置
func (r *ConfigRepository) Update(ctx context.Context, config *model.SystemConfig) error {
	query := `UPDATE system_configs SET value = $1, remark = $2, updated_at = $3 WHERE key = $4`
	_, err := r.db.Exec(ctx, query, config.Value, config.Remark, config.UpdatedAt, config.Key)
	return err
}

// Create 创建配置
func (r *ConfigRepository) Create(ctx context.Context, config *model.SystemConfig) error {
	query := `INSERT INTO system_configs (key, value, remark, updated_at) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRow(ctx, query, config.Key, config.Value, config.Remark, config.UpdatedAt).Scan(&config.ID)
}
