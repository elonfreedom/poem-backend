package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/model"
)

type AnnouncementRepository struct {
	db *pgxpool.Pool
}

func NewAnnouncementRepository(db *pgxpool.Pool) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

// List 获取公告列表
func (r *AnnouncementRepository) List(ctx context.Context) ([]model.Announcement, error) {
	query := `SELECT id, title, content, status, created_at, updated_at FROM announcements ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announcements []model.Announcement
	for rows.Next() {
		var a model.Announcement
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		announcements = append(announcements, a)
	}
	return announcements, rows.Err()
}

// GetByID 根据 ID 获取公告
func (r *AnnouncementRepository) GetByID(ctx context.Context, id int64) (*model.Announcement, error) {
	query := `SELECT id, title, content, status, created_at, updated_at FROM announcements WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var a model.Announcement
	err := row.Scan(&a.ID, &a.Title, &a.Content, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Create 创建公告
func (r *AnnouncementRepository) Create(ctx context.Context, announcement *model.Announcement) error {
	query := `
		INSERT INTO announcements (title, content, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		announcement.Title, announcement.Content, announcement.Status,
		announcement.CreatedAt, announcement.UpdatedAt,
	).Scan(&announcement.ID)
}

// Update 更新公告
func (r *AnnouncementRepository) Update(ctx context.Context, announcement *model.Announcement) error {
	query := `
		UPDATE announcements SET title = $1, content = $2, status = $3, updated_at = $4 WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query,
		announcement.Title, announcement.Content, announcement.Status,
		announcement.UpdatedAt, announcement.ID,
	)
	return err
}

// Delete 删除公告
func (r *AnnouncementRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM announcements WHERE id = $1", id)
	return err
}
