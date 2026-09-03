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
		INSERT INTO poems (title, author, dynasty, content, translation, appreciation, source, tags, cover_url, status, created_by, created_at, updated_at,
		                   title_pinyin, content_pinyin, title_sc, author_sc, content_sc, translation_sc, appreciation_sc, author_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		poem.Title, poem.Author, poem.Dynasty, poem.Content,
		poem.Translation, poem.Appreciation, poem.Source, poem.Tags, poem.CoverURL,
		poem.Status, poem.CreatedBy, poem.CreatedAt, poem.UpdatedAt,
		poem.TitlePinyin, poem.ContentPinyin, poem.TitleSC, poem.AuthorSC, poem.ContentSC,
		poem.TranslationSC, poem.AppreciationSC,
		poem.AuthorID,
	).Scan(&poem.ID)
}

// ExistsByTitleAuthorFirstLine 检查标题+作者+正文首句是否已存在
func (r *PoemRepository) ExistsByTitleAuthorFirstLine(ctx context.Context, title, author, firstLine string) (bool, error) {
	var exists bool
	// 使用 SPLIT_PART 提取数据库中 content 的首句（兼容 \n 和 \r\n）
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM poems
			WHERE title = $1 AND author = $2
			  AND SPLIT_PART(REPLACE(content, E'\r\n', E'\n'), E'\n', 1) = $3
		)
	`, title, author, firstLine).Scan(&exists)
	return exists, err
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
		where += " AND p.status = $" + string(rune('0'+argIdx))
		args = append(args, status)
		argIdx++
	}
	if dynasty != "" {
		where += " AND COALESCE(a.dynasty, p.dynasty) = $" + string(rune('0'+argIdx))
		args = append(args, dynasty)
		argIdx++
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM poems p LEFT JOIN authors a ON p.author_id = a.id " + where
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表（JOIN authors 表，以作者朝代为准）
	query := `
		SELECT p.id, p.title, p.author, COALESCE(a.dynasty, p.dynasty) AS dynasty, p.content, p.translation, p.appreciation, p.source, p.category_id, p.tags, p.cover_url, p.status, p.created_by, p.created_at, p.updated_at,
		       p.title_pinyin, p.content_pinyin, p.title_sc, p.author_sc, p.content_sc, p.translation_sc, p.appreciation_sc, p.author_id
		FROM poems p
		LEFT JOIN authors a ON p.author_id = a.id ` + where + `
		ORDER BY p.created_at DESC
		LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	poems := make([]model.Poem, 0)
	for rows.Next() {
		var p model.Poem
		err := rows.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
			&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.Tags, &p.CoverURL,
			&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
			&p.TitlePinyin, &p.ContentPinyin, &p.TitleSC, &p.AuthorSC, &p.ContentSC,
			&p.TranslationSC, &p.AppreciationSC, &p.AuthorID)
		if err != nil {
			return nil, 0, err
		}
		poems = append(poems, p)
	}
	return poems, total, rows.Err()
}

