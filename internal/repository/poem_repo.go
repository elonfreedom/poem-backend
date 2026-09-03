package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

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
// searchScope: "title" 只搜标题, "author" 只搜作者, "content" 只搜内容, 空/"all" 搜全部
func (r *PoemRepository) Search(ctx context.Context, keyword string, page, pageSize int, searchScope string) ([]model.Poem, int64, error) {
	where := "WHERE p.status = 'published' AND ("
	likePattern := "%" + keyword + "%"

	switch searchScope {
	case "title":
		where += "p.title ILIKE $1 OR p.title_sc ILIKE $2"
		args := []interface{}{likePattern, likePattern}
		return r.searchExec(ctx, where+")", args, page, pageSize)
	case "author":
		where += "p.author ILIKE $1 OR p.author_sc ILIKE $2"
		args := []interface{}{likePattern, likePattern}
		return r.searchExec(ctx, where+")", args, page, pageSize)
	case "content":
		where += "p.content ILIKE $1 OR p.content_sc ILIKE $2"
		args := []interface{}{likePattern, likePattern}
		return r.searchExec(ctx, where+")", args, page, pageSize)
	default: // "" 或 "all"：搜全部
		where += "p.title ILIKE $1 OR p.author ILIKE $2 OR p.content ILIKE $3 OR p.title_sc ILIKE $4 OR p.author_sc ILIKE $5 OR p.content_sc ILIKE $6)"
		args := []interface{}{likePattern, likePattern, likePattern, likePattern, likePattern, likePattern}
		return r.searchExec(ctx, where, args, page, pageSize)
	}
}

// searchExec 执行搜索查询（复用 count + list 逻辑）
func (r *PoemRepository) searchExec(ctx context.Context, where string, args []interface{}, page, pageSize int) ([]model.Poem, int64, error) {
	countQuery := "SELECT COUNT(*) FROM poems p LEFT JOIN authors a ON p.author_id = a.id " + where
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	argIdx := len(args) + 1
	query := fmt.Sprintf(`
		SELECT p.id, p.title, p.author, COALESCE(a.dynasty, p.dynasty) AS dynasty, p.content, p.translation, p.appreciation, p.source, p.category_id, p.tags, p.cover_url, p.status, p.created_by, p.created_at, p.updated_at,
		       p.title_pinyin, p.content_pinyin, p.title_sc, p.author_sc, p.content_sc, p.translation_sc, p.appreciation_sc, p.author_id
		FROM poems p
		LEFT JOIN authors a ON p.author_id = a.id %s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
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
// searchScope: "title" 只搜标题, "author" 只搜作者, 空或其他值搜全部
func (r *PoemRepository) ListAll(ctx context.Context, page, pageSize int, categoryID *int64, status, keyword, dynasty string, authorID *int64, searchScope string) ([]PoemWithCategory, int64, error) {
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
		likePattern := "%" + keyword + "%"
		if searchScope == "title" {
			where += fmt.Sprintf(" AND (p.title ILIKE $%d OR p.title_sc ILIKE $%d)", argIdx, argIdx+1)
			args = append(args, likePattern, likePattern)
			argIdx += 2
		} else if searchScope == "author" {
			where += fmt.Sprintf(" AND (p.author ILIKE $%d OR p.author_sc ILIKE $%d)", argIdx, argIdx+1)
			args = append(args, likePattern, likePattern)
			argIdx += 2
		} else {
			where += fmt.Sprintf(" AND (p.title ILIKE $%d OR p.author ILIKE $%d OR p.title_sc ILIKE $%d OR p.author_sc ILIKE $%d)", argIdx, argIdx+1, argIdx+2, argIdx+3)
			args = append(args, likePattern, likePattern, likePattern, likePattern)
			argIdx += 4
		}
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

// DedupPoem 去重扫描用的诗文信息（包含分类名称）
type DedupPoem struct {
	ID           int64
	Title        string
	TitleSC      string
	Author       string
	AuthorSC     string
	Dynasty      string
	Content      string
	ContentSC    string
	Translation  string
	Appreciation string
	CategoryID   *int64
	CategoryName *string
	Tags         []string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DedupGroupSummary 重复组摘要（SQL 分组结果）
type DedupGroupSummary struct {
	MatchKey    string // 匹配键（用于分组）
	Title       string // 组内任意一个标题（用于展示）
	Author      string // 组内任意一个作者（用于展示）
	PoemCount   int64  // 组内诗文数量
	PoemIDs     []int64 // 组内所有诗文 ID
}

// DedupScanResult 去重扫描结果
type DedupScanResult struct {
	TotalScanned   int64
	TotalGroups    int64
	TotalDuplicates int64
	Groups         []DedupGroupSummary
}

// ScanDedupGroups SQL 分组查询重复诗文（分页返回组摘要，不加载全部诗文详情）
func (r *PoemRepository) ScanDedupGroups(ctx context.Context, matchFields []string, statusFilter, dynastyFilter string, page, pageSize int) (*DedupScanResult, error) {
	// 构建 GROUP BY 字段（使用简体字段比较，忽略繁简差异）
	groupFields := make([]string, 0, len(matchFields))
	for _, f := range matchFields {
		switch f {
		case "title":
			groupFields = append(groupFields, "p.title_sc")
		case "author":
			groupFields = append(groupFields, "p.author_sc")
		case "content":
			// 内容用 MD5 hash 比较（忽略标点和空格差异在 SQL 中较复杂，这里用标准化后的 hash）
			groupFields = append(groupFields, "md5(lower(regexp_replace(p.content_sc, '[[:punct:][:space:]]', '', 'g')))")
		}
	}
	groupBy := strings.Join(groupFields, ", ")

	// 构建 WHERE 条件
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if statusFilter != "" {
		where += fmt.Sprintf(" AND p.status = $%d", argIdx)
		args = append(args, statusFilter)
		argIdx++
	}
	if dynastyFilter != "" {
		where += fmt.Sprintf(" AND COALESCE(a.dynasty, p.dynasty) = $%d", argIdx)
		args = append(args, dynastyFilter)
		argIdx++
	}

	// 子查询：按 matchFields 分组，筛选 count > 1 的组
	subQuery := fmt.Sprintf(`
		SELECT %s, COUNT(*) AS cnt, ARRAY_AGG(p.id ORDER BY p.created_at ASC) AS ids,
		       MIN(p.title) AS title, MIN(p.author) AS author
		FROM poems p
		LEFT JOIN authors a ON p.author_id = a.id
		%s
		GROUP BY %s
		HAVING COUNT(*) > 1
	`, groupBy, where, groupBy)

	// 统计总数
	countQuery := fmt.Sprintf(`
		WITH groups AS (%s)
		SELECT COUNT(*) AS total_groups, COALESCE(SUM(cnt), 0) AS total_poems, COALESCE(SUM(cnt - 1), 0) AS total_duplicates
		FROM groups
	`, subQuery)

	var totalGroups, totalPoems, totalDuplicates int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalGroups, &totalPoems, &totalDuplicates)
	if err != nil {
		return nil, fmt.Errorf("统计重复组失败: %w", err)
	}

	// 分页查询组列表
	offset := (page - 1) * pageSize
	pageQuery := fmt.Sprintf(`
		WITH groups AS (%s)
		SELECT title, author, cnt, ids
		FROM groups
		ORDER BY cnt DESC, title ASC
		LIMIT $%d OFFSET $%d
	`, subQuery, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, pageQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("查询重复组失败: %w", err)
	}
	defer rows.Close()

	groups := make([]DedupGroupSummary, 0)
	for rows.Next() {
		var g DedupGroupSummary
		err := rows.Scan(&g.Title, &g.Author, &g.PoemCount, &g.PoemIDs)
		if err != nil {
			return nil, fmt.Errorf("扫描重复组行失败: %w", err)
		}
		g.MatchKey = buildMatchKey(g.Title, g.Author)
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 计算扫描总数（所有符合筛选条件的诗文）
	totalScannedQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM poems p
		LEFT JOIN authors a ON p.author_id = a.id
		%s
	`, where)
	// 去掉分页参数重新查
	scanArgs := args[:len(args)-2]
	var totalScanned int64
	err = r.db.QueryRow(ctx, totalScannedQuery, scanArgs...).Scan(&totalScanned)
	if err != nil {
		return nil, fmt.Errorf("统计扫描总数失败: %w", err)
	}

	return &DedupScanResult{
		TotalScanned:    totalScanned,
		TotalGroups:     totalGroups,
		TotalDuplicates: totalDuplicates,
		Groups:          groups,
	}, nil
}

