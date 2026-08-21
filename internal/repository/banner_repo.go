package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/model"
)

type BannerRepository struct {
	db *pgxpool.Pool
}

func NewBannerRepository(db *pgxpool.Pool) *BannerRepository {
	return &BannerRepository{db: db}
}

// List 获取 Banner 列表
func (r *BannerRepository) List(ctx context.Context) ([]model.Banner, error) {
	query := `SELECT id, title, image_url, link_type, link_value, sort, status, created_at, updated_at FROM banners ORDER BY sort, id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []model.Banner
	for rows.Next() {
		var b model.Banner
		if err := rows.Scan(&b.ID, &b.Title, &b.ImageURL, &b.LinkType, &b.LinkValue, &b.Sort, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}
	return banners, rows.Err()
}

// GetByID 根据 ID 获取 Banner
func (r *BannerRepository) GetByID(ctx context.Context, id int64) (*model.Banner, error) {
	query := `SELECT id, title, image_url, link_type, link_value, sort, status, created_at, updated_at FROM banners WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var b model.Banner
	err := row.Scan(&b.ID, &b.Title, &b.ImageURL, &b.LinkType, &b.LinkValue, &b.Sort, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// Create 创建 Banner
func (r *BannerRepository) Create(ctx context.Context, banner *model.Banner) error {
	query := `
		INSERT INTO banners (title, image_url, link_type, link_value, sort, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		banner.Title, banner.ImageURL, banner.LinkType, banner.LinkValue,
		banner.Sort, banner.Status, banner.CreatedAt, banner.UpdatedAt,
	).Scan(&banner.ID)
}

// Update 更新 Banner
func (r *BannerRepository) Update(ctx context.Context, banner *model.Banner) error {
	query := `
		UPDATE banners SET title = $1, image_url = $2, link_type = $3, link_value = $4,
			sort = $5, status = $6, updated_at = $7 WHERE id = $8
	`
	_, err := r.db.Exec(ctx, query,
		banner.Title, banner.ImageURL, banner.LinkType, banner.LinkValue,
		banner.Sort, banner.Status, banner.UpdatedAt, banner.ID,
	)
	return err
}

// Delete 删除 Banner
func (r *BannerRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM banners WHERE id = $1", id)
	return err
}