// GetByID 根据 ID 获取诗歌（JOIN authors 表，以作者朝代为准）
func (r *PoemRepository) GetByID(ctx context.Context, id int64) (*model.Poem, error) {
	query := `
		SELECT p.id, p.title, p.author, COALESCE(a.dynasty, p.dynasty) AS dynasty, p.content, p.translation, p.appreciation, p.source, p.category_id, p.tags, p.cover_url, p.status, p.created_by, p.created_at, p.updated_at,
		       p.title_pinyin, p.content_pinyin, p.title_sc, p.author_sc, p.content_sc, p.translation_sc, p.appreciation_sc, p.author_id
		FROM poems p
		LEFT JOIN authors a ON p.author_id = a.id
		WHERE p.id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	var p model.Poem
	err := row.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
		&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.Tags, &p.CoverURL,
		&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		&p.TitlePinyin, &p.ContentPinyin, &p.TitleSC, &p.AuthorSC, &p.ContentSC,
		&p.TranslationSC, &p.AppreciationSC, &p.AuthorID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Search 搜索诗歌（JOIN authors 表，以作者朝代为准）
func (r *PoemRepository) Search(ctx context.Context, keyword string, page, pageSize int) ([]model.Poem, int64, error) {
	where := `WHERE p.status = 'published' AND (p.title ILIKE $1 OR p.author ILIKE $2 OR p.content ILIKE $3 OR p.title_sc ILIKE $4 OR p.author_sc ILIKE $5 OR p.content_sc ILIKE $6)`
	likePattern := "%" + keyword + "%"
	args := []interface{}{likePattern, likePattern, likePattern, likePattern, likePattern, likePattern}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM poems p LEFT JOIN authors a ON p.author_id = a.id " + where
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT p.id, p.title, p.author, COALESCE(a.dynasty, p.dynasty) AS dynasty, p.content, p.translation, p.appreciation, p.source, p.category_id, p.tags, p.cover_url, p.status, p.created_by, p.created_at, p.updated_at,
		       p.title_pinyin, p.content_pinyin, p.title_sc, p.author_sc, p.content_sc, p.translation_sc, p.appreciation_sc, p.author_id
		FROM poems p
		LEFT JOIN authors a ON p.author_id = a.id ` + where + `
		ORDER BY p.created_at DESC
		LIMIT $7 OFFSET $8
	`
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	poems := make([]model.Poem, 0)
	for rows.Next() {
		var p model.Poem
		err := rows.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
			&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.Tags, &p.CoverURL,
			&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
			&p.TitlePinyin, &p.ContentPinyin, &p.TitleSC, &p.AuthorSC, &p.ContentSC,
			&p.TranslationSC, &p.AppreciationSC, &p.AuthorID)
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
		SELECT p.id, p.title, p.author, COALESCE(a.dynasty, p.dynasty) AS dynasty, p.content, p.translation, p.appreciation, p.source, p.category_id, p.tags, p.cover_url, p.status, p.created_by, p.created_at, p.updated_at,
		       p.title_pinyin, p.content_pinyin, p.title_sc, p.author_sc, p.content_sc, p.translation_sc, p.appreciation_sc, p.author_id
		FROM poems p
		LEFT JOIN authors a ON p.author_id = a.id
		WHERE p.status = 'published'
		ORDER BY RANDOM() LIMIT 1
	`
	row := r.db.QueryRow(ctx, query)
	var p model.Poem
	err := row.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
		&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.Tags, &p.CoverURL,
		&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		&p.TitlePinyin, &p.ContentPinyin, &p.TitleSC, &p.AuthorSC, &p.ContentSC,
		&p.TranslationSC, &p.AppreciationSC, &p.AuthorID)
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
func (r *PoemRepository) ListAll(ctx context.Context, page, pageSize int, categoryID *int64, status, keyword, dynasty string, authorID *int64) ([]PoemWithCategory, int64, error) {
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
	if dynasty != "" {
		where += fmt.Sprintf(" AND COALESCE(a.dynasty, p.dynasty) = $%d", argIdx)
		args = append(args, dynasty)
		argIdx++
	}
	if authorID != nil {
		where += fmt.Sprintf(" AND p.author_id = $%d", argIdx)
		args = append(args, *authorID)
		argIdx++
	}
	if keyword != "" {
		where += fmt.Sprintf(" AND (p.title ILIKE $%d OR p.author ILIKE $%d OR p.title_sc ILIKE $%d OR p.author_sc ILIKE $%d)", argIdx, argIdx+1, argIdx+2, argIdx+3)
		likePattern := "%" + keyword + "%"
		args = append(args, likePattern, likePattern, likePattern, likePattern)
		argIdx += 4
	}

	countQuery := "SELECT COUNT(*) FROM poems p LEFT JOIN authors a ON p.author_id = a.id " + where
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.title, p.author, COALESCE(a.dynasty, p.dynasty) AS dynasty, p.content, p.translation, p.appreciation, p.source,
		       p.category_id, c.name AS category_name, p.tags, p.cover_url, p.status,
		       p.created_by, p.created_at, p.updated_at,
		       p.title_pinyin, p.content_pinyin, p.title_sc, p.author_sc, p.content_sc, p.translation_sc, p.appreciation_sc
		FROM poems p
		LEFT JOIN authors a ON p.author_id = a.id
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

	poems := make([]PoemWithCategory, 0)
	for rows.Next() {
		var p PoemWithCategory
		err := rows.Scan(&p.ID, &p.Title, &p.Author, &p.Dynasty, &p.Content,
			&p.Translation, &p.Appreciation, &p.Source, &p.CategoryID, &p.CategoryName, &p.Tags, &p.CoverURL,
			&p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
			&p.TitlePinyin, &p.ContentPinyin, &p.TitleSC, &p.AuthorSC, &p.ContentSC, &p.TranslationSC, &p.AppreciationSC)
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
			cover_url = $10, status = $11, updated_at = $12,
			title_pinyin = $13, content_pinyin = $14, title_sc = $15, author_sc = $16, content_sc = $17,
			translation_sc = $18, appreciation_sc = $19, author_id = $20
		WHERE id = $21
	`
	_, err := r.db.Exec(ctx, query,
		poem.Title, poem.Author, poem.Dynasty, poem.Content,
		poem.Translation, poem.Appreciation, poem.Source, poem.CategoryID, poem.Tags,
		poem.CoverURL, poem.Status, poem.UpdatedAt,
		poem.TitlePinyin, poem.ContentPinyin, poem.TitleSC, poem.AuthorSC, poem.ContentSC,
		poem.TranslationSC, poem.AppreciationSC, poem.AuthorID, poem.ID,
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