// FetchPoemsByIDs 根据 ID 列表批量获取诗文详情
func (r *PoemRepository) FetchPoemsByIDs(ctx context.Context, ids []int64) ([]DedupPoem, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT p.id, p.title, p.title_sc, p.author, p.author_sc,
		       COALESCE(a.dynasty, p.dynasty) AS dynasty,
		       p.content, p.content_sc, p.translation, p.appreciation,
		       p.category_id, c.name AS category_name, p.tags, p.status,
		       p.created_at, p.updated_at
		FROM poems p
		LEFT JOIN authors a ON p.author_id = a.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = ANY($1)
		ORDER BY p.created_at ASC
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	poems := make([]DedupPoem, 0, len(ids))
	for rows.Next() {
		var p DedupPoem
		err := rows.Scan(
			&p.ID, &p.Title, &p.TitleSC, &p.Author, &p.AuthorSC,
			&p.Dynasty, &p.Content, &p.ContentSC, &p.Translation, &p.Appreciation,
			&p.CategoryID, &p.CategoryName, &p.Tags, &p.Status,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		poems = append(poems, p)
	}
	return poems, rows.Err()
}

// buildMatchKey 构建匹配键（用于前端展示）
func buildMatchKey(title, author string) string {
	title = strings.TrimSpace(title)
	author = strings.TrimSpace(author)
	if title != "" && author != "" {
		return title + " - " + author
	}
	if title != "" {
		return title
	}
	return author
}

// ArchivePoems 批量归档诗文（status 改为 archived）
func (r *PoemRepository) ArchivePoems(ctx context.Context, ids []int64) (int64, error) {
	return r.BatchUpdateStatus(ctx, ids, "archived")
}

// DeletePoems 批量删除诗文
func (r *PoemRepository) DeletePoems(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := r.db.Exec(ctx, "DELETE FROM poems WHERE id = ANY($1)", ids)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
