package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/pkg/convert"
)

type AuthorRepository struct {
	db *pgxpool.Pool
}

func NewAuthorRepository(db *pgxpool.Pool) *AuthorRepository {
	return &AuthorRepository{db: db}
}

// List 分页获取作者列表（支持排序）
func (r *AuthorRepository) List(ctx context.Context, page, pageSize int, keyword, sortField, sortOrder string) ([]model.Author, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if keyword != "" {
		where += fmt.Sprintf(" AND (a.name ILIKE $%d OR a.name_traditional ILIKE $%d OR a.dynasty ILIKE $%d)", argIdx, argIdx+1, argIdx+2)
		likePattern := "%" + keyword + "%"
		args = append(args, likePattern, likePattern, likePattern)
		argIdx += 3
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM authors a " + where
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count authors failed: %w", err)
	}

	// 构建 ORDER BY
	orderBy := buildAuthorOrderBy(sortField, sortOrder)

	// 获取列表（LEFT JOIN 统计诗歌数量）
	query := fmt.Sprintf(`
		SELECT a.id, a.name, a.name_traditional, a.dynasty, a.biography, a.created_at, a.updated_at, COUNT(p.id) AS poem_count
		FROM authors a
		LEFT JOIN poems p ON p.author_id = a.id
		%s
		GROUP BY a.id
		%s
		LIMIT $%d OFFSET $%d
	`, where, orderBy, argIdx, argIdx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query authors failed: %w", err)
	}
	defer rows.Close()

	var authors []model.Author
	for rows.Next() {
		var a model.Author
		if err := rows.Scan(&a.ID, &a.Name, &a.NameTraditional, &a.Dynasty, &a.Biography, &a.CreatedAt, &a.UpdatedAt, &a.PoemCount); err != nil {
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

// GenerateAuthorsFromPoems 从诗歌中提取不重复的作者名（含朝代），自动创建作者记录
// 返回统计信息：唯一作者数、新建数、跳过数、附带朝代数
// 朝代策略：按作者名分组，取出现最频的朝代；频率相同则取最近创建的诗歌的朝代
func (r *AuthorRepository) GenerateAuthorsFromPoems(ctx context.Context) (*adminmodel.AdminToolGenerateAuthorsResponse, error) {
	// 1. 查询所有不重复的「author + dynasty」组合及最近创建时间
	rows, err := r.db.Query(ctx, `
		SELECT author, dynasty, COUNT(*) AS cnt, MAX(created_at) AS latest
		FROM poems
		WHERE author IS NOT NULL AND author != ''
		GROUP BY author, dynasty
		ORDER BY author, cnt DESC, latest DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query author-dynasty groups failed: %w", err)
	}

	// 按作者名聚合，取出现最频的朝代（频率相同取最近）
	type authorDynasty struct {
		dynasty string
		count   int
		latest  time.Time
	}
	authorDynasties := make(map[string]authorDynasty)

	for rows.Next() {
		var name, dynasty string
		var count int
		var latest time.Time
		if err := rows.Scan(&name, &dynasty, &count, &latest); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan author-dynasty failed: %w", err)
		}
		// 由于 ORDER BY cnt DESC, latest DESC，每组第一行即是最优解
		if _, exists := authorDynasties[name]; !exists {
			authorDynasties[name] = authorDynasty{dynasty: dynasty, count: count, latest: latest}
		}
	}
	rows.Close()

	if len(authorDynasties) == 0 {
		return &adminmodel.AdminToolGenerateAuthorsResponse{}, nil
	}

	// 2. 查询已存在的作者（简体 + 繁体 + 朝代）
	existingRows, err := r.db.Query(ctx, `SELECT name, name_traditional, dynasty FROM authors`)
	if err != nil {
		return nil, fmt.Errorf("query existing authors failed: %w", err)
	}

	existingNames := make(map[string]bool)
	backfillNames := make(map[string]bool) // 需要回填朝代的作者名
	for existingRows.Next() {
		var name, nameTraditional, dynasty string
		if err := existingRows.Scan(&name, &nameTraditional, &dynasty); err != nil {
			existingRows.Close()
			return nil, fmt.Errorf("scan existing author failed: %w", err)
		}
		existingNames[name] = true
		existingNames[nameTraditional] = true
		if dynasty == "" || dynasty == "未知" {
			backfillNames[name] = true
		}
	}
	existingRows.Close()

	// 3. 遍历，跳过已存在的，创建新的
	result := &adminmodel.AdminToolGenerateAuthorsResponse{
		TotalUnique: len(authorDynasties),
	}

	for name, ad := range authorDynasties {
		if existingNames[name] {
			result.Skipped++
			continue
		}

		// 自动从简体生成繁体
		nameTraditional := convert.MustSimplifiedToTraditional(name)

		dynasty := ad.dynasty
		if dynasty == "" {
			dynasty = "未知"
		}
		if ad.dynasty != "" {
			result.WithDynasty++
		}

		_, err := r.db.Exec(ctx, `
			INSERT INTO authors (name, name_traditional, dynasty, biography)
			VALUES ($1, $2, $3, '')
		`, name, nameTraditional, dynasty)
		if err != nil {
			return nil, fmt.Errorf("insert author %q failed: %w", name, err)
		}
		result.Created++
	}

	// 4. 回填：为 dynasty 为空/未知的已有作者从 poems 表补充朝代
	for name := range backfillNames {
		// 只回填 poems 中有朝代信息的作者
		ad, hasDynasty := authorDynasties[name]
		if !hasDynasty || ad.dynasty == "" {
			continue
		}

		_, err := r.db.Exec(ctx, `
			UPDATE authors SET dynasty = $1
			WHERE name = $2 AND (dynasty = '' OR dynasty = '未知')
		`, ad.dynasty, name)
		if err != nil {
			return nil, fmt.Errorf("backfill dynasty for %q failed: %w", name, err)
		}
		result.Backfilled++
	}

	return result, nil
}

// BatchMatchPoems 批量匹配诗歌关联作者
// 根据诗歌的 author 文本匹配已有作者，自动设置 author_id
// poetryIDs 为空时处理全部诗歌，非空时只处理指定 ID
func (r *AuthorRepository) BatchMatchPoems(ctx context.Context, poetryIDs []int64) (matched, unmatched int64, err error) {
	// 获取诗歌的 author 文本（空数组=全部诗歌，非空=指定ID）
	var rows pgx.Rows
	if len(poetryIDs) == 0 {
		rows, err = r.db.Query(ctx, `
			SELECT id, author FROM poems WHERE author_id IS NULL OR author_id = 0
		`)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, author FROM poems WHERE id = ANY($1) AND (author_id IS NULL OR author_id = 0)
		`, poetryIDs)
	}
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

// AuthorDedupScan 扫描重复作者组
// matchBy: "name" 仅按姓名分组，"name_dynasty" 按姓名+朝代分组
func (r *AuthorRepository) AuthorDedupScan(ctx context.Context, matchBy string) (*adminmodel.AdminToolAuthorDedupScanResponse, error) {
	if matchBy != "name_dynasty" {
		matchBy = "name"
	}

	// 查询重复组
	var groupBy string
	var matchReason string
	if matchBy == "name_dynasty" {
		groupBy = "name, dynasty"
		matchReason = "姓名+朝代相同"
	} else {
		groupBy = "name"
		matchReason = "姓名相同"
	}

	// 先获取总数
	var totalScanned int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM authors").Scan(&totalScanned); err != nil {
		return nil, fmt.Errorf("count authors failed: %w", err)
	}

	// 查询重复的组键
	groupRows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT %s, COUNT(*) AS cnt
		FROM authors
		GROUP BY %s
		HAVING COUNT(*) > 1
		ORDER BY cnt DESC
	`, groupBy, groupBy))
	if err != nil {
		return nil, fmt.Errorf("query duplicate groups failed: %w", err)
	}

	type groupKey struct {
		name    string
		dynasty string
	}
	var groups []groupKey
	for groupRows.Next() {
		var gk groupKey
		var cnt int
		if matchBy == "name_dynasty" {
			if err := groupRows.Scan(&gk.name, &gk.dynasty, &cnt); err != nil {
				groupRows.Close()
				return nil, fmt.Errorf("scan group key failed: %w", err)
			}
		} else {
			if err := groupRows.Scan(&gk.name, &cnt); err != nil {
				groupRows.Close()
				return nil, fmt.Errorf("scan group key failed: %w", err)
			}
		}
		groups = append(groups, gk)
	}
	groupRows.Close()

	result := &adminmodel.AdminToolAuthorDedupScanResponse{
		TotalScanned: totalScanned,
		TotalGroups:  len(groups),
		Groups:       make([]adminmodel.AdminToolAuthorDedupGroup, 0, len(groups)),
	}

	// 对每个重复组，获取作者详情
	for _, gk := range groups {
		var whereClause string
		var args []interface{}
		if matchBy == "name_dynasty" {
			whereClause = "WHERE name = $1 AND dynasty = $2"
			args = []interface{}{gk.name, gk.dynasty}
		} else {
			whereClause = "WHERE name = $1"
			args = []interface{}{gk.name}
		}

		authorRows, err := r.db.Query(ctx, fmt.Sprintf(`
			SELECT a.id, a.name, a.dynasty, a.biography, COUNT(p.id) AS poem_count
			FROM authors a
			LEFT JOIN poems p ON p.author_id = a.id
			%s
			GROUP BY a.id
			ORDER BY poem_count DESC, a.id
		`, whereClause), args...)
		if err != nil {
			return nil, fmt.Errorf("query group authors failed: %w", err)
		}

		var authors []adminmodel.AdminToolAuthorDedupItem
		for authorRows.Next() {
			var a adminmodel.AdminToolAuthorDedupItem
			if err := authorRows.Scan(&a.ID, &a.Name, &a.Dynasty, &a.Biography, &a.PoemCount); err != nil {
				authorRows.Close()
				return nil, fmt.Errorf("scan author failed: %w", err)
			}
			authors = append(authors, a)
		}
		authorRows.Close()

		groupKey := gk.name
		if matchBy == "name_dynasty" {
			groupKey = gk.name + "·" + gk.dynasty
		}

		result.Groups = append(result.Groups, adminmodel.AdminToolAuthorDedupGroup{
			GroupKey:    groupKey,
			MatchReason: matchReason,
			AuthorCount: len(authors),
			Authors:     authors,
		})
	}

	return result, nil
}

// AuthorDedupMerge 合并重复作者
// 将 merge_ids 的所有作者的 poems 关联到 keep_id，合并 biography/dynasty，然后删除 merge_ids
func (r *AuthorRepository) AuthorDedupMerge(ctx context.Context, keepID int64, mergeIDs []int64) (*adminmodel.AdminToolAuthorDedupMergeResponse, error) {
	// 过滤掉 keep_id 本身
	filtered := make([]int64, 0, len(mergeIDs))
	for _, id := range mergeIDs {
		if id != keepID {
			filtered = append(filtered, id)
		}
	}
	mergeIDs = filtered

	if len(mergeIDs) == 0 {
		return &adminmodel.AdminToolAuthorDedupMergeResponse{
			KeepID:  keepID,
			Merged:  0,
			Message: "没有需要合并的作者",
		}, nil
	}

	// 检查 keep_id 存在
	var keepExists bool
	if err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM authors WHERE id = $1)", keepID).Scan(&keepExists); err != nil {
		return nil, fmt.Errorf("check keep_id failed: %w", err)
	}
	if !keepExists {
		return nil, fmt.Errorf("keep_id %d 不存在", keepID)
	}

	// 1. 重新关联诗歌
	var reassignedPoems int64
	if len(mergeIDs) > 0 {
		tag, err := r.db.Exec(ctx, `
			UPDATE poems SET author_id = $1, updated_at = NOW()
			WHERE author_id = ANY($2)
		`, keepID, mergeIDs)
		if err != nil {
			return nil, fmt.Errorf("reassign poems failed: %w", err)
		}
		reassignedPoems = tag.RowsAffected()
	}

	// 2. 合并 biography 和 dynasty（如果 keep 为空）
	if err := r.db.QueryRow(ctx, `
		WITH merged AS (
			SELECT
				COALESCE(MAX(NULLIF(biography, '')), '') AS bio,
				COALESCE(MAX(NULLIF(dynasty, '')), '') AS dyn
			FROM authors
			WHERE id = ANY($1)
		)
		UPDATE authors SET
			biography = CASE WHEN authors.biography = '' OR authors.biography IS NULL THEN merged.bio ELSE authors.biography END,
			dynasty = CASE WHEN authors.dynasty = '' OR authors.dynasty IS NULL OR authors.dynasty = '未知' THEN merged.dyn ELSE authors.dynasty END,
			updated_at = NOW()
		FROM merged
		WHERE authors.id = $2
	`, mergeIDs, keepID).Scan(); err != nil && err.Error() != "no rows in result set" {
		// CTE UPDATE 不需要 scan，忽略
	}

	// 3. 删除被合并的作者
	if len(mergeIDs) > 0 {
		if _, err := r.db.Exec(ctx, "DELETE FROM authors WHERE id = ANY($1)", mergeIDs); err != nil {
			return nil, fmt.Errorf("delete merged authors failed: %w", err)
		}
	}

	return &adminmodel.AdminToolAuthorDedupMergeResponse{
		KeepID:          keepID,
		Merged:          len(mergeIDs),
		ReassignedPoems: reassignedPoems,
		Message:         fmt.Sprintf("合并完成：%d 个作者已合并，%d 首诗歌已重新关联", len(mergeIDs), reassignedPoems),
	}, nil
}

// CleanupAuthorNames 清理 name = name_traditional 的记录（清空冗余的 name_traditional）
func (r *AuthorRepository) CleanupAuthorNames(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE authors SET name_traditional = '', updated_at = NOW()
		WHERE name = name_traditional AND name_traditional != ''
	`)
	if err != nil {
		return 0, fmt.Errorf("cleanup author names failed: %w", err)
	}
	return tag.RowsAffected(), nil
}

// buildAuthorOrderBy 构建作者列表排序子句
func buildAuthorOrderBy(sortField, sortOrder string) string {
	// 白名单允许的排序字段
	var column string
	switch sortField {
	case "name":
		column = "a.name"
	case "poem_count":
		column = "poem_count"
	case "created_at":
		column = "a.created_at"
	case "id":
		column = "a.id"
	default:
		column = "a.id" // 默认按 id
	}

	// 排序方向
	direction := "DESC"
	if sortOrder == "asc" {
		direction = "ASC"
	}

	return fmt.Sprintf("ORDER BY %s %s", column, direction)
}

// ConvertAuthorNamesToTraditional 将作者姓名从简体转为繁体，写入 name_traditional 字段
func (r *AuthorRepository) ConvertAuthorNamesToTraditional(ctx context.Context) (int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name FROM authors
		WHERE name != ''
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("query authors failed: %w", err)
	}
	defer rows.Close()

	var processed int64
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return processed, fmt.Errorf("scan author failed: %w", err)
		}

		traditional := convert.MustSimplifiedToTraditional(name)
		if traditional == name {
			continue // 转换后无变化，跳过
		}

		if _, err := r.db.Exec(ctx, `
			UPDATE authors SET name_traditional = $1, updated_at = NOW()
			WHERE id = $2
		`, traditional, id); err != nil {
			return processed, fmt.Errorf("update author %d failed: %w", id, err)
		}
		processed++
	}
	return processed, rows.Err()
}

