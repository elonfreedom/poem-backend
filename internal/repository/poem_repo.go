package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/model"
)

type PoemRepository struct {
	db *pgxpool.Pool
}

func NewPoemRepository(db *pgxpool.Pool) *PoemRepository {
	return &PoemRepository{db: db}
}

// Create 创建诗歌
func (r *PoemRepository) Create(ctx context.Context, poem *model.Poem) error {
	query := `
		INSERT INTO poems (title, author, dynasty, content, translation, appreciation, source, tags, cover_url, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		poem.Title, poem.Author, poem.Dynasty, poem.Content,
		poem.Translation, poem.Appreciation, poem.Source, poem.Tags, poem.CoverURL,
		poem.Status, poem.CreatedBy, poem.CreatedAt, poem.UpdatedAt,
	).Scan(&poem.ID)
}

// List 获取诗歌列表
func (r *PoemRepository) List(ctx context.Context, page, pageSize int, categoryID *int64, status string, dynasty string) ([]model.Poem, int64, error) {
	where := "WHERE status = 'published'"
	args := []interface{}{}
	argIdx := 1

	if categoryID != nil {
		where += " AND category_id = $" + string(rune('0'+argIdx))
		args = append(args, *categoryID)
		argIdx++
	}
	if status != "" {
		where += " AND status = $" + string(rune('0'+argIdx))
		args = append(args, status)
		argIdx++
	}
	if dynasty != "" {
		where += " AND dynasty = $" + string(rune('0'+argIdx))
		args = append(args, dynasty)
		argIdx++
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM poems " + where
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT id, title, author, dynasty, content, translation, appreciation, source, category_id, tags, cover_url, status, created_by, created_at, updated_at
		FROM poems ` + where + `
		ORDER BY created_at DESC
		LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var poems []model.Poem
	for rows.Next() {
		var p model.Poem
		err := rows.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
			&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.Tags, &p.CoverURL,
			&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		poems = append(poems, p)
	}
	return poems, total, rows.Err()
}

// GetByID 根据 ID 获取诗歌
func (r *PoemRepository) GetByID(ctx context.Context, id int64) (*model.Poem, error) {
	query := `
		SELECT id, title, author, dynasty, content, translation, appreciation, source, category_id, tags, cover_url, status, created_by, created_at, updated_at
		FROM poems WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	var p model.Poem
	err := row.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
		&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.Tags, &p.CoverURL,
		&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Search 搜索诗歌
func (r *PoemRepository) Search(ctx context.Context, keyword string, page, pageSize int) ([]model.Poem, int64, error) {
	where := `WHERE status = 'published' AND (title ILIKE $1 OR author ILIKE $2 OR content ILIKE $3)`
	likePattern := "%" + keyword + "%"
	args := []interface{}{likePattern, likePattern, likePattern}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM poems " + where
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT id, title, author, dynasty, content, translation, appreciation, source, category_id, tags, cover_url, status, created_by, created_at, updated_at
		FROM poems ` + where + `
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var poems []model.Poem
	for rows.Next() {
		var p model.Poem
		err := rows.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
			&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.Tags, &p.CoverURL,
			&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		poems = append(poems, p)
	}
	return poems, total, rows.Err()
}

// GetDailyRecommendation 获取每日推荐
func (r *PoemRepository) GetDailyRecommendation(ctx context.Context) (*model.Poem, error) {
	// 简单实现：随机获取一首已发布的诗歌
	query := `
		SELECT id, title, author, dynasty, content, translation, appreciation, source, category_id, tags, cover_url, status, created_by, created_at, updated_at
		FROM poems WHERE status = 'published'
		ORDER BY RANDOM() LIMIT 1
	`
	row := r.db.QueryRow(ctx, query)
	var p model.Poem
	err := row.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
		&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.Tags, &p.CoverURL,
		&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// PoemWithCategory 诗歌及其分类名称
type PoemWithCategory struct {
	model.Poem
	CategoryName *string
}

// ListAll 获取诗歌列表（admin 用，不过滤 status）
func (r *PoemRepository) ListAll(ctx context.Context, page, pageSize int, categoryID *int64, status, keyword string) ([]PoemWithCategory, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if categoryID != nil {
		where += fmt.Sprintf(" AND p.category_id = $%d", argIdx)
		args = append(args, *categoryID)
		argIdx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND p.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if keyword != "" {
		where += fmt.Sprintf(" AND (p.title ILIKE $%d OR p.author ILIKE $%d)", argIdx, argIdx+1)
		likePattern := "%" + keyword + "%"
		args = append(args, likePattern, likePattern)
		argIdx += 2
	}

	countQuery := "SELECT COUNT(*) FROM poems p " + where
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.title, p.author, p.dynasty, p.content, p.translation, p.appreciation, p.source,
		       p.category_id, c.name AS category_name, p.tags, p.cover_url, p.status,
		       p.created_by, p.created_at, p.updated_at
		FROM poems p
		LEFT JOIN categories c ON p.category_id = c.id
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var poems []PoemWithCategory
	for rows.Next() {
		var p PoemWithCategory
		err := rows.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
			&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.CategoryName, &p.Tags, &p.CoverURL,
			&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		poems = append(poems, p)
	}
	return poems, total, rows.Err()
}

// Update 更新诗歌
func (r *PoemRepository) Update(ctx context.Context, poem *model.Poem) error {
	query := `
		UPDATE poems SET title = $1, author = $2, dynasty = $3, content = $4,
			translation = $5, appreciation = $6, source = $7, category_id = $8, tags = $9,
			cover_url = $10, status = $11, updated_at = $12
		WHERE id = $13
	`
	_, err := r.db.Exec(ctx, query,
		poem.Title, poem.Author, poem.Dynasty, poem.Content,
		poem.Translation, poem.Appreciation, poem.Source, poem.CategoryID, poem.Tags,
		poem.CoverURL, poem.Status, poem.UpdatedAt, poem.ID,
	)
	return err
}

// Delete 删除诗歌
func (r *PoemRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM poems WHERE id = $1", id)
	return err
}

// UpdateStatus 更新诗歌状态
func (r *PoemRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.Exec(ctx, "UPDATE poems SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err
}

// BatchUpdateStatus 批量更新诗歌状态
func (r *PoemRepository) BatchUpdateStatus(ctx context.Context, ids []int64, status string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := r.db.Exec(ctx,
		"UPDATE poems SET status = $1, updated_at = NOW() WHERE id = ANY($2)",
		status, ids,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// RecordView 记录浏览
func (r *PoemRepository) RecordView(ctx context.Context, poemID int64, userID *string) error {
	query := `INSERT INTO poem_views (poem_id, user_id) VALUES ($1, $2)`
	_, err := r.db.Exec(ctx, query, poemID, userID)
	return err
}

// IsFavorited 检查是否已收藏
func (r *PoemRepository) IsFavorited(ctx context.Context, userID string, poemID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND poem_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, userID, poemID).Scan(&exists)
	return exists, err
}
