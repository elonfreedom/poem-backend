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

// CategoryWithCount 分类及其诗歌数量
type CategoryWithCount struct {
	model.Category
	PoemCount int64
}

// List 获取分类列表（含诗歌数量）
func (r *CategoryRepository) List(ctx context.Context) ([]CategoryWithCount, error) {
	query := `
		SELECT c.id, c.name, c.sort, c.created_at, c.updated_at,
		       COUNT(p.id) AS poem_count
		FROM categories c
		LEFT JOIN poems p ON p.category_id = c.id
		GROUP BY c.id, c.name, c.sort, c.created_at, c.updated_at
		ORDER BY c.sort, c.id
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []CategoryWithCount
	for rows.Next() {
		var c CategoryWithCount
		if err := rows.Scan(&c.ID, &c.Name, &c.Sort, &c.CreatedAt, &c.UpdatedAt, &c.PoemCount); err != nil {
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

// GetPoemCount 获取分类下的诗歌数量
func (r *CategoryRepository) GetPoemCount(ctx context.Context, categoryID int64) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM poems WHERE category_id = $1`
	err := r.db.QueryRow(ctx, query, categoryID).Scan(&count)
	return count, err
}

// Delete 删除分类
func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM categories WHERE id = $1", id)
	return err
}
