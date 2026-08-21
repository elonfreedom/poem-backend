package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/model"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// List 获取分类列表
func (r *CategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	query := `SELECT id, name, sort, created_at, updated_at FROM categories ORDER BY sort, id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Sort, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// GetByID 根据 ID 获取分类
func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*model.Category, error) {
	query := `SELECT id, name, sort, created_at, updated_at FROM categories WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var c model.Category
	err := row.Scan(&c.ID, &c.Name, &c.Sort, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Create 创建分类
func (r *CategoryRepository) Create(ctx context.Context, category *model.Category) error {
	query := `INSERT INTO categories (name, sort, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRow(ctx, query, category.Name, category.Sort, category.CreatedAt, category.UpdatedAt).Scan(&category.ID)
}

// Update 更新分类
func (r *CategoryRepository) Update(ctx context.Context, category *model.Category) error {
	query := `UPDATE categories SET name = $1, sort = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, category.Name, category.Sort, category.UpdatedAt, category.ID)
	return err
}

// Delete 删除分类
func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM categories WHERE id = $1", id)
	return err
}