// EnsureAuthorNamesSimplified 确保 authors 表的 name 字段为简体字
// 扫描 name 包含繁体字的记录，转为简体（原值保留到 name_traditional）
func (r *AuthorRepository) EnsureAuthorNamesSimplified(ctx context.Context) (int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, name_traditional FROM authors
		WHERE name != ''
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("query authors failed: %w", err)
	}
	defer rows.Close()

	var processed int64
	for rows.Next() {
		var id int64
		var name, nameTraditional string
		if err := rows.Scan(&id, &name, &nameTraditional); err != nil {
			return processed, fmt.Errorf("scan author failed: %w", err)
		}

		charType := convert.DetectCharsType(name)
		// 只有包含繁体差异字（纯繁体或混合）才需要转换
		if charType != convert.CharsTypeTraditional && charType != convert.CharsTypeMixed {
			continue
		}

		simplified := convert.MustTraditionalToSimplified(name)
		if simplified == name {
			continue // 转换后无变化，跳过
		}

		// 如果 name_traditional 为空，保留原值
		newTraditional := nameTraditional
		if newTraditional == "" {
			newTraditional = name
		}

		if _, err := r.db.Exec(ctx, `
			UPDATE authors SET name = $1, name_traditional = $2, updated_at = NOW()
			WHERE id = $3
		`, simplified, newTraditional, id); err != nil {
			return processed, fmt.Errorf("update author %d failed: %w", id, err)
		}
		processed++
	}
	return processed, nil
}
