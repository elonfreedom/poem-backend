package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/model"
)

type TagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(db *pgxpool.Pool) *TagRepository {
	return &TagRepository{db: db}
}

// List 获取标签列表
func (r *TagRepository) List(ctx context.Context) ([]model.Tag, error) {
	query := `SELECT id, name, created_at FROM tags ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// GetByID 根据 ID 获取标签
func (r *TagRepository) GetByID(ctx context.Context, id int64) (*model.Tag, error) {
	query := `SELECT id, name, created_at FROM tags WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var t model.Tag
	err := row.Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Create 创建标签
func (r *TagRepository) Create(ctx context.Context, tag *model.Tag) error {
	query := `INSERT INTO tags (name, created_at) VALUES ($1, $2) RETURNING id`
	return r.db.QueryRow(ctx, query, tag.Name, tag.CreatedAt).Scan(&tag.ID)
}

// Delete 删除标签
func (r *TagRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM tags WHERE id = $1", id)
	return err
}
