package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/model"
	"poem-backend/pkg/convert"
)

type AuthorRepository struct {
	db *pgxpool.Pool
}

func NewAuthorRepository(db *pgxpool.Pool) *AuthorRepository {
	return &AuthorRepository{db: db}
}

// List 分页获取作者列表
func (r *AuthorRepository) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Author, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if keyword != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR name_traditional ILIKE $%d OR dynasty ILIKE $%d)", argIdx, argIdx+1, argIdx+2)
		likePattern := "%" + keyword + "%"
		args = append(args, likePattern, likePattern, likePattern)
		argIdx += 3
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM authors " + where
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count authors failed: %w", err)
	}

	// 获取列表
	query := fmt.Sprintf(`
		SELECT id, name, name_traditional, dynasty, biography, created_at, updated_at
		FROM authors %s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query authors failed: %w", err)
	}
	defer rows.Close()

	var authors []model.Author
	for rows.Next() {
		var a model.Author
		if err := rows.Scan(&a.ID, &a.Name, &a.NameTraditional, &a.Dynasty, &a.Biography, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan author failed: %w", err)
		}
		authors = append(authors, a)
	}
	return authors, total, rows.Err()
}

// GetByID 根据 ID 获取作者
func (r *AuthorRepository) GetByID(ctx context.Context, id int64) (*model.Author, error) {
	query := `
		SELECT id, name, name_traditional, dynasty, biography, created_at, updated_at
		FROM authors WHERE id = $1
	`
	var a model.Author
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.Name, &a.NameTraditional, &a.Dynasty, &a.Biography, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get author failed: %w", err)
	}
	return &a, nil
}

// Create 创建作者
func (r *AuthorRepository) Create(ctx context.Context, author *model.Author) error {
	query := `
		INSERT INTO authors (name, name_traditional, dynasty, biography)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		author.Name, author.NameTraditional, author.Dynasty, author.Biography,
	).Scan(&author.ID, &author.CreatedAt, &author.UpdatedAt)
}

// Update 更新作者
func (r *AuthorRepository) Update(ctx context.Context, author *model.Author) error {
	query := `
		UPDATE authors SET name = $1, name_traditional = $2, dynasty = $3, biography = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query,
		author.Name, author.NameTraditional, author.Dynasty, author.Biography, author.ID,
	)
	return err
}

// Delete 删除作者
func (r *AuthorRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM authors WHERE id = $1", id)
	return err
}

// SearchByKeyword 根据关键词搜索作者（用于下拉框）
func (r *AuthorRepository) SearchByKeyword(ctx context.Context, keyword string, limit int) ([]model.Author, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query := `
		SELECT id, name, name_traditional, dynasty, biography, created_at, updated_at
		FROM authors
		WHERE name ILIKE $1 OR name_traditional ILIKE $2 OR dynasty ILIKE $3
		ORDER BY id DESC
		LIMIT $4
	`
	likePattern := "%" + keyword + "%"
	rows, err := r.db.Query(ctx, query, likePattern, likePattern, likePattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search authors failed: %w", err)
	}
	defer rows.Close()

	var authors []model.Author
	for rows.Next() {
		var a model.Author
		if err := rows.Scan(&a.ID, &a.Name, &a.NameTraditional, &a.Dynasty, &a.Biography, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan author failed: %w", err)
		}
		authors = append(authors, a)
	}
	return authors, rows.Err()
}

// GetPoemCount 获取作者关联的诗歌数量
func (r *AuthorRepository) GetPoemCount(ctx context.Context, authorID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM poems WHERE author_id = $1", authorID).Scan(&count)
	return count, err
}

// GenerateAuthorsFromPoems 从诗歌中提取不重复的作者名，自动创建作者记录
// 返回统计信息：唯一作者数、新建数、跳过数
func (r *AuthorRepository) GenerateAuthorsFromPoems(ctx context.Context) (*adminmodel.AdminToolGenerateAuthorsResponse, error) {
	// 1. 查询所有不重复的非空 author
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT author FROM poems
		WHERE author IS NOT NULL AND author != ''
		ORDER BY author
	`)
	if err != nil {
		return nil, fmt.Errorf("query distinct authors failed: %w", err)
	}

	var authorNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan author name failed: %w", err)
		}
		authorNames = append(authorNames, name)
	}
	rows.Close()

	if len(authorNames) == 0 {
		return &adminmodel.AdminToolGenerateAuthorsResponse{}, nil
	}

	// 2. 查询已存在的作者名（简体 + 繁体）
	existingRows, err := r.db.Query(ctx, `SELECT name, name_traditional FROM authors`)
	if err != nil {
		return nil, fmt.Errorf("query existing authors failed: %w", err)
	}

	existingNames := make(map[string]bool)
	for existingRows.Next() {
		var name, nameTraditional string
		if err := existingRows.Scan(&name, &nameTraditional); err != nil {
			existingRows.Close()
			return nil, fmt.Errorf("scan existing author failed: %w", err)
		}
		existingNames[name] = true
		existingNames[nameTraditional] = true
	}
	existingRows.Close()

	// 3. 遍历，跳过已存在的，创建新的
	result := &adminmodel.AdminToolGenerateAuthorsResponse{
		TotalUnique: len(authorNames),
	}

	for _, name := range authorNames {
		if existingNames[name] {
			result.Skipped++
			continue
		}

		// 自动从简体生成繁体
		nameTraditional := convert.MustSimplifiedToTraditional(name)

		_, err := r.db.Exec(ctx, `
			INSERT INTO authors (name, name_traditional, dynasty, biography)
			VALUES ($1, $2, '未知', '')
		`, name, nameTraditional)
		if err != nil {
			return nil, fmt.Errorf("insert author %q failed: %w", name, err)
		}
		result.Created++
	}

	return result, nil
}

