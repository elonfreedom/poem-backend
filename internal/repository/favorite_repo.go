package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	usermodel "poem-backend/internal/model/user"
)

type FavoriteRepository struct {
	db *pgxpool.Pool
}

func NewFavoriteRepository(db *pgxpool.Pool) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Create 创建收藏
func (r *FavoriteRepository) Create(ctx context.Context, userID string, poemID int64) error {
	query := `
		INSERT INTO favorites (user_id, poem_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, poem_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, userID, poemID)
	return err
}

// Delete 取消收藏
func (r *FavoriteRepository) Delete(ctx context.Context, userID string, poemID int64) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND poem_id = $2`
	_, err := r.db.Exec(ctx, query, userID, poemID)
	return err
}

// CountByUserID 获取用户收藏数量
func (r *FavoriteRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// List 获取收藏列表
func (r *FavoriteRepository) List(ctx context.Context, userID string, page, pageSize int) ([]usermodel.Favorite, int64, error) {
	// 获取总数
	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT user_id, poem_id, created_at
		FROM favorites
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var favorites []usermodel.Favorite
	for rows.Next() {
		var f usermodel.Favorite
		err := rows.Scan(&f.UserID, &f.PoemID, &f.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		favorites = append(favorites, f)
	}
	return favorites, total, rows.Err()
}