// BatchMatchPoems 批量匹配诗歌关联作者
// 根据诗歌的 author 文本匹配已有作者，自动设置 author_id
func (r *AuthorRepository) BatchMatchPoems(ctx context.Context, poetryIDs []int64) (matched, unmatched int64, err error) {
	if len(poetryIDs) == 0 {
		return 0, 0, nil
	}

	// 获取这些诗歌的 author 文本
	rows, err := r.db.Query(ctx, `
		SELECT id, author FROM poems WHERE id = ANY($1) AND (author_id IS NULL OR author_id = 0)
	`, poetryIDs)
	if err != nil {
		return 0, 0, fmt.Errorf("query poems failed: %w", err)
	}

	type poemAuthor struct {
		id     int64
		author string
	}
	var poems []poemAuthor
	for rows.Next() {
		var p poemAuthor
		if err := rows.Scan(&p.id, &p.author); err != nil {
			rows.Close()
			return 0, 0, fmt.Errorf("scan poem failed: %w", err)
		}
		poems = append(poems, p)
	}
	rows.Close()

	if len(poems) == 0 {
		return 0, 0, nil
	}

	// 获取所有作者用于匹配
	authorRows, err := r.db.Query(ctx, `SELECT id, name, name_traditional FROM authors`)
	if err != nil {
		return 0, 0, fmt.Errorf("query authors failed: %w", err)
	}
	defer authorRows.Close()

	type authorInfo struct {
		id              int64
		name            string
		nameTraditional string
	}
	var authors []authorInfo
	for authorRows.Next() {
		var a authorInfo
		if err := authorRows.Scan(&a.id, &a.name, &a.nameTraditional); err != nil {
			return 0, 0, fmt.Errorf("scan author failed: %w", err)
		}
		authors = append(authors, a)
	}

	// 逐首诗歌尝试匹配
	for _, p := range poems {
		var found bool
		for _, a := range authors {
			if p.author == a.name || p.author == a.nameTraditional {
				_, err := r.db.Exec(ctx, `UPDATE poems SET author_id = $1, updated_at = NOW() WHERE id = $2`, a.id, p.id)
				if err != nil {
					return matched, unmatched, fmt.Errorf("update poem %d failed: %w", p.id, err)
				}
				matched++
				found = true
				break
			}
		}
		if !found {
			unmatched++
		}
	}

	return matched, unmatched, nil
}
