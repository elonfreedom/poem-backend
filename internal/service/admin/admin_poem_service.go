package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
	"poem-backend/pkg/convert"
	"poem-backend/pkg/pinyin"
	"poem-backend/pkg/response"
)

type AdminPoemService struct {
	poemRepo *repository.PoemRepository
}

func NewAdminPoemService(poemRepo *repository.PoemRepository) *AdminPoemService {
	return &AdminPoemService{poemRepo: poemRepo}
}

// ExistsByTitleAuthorFirstLine 检查标题+作者+正文首句是否已存在
func (s *AdminPoemService) ExistsByTitleAuthorFirstLine(ctx context.Context, title, author, firstLine string) (bool, error) {
	return s.poemRepo.ExistsByTitleAuthorFirstLine(ctx, title, author, firstLine)
}

// List 分页获取诗歌列表
// searchScope: "title" 只搜标题, "author" 只搜作者, 空或其他值搜全部
func (s *AdminPoemService) List(ctx context.Context, page, pageSize int, categoryID *int64, status, keyword, dynasty string, authorID *int64, searchScope string, hasTranslation, hasAppreciation string) (*response.PageData[adminmodel.AdminPoemResponse], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	poems, total, err := s.poemRepo.ListAll(ctx, page, pageSize, categoryID, status, keyword, dynasty, authorID, searchScope, hasTranslation, hasAppreciation)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询诗歌列表失败: %v", err)}
	}

	items := make([]adminmodel.AdminPoemResponse, 0, len(poems))
	for _, p := range poems {
		items = append(items, toAdminPoemResponse(p.Poem, p.CategoryName))
	}
	return &response.PageData[adminmodel.AdminPoemResponse]{Items: items, Total: total}, nil
}

// GetByID 获取诗歌详情
func (s *AdminPoemService) GetByID(ctx context.Context, id int64) (*adminmodel.AdminPoemResponse, error) {
	poem, err := s.poemRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "poem not found", Detail: fmt.Sprintf("诗歌不存在: id=%d", id)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询诗歌失败: id=%d, error=%v", id, err)}
	}
	resp := toAdminPoemResponse(*poem, nil)
	return &resp, nil
}

// Create 创建诗歌
func (s *AdminPoemService) Create(ctx context.Context, req *adminmodel.AdminPoemCreateRequest, createdBy *string) (*adminmodel.AdminPoemResponse, error) {
	now := time.Now()

	// 校验唯一性：标题+作者+正文首句
	firstLine := req.Content
	if idx := strings.IndexAny(req.Content, "\n\r"); idx >= 0 {
		firstLine = req.Content[:idx]
	}
	exists, err := s.poemRepo.ExistsByTitleAuthorFirstLine(ctx, req.Title, req.Author, firstLine)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("检查诗歌唯一性失败: %v", err)}
	}
	if exists {
		return nil, fuego.ConflictError{Title: "poem exists", Detail: "该诗歌已存在（标题+作者+正文首句重复）"}
	}

	// 生成拼音（如果未提供则自动生成）
	titlePinyin := req.TitlePinyin
	if titlePinyin == "" {
		titlePinyin = pinyin.ToPinyin(req.Title)
	}
	contentPinyin := req.ContentPinyin
	if contentPinyin == "" {
		contentPinyin = pinyin.ToPinyinLines(req.Content)
	}

	// 简繁体双向转换：填一端自动生成另一端，都填则以用户输入为准
	titleSC := req.TitleSC
	if titleSC == "" && req.Title != "" {
		titleSC = convert.MustTraditionalToSimplified(req.Title)
	}
	authorSC := req.AuthorSC
	if authorSC == "" && req.Author != "" {
		authorSC = convert.MustTraditionalToSimplified(req.Author)
	}
	contentSC := req.ContentSC
	if contentSC == "" && req.Content != "" {
		contentSC = convert.MustTraditionalToSimplified(req.Content)
	}
	translationSC := req.TranslationSC
	if translationSC == "" && req.Translation != "" {
		translationSC = convert.MustTraditionalToSimplified(req.Translation)
	}
	appreciationSC := req.AppreciationSC
	if appreciationSC == "" && req.Appreciation != "" {
		appreciationSC = convert.MustTraditionalToSimplified(req.Appreciation)
	}

	poem := &model.Poem{
		Title:          req.Title,
		Author:         req.Author,
		Dynasty:        req.Dynasty,
		Content:        req.Content,
		Translation:    req.Translation,
		Appreciation:   req.Appreciation,
		Source:         req.Source,
		CategoryID:     req.CategoryID,
		AuthorID:       req.AuthorID,
		Tags:           req.Tags,
		CoverURL:       req.CoverURL,
		Status:         req.Status,
		CreatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
		TitlePinyin:    titlePinyin,
		ContentPinyin:  contentPinyin,
		TitleSC:        titleSC,
		AuthorSC:       authorSC,
		ContentSC:      contentSC,
		TranslationSC:  translationSC,
		AppreciationSC: appreciationSC,
	}
	if poem.Status == "" {
		poem.Status = "draft"
	}

	if err := s.poemRepo.Create(ctx, poem); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("创建诗歌失败: %v", err)}
	}

	resp := toAdminPoemResponse(*poem, nil)
	return &resp, nil
}

// Update 更新诗歌
func (s *AdminPoemService) Update(ctx context.Context, id int64, req *adminmodel.AdminPoemUpdateRequest) error {
	poem, err := s.poemRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "poem not found", Detail: fmt.Sprintf("诗歌不存在: id=%d", id)}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询诗歌失败: id=%d, error=%v", id, err)}
	}

	poem.Title = req.Title
	poem.Author = req.Author
	poem.Dynasty = req.Dynasty
	poem.Content = req.Content
	poem.Translation = req.Translation
	poem.Appreciation = req.Appreciation
	poem.Source = req.Source
	poem.CategoryID = req.CategoryID
	poem.AuthorID = req.AuthorID
	poem.Tags = req.Tags
	poem.CoverURL = req.CoverURL
	if req.Status != "" {
		poem.Status = req.Status
	}

	// 更新拼音（如果未提供则自动生成）
	if req.TitlePinyin != "" {
		poem.TitlePinyin = req.TitlePinyin
	} else {
		poem.TitlePinyin = pinyin.ToPinyin(req.Title)
	}
	if req.ContentPinyin != "" {
		poem.ContentPinyin = req.ContentPinyin
	} else {
		poem.ContentPinyin = pinyin.ToPinyinLines(req.Content)
	}
	// 简繁体双向转换：填一端自动生成另一端，都填则以用户输入为准
	if req.TitleSC == "" && req.Title != "" {
		poem.TitleSC = convert.MustTraditionalToSimplified(req.Title)
	}
	if req.AuthorSC == "" && req.Author != "" {
		poem.AuthorSC = convert.MustTraditionalToSimplified(req.Author)
	}
	if req.ContentSC == "" && req.Content != "" {
		poem.ContentSC = convert.MustTraditionalToSimplified(req.Content)
	}
	if req.TranslationSC == "" && req.Translation != "" {
		poem.TranslationSC = convert.MustTraditionalToSimplified(req.Translation)
	}
	if req.AppreciationSC == "" && req.Appreciation != "" {
		poem.AppreciationSC = convert.MustTraditionalToSimplified(req.Appreciation)
	}

	poem.UpdatedAt = time.Now()

	if err := s.poemRepo.Update(ctx, poem); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新诗歌失败: id=%d, error=%v", id, err)}
	}
	return nil
}

// Delete 删除诗歌
func (s *AdminPoemService) Delete(ctx context.Context, id int64) error {
	if err := s.poemRepo.Delete(ctx, id); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("删除诗歌失败: id=%d, error=%v", id, err)}
	}
	return nil
}

// validStatusTransitions 定义允许的状态转换（单向：draft → published → archived）
var validStatusTransitions = map[string]map[string]bool{
	"draft":     {"published": true, "archived": true},
	"published": {"archived": true},
	"archived":  {}, // 终态，不可转换
}

// UpdateStatus 更新诗歌状态（带状态机校验）
func (s *AdminPoemService) UpdateStatus(ctx context.Context, id int64, newStatus string) error {
	// 获取当前诗歌状态
	poem, err := s.poemRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "poem not found", Detail: fmt.Sprintf("诗歌不存在: id=%d", id)}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询诗歌失败: %v", err)}
	}

	// 校验状态转换合法性
	if allowed, ok := validStatusTransitions[poem.Status]; !ok || !allowed[newStatus] {
		return fuego.BadRequestError{
			Title:  "invalid status transition",
			Detail: fmt.Sprintf("不允许的状态转换: %s → %s（允许: draft→published, draft→archived, published→archived）", poem.Status, newStatus),
		}
	}

	if err := s.poemRepo.UpdateStatus(ctx, id, newStatus); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新诗歌状态失败: id=%d, error=%v", id, err)}
	}
	return nil
}

// BatchUpdateStatus 批量更新诗歌状态
func (s *AdminPoemService) BatchUpdateStatus(ctx context.Context, ids []int64, status string) (int64, error) {
	return s.poemRepo.BatchUpdateStatus(ctx, ids, status)
}

// EnsureSimplifiedForAllPoems 为存量诗歌批量生成简体（繁体 → 简体）
// 返回成功处理的记录数
func (s *AdminPoemService) EnsureSimplifiedForAllPoems(ctx context.Context) (int, error) {
	return s.poemRepo.EnsureSimplifiedForAllPoems(ctx)
}

// EnsurePinyinForAllPoems 为存量诗歌批量生成拼音
// 扫描 title_pinyin 为空的记录，自动生成拼音字段
// 返回成功处理的记录数
func (s *AdminPoemService) EnsurePinyinForAllPoems(ctx context.Context) (int, error) {
	return s.poemRepo.EnsurePinyinForAllPoems(ctx)
}

// BatchConvertChars 批量转换指定诗歌的字符类型
// target: "simplified" 或 "traditional"
// 返回处理结果统计
func (s *AdminPoemService) BatchConvertChars(ctx context.Context, poetryIDs []int64, target string) (*adminmodel.AdminToolBatchConvertCharsResponse, error) {
	result, err := s.poemRepo.BatchConvertChars(ctx, poetryIDs, target)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// toAdminPoemResponse 转换 Poem 为 AdminPoemResponse
func toAdminPoemResponse(p model.Poem, categoryName *string) adminmodel.AdminPoemResponse {
	resp := adminmodel.AdminPoemResponse{
		ID:            p.ID,
		Title:         p.Title,
		Author:        p.Author,
		Dynasty:       p.Dynasty,
		Content:       p.Content,
		Translation:   p.Translation,
		Appreciation:  p.Appreciation,
		Source:        p.Source,
		CategoryID:    p.CategoryID,
		AuthorID:      p.AuthorID,
		Tags:          p.Tags,
		CoverURL:      p.CoverURL,
		Status:        p.Status,
		CreatedBy:     p.CreatedBy,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		TitlePinyin:    p.TitlePinyin,
		ContentPinyin:  p.ContentPinyin,
		TitleSC:        p.TitleSC,
		AuthorSC:       p.AuthorSC,
		ContentSC:      p.ContentSC,
		TranslationSC:  p.TranslationSC,
		AppreciationSC: p.AppreciationSC,
	}
	if categoryName != nil {
		resp.CategoryName = *categoryName
	}
	return resp
}

// ==================== 诗文去重工具 ====================

// DedupScan 扫描重复诗文组（SQL 分组 + 分页，按需加载诗文详情）
// matchFields: 匹配维度数组，所有条件都满足（AND 逻辑）才视为重复
func (s *AdminPoemService) DedupScan(ctx context.Context, matchFields []string, statusFilter, dynastyFilter string, page, pageSize int) (*adminmodel.AdminToolDedupScanResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 20
	}

	// SQL 分组 + 分页，只返回组摘要（不加载全部诗文详情）
	scanResult, err := s.poemRepo.ScanDedupGroups(ctx, matchFields, statusFilter, dynastyFilter, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 构建 matchReason
	matchReason := buildMatchReason(matchFields)

	// 对每个组，按需加载诗文详情
	groups := make([]adminmodel.AdminToolDedupGroup, 0, len(scanResult.Groups))
	for _, summary := range scanResult.Groups {
		// 按需加载当前组的诗文详情
		poems, err := s.poemRepo.FetchPoemsByIDs(ctx, summary.PoemIDs)
		if err != nil {
			return nil, err
		}

		// 排序：推荐保留的排最前
		sortDedupPoems(poems)

		dedupPoems := make([]adminmodel.AdminToolDedupPoem, 0, len(poems))
		for _, p := range poems {
			dedupPoems = append(dedupPoems, toDedupPoem(p))
		}

		groups = append(groups, adminmodel.AdminToolDedupGroup{
			GroupID:           fmt.Sprintf("%x", sha256String(summary.MatchKey))[:8],
			MatchReason:       matchReason,
			MatchKey:          summary.MatchKey,
			PoemCount:         summary.PoemCount,
			Poems:             dedupPoems,
			RecommendedKeepID: poems[0].ID, // 排序后第一个是推荐保留的
		})
	}

	return &adminmodel.AdminToolDedupScanResponse{
		TotalScanned:    int(scanResult.TotalScanned),
		TotalGroups:     int(scanResult.TotalGroups),
		TotalDuplicates: int(scanResult.TotalDuplicates),
		Page:            page,
		PageSize:        pageSize,
		Groups:          groups,
	}, nil
}

// DedupExecute 执行去重（归档 + 删除）
func (s *AdminPoemService) DedupExecute(ctx context.Context, archiveIDs, deleteIDs []int64) (*adminmodel.AdminToolDedupExecuteResponse, error) {
	result := &adminmodel.AdminToolDedupExecuteResponse{}

	// 归档
	if len(archiveIDs) > 0 {
		archived, err := s.poemRepo.ArchivePoems(ctx, archiveIDs)
		if err != nil {
			return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("归档失败: %v", err)}
		}
		result.Archived = int(archived)
	}

	// 删除
	if len(deleteIDs) > 0 {
		deleted, err := s.poemRepo.DeletePoems(ctx, deleteIDs)
		if err != nil {
			return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("删除失败: %v", err)}
		}
		result.Deleted = int(deleted)
	}

	result.Message = fmt.Sprintf("处理完成：归档 %d 首，删除 %d 首", result.Archived, result.Deleted)
	return result, nil
}

// DedupMerge 智能合并重复诗文
// 将 merge_ids 中的诗文数据合并到 keep_id 对应的诗中，仅填充保留诗的空缺字段
// 合并后将被合并的诗归档（status→archived）
func (s *AdminPoemService) DedupMerge(ctx context.Context, keepID int64, mergeIDs []int64) (*adminmodel.AdminToolDedupMergeResponse, error) {
	result := &adminmodel.AdminToolDedupMergeResponse{
		KeepID:       keepID,
		MergedFields: []string{},
	}

	// 1. 获取保留诗
	keepPoem, err := s.poemRepo.GetByID(ctx, keepID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "poem not found", Detail: fmt.Sprintf("保留的诗文不存在: id=%d", keepID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询保留诗文失败: %v", err)}
	}

	// 2. 过滤掉 keep_id（防止自合并）
	filteredIDs := make([]int64, 0, len(mergeIDs))
	for _, id := range mergeIDs {
		if id != keepID {
			filteredIDs = append(filteredIDs, id)
		}
	}

	if len(filteredIDs) == 0 {
		result.Message = "没有需要合并的诗文"
		return result, nil
	}

	// 3. 获取要合并的诗（不存在的 ID 会被自动跳过）
	mergePoems, err := s.poemRepo.FetchPoemsByIDs(ctx, filteredIDs)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询待合并诗文失败: %v", err)}
	}

	if len(mergePoems) == 0 {
		result.Message = "没有找到有效的待合并诗文"
		return result, nil
	}

	// 4. 合并字段（仅当保留诗该字段为空时才填充）
	mergedFields := []string{}

	// 文本字段合并
	textFields := []struct {
		name     string
		keepPtr  *string
		getValue func(p repository.DedupPoem) string
	}{
		{"translation", &keepPoem.Translation, func(p repository.DedupPoem) string { return p.Translation }},
		{"appreciation", &keepPoem.Appreciation, func(p repository.DedupPoem) string { return p.Appreciation }},
	}

	for _, f := range textFields {
		if *f.keepPtr == "" {
			for _, p := range mergePoems {
				val := f.getValue(p)
				if val != "" {
					*f.keepPtr = val
					mergedFields = append(mergedFields, f.name)
					break
				}
			}
		}
	}

	// 拼音字段和 cover_url 合并（DedupPoem 不含这些字段，需要从完整 Poem 获取）
	needFullPoem := keepPoem.TitlePinyin == "" || keepPoem.ContentPinyin == "" || keepPoem.CoverURL == ""
	if needFullPoem {
		for _, mp := range mergePoems {
			mergeFull, err := s.poemRepo.GetByID(ctx, mp.ID)
			if err != nil {
				continue // 跳过不存在的
			}
			if keepPoem.TitlePinyin == "" && mergeFull.TitlePinyin != "" {
				keepPoem.TitlePinyin = mergeFull.TitlePinyin
				mergedFields = append(mergedFields, "title_pinyin")
			}
			if keepPoem.ContentPinyin == "" && mergeFull.ContentPinyin != "" {
				keepPoem.ContentPinyin = mergeFull.ContentPinyin
				mergedFields = append(mergedFields, "content_pinyin")
			}
			if keepPoem.CoverURL == "" && mergeFull.CoverURL != "" {
				keepPoem.CoverURL = mergeFull.CoverURL
				mergedFields = append(mergedFields, "cover_url")
			}
			// 如果所有需要的字段都有了，提前退出
			if keepPoem.TitlePinyin != "" && keepPoem.ContentPinyin != "" && keepPoem.CoverURL != "" {
				break
			}
		}
	}

	// category_id 合并
	if keepPoem.CategoryID == nil {
		for _, p := range mergePoems {
			if p.CategoryID != nil {
				keepPoem.CategoryID = p.CategoryID
				mergedFields = append(mergedFields, "category_id")
				break
			}
		}
	}

	// tags 合并（取保留诗 + 所有合并诗的并集去重）
	if len(keepPoem.Tags) == 0 {
		// 保留诗没有标签，从合并诗中收集所有标签
		tagSet := make(map[string]bool)
		for _, p := range mergePoems {
			for _, tag := range p.Tags {
				tagSet[tag] = true
			}
		}
		if len(tagSet) > 0 {
			allTags := make([]string, 0, len(tagSet))
			for tag := range tagSet {
				allTags = append(allTags, tag)
			}
			sort.Strings(allTags)
			keepPoem.Tags = allTags
			mergedFields = append(mergedFields, "tags")
		}
	} else {
		// 保留诗已有标签，补充合并诗独有的标签
		existingTags := make(map[string]bool)
		for _, tag := range keepPoem.Tags {
			existingTags[tag] = true
		}
		newTags := []string{}
		for _, p := range mergePoems {
			for _, tag := range p.Tags {
				if !existingTags[tag] {
					existingTags[tag] = true
					newTags = append(newTags, tag)
				}
			}
		}
		if len(newTags) > 0 {
			sort.Strings(newTags)
			keepPoem.Tags = append(keepPoem.Tags, newTags...)
			mergedFields = append(mergedFields, "tags")
		}
	}

	// 5. 保存更新后的保留诗
	keepPoem.UpdatedAt = time.Now()
	if err := s.poemRepo.Update(ctx, keepPoem); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新保留诗文失败: %v", err)}
	}

	// 6. 归档被合并的诗
	archived, err := s.poemRepo.ArchivePoems(ctx, filteredIDs)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("归档重复诗文失败: %v", err)}
	}

	result.MergedFields = mergedFields
	result.Archived = int(archived)
	result.Message = fmt.Sprintf("合并完成：补充了 %s，已归档 %d 首重复诗文",
		strings.Join(mergedFields, "、"), result.Archived)

	return result, nil
}

// buildMatchReason 构建匹配原因描述
func buildMatchReason(matchFields []string) string {
	parts := make([]string, 0, len(matchFields))
	for _, f := range matchFields {
		switch f {
		case "title":
			parts = append(parts, "标题")
		case "author":
			parts = append(parts, "作者")
		case "content":
			parts = append(parts, "内容")
		}
	}
	return strings.Join(parts, "+") + "相同"
}

// sha256String 计算字符串的 SHA256 hash
func sha256String(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// sortDedupPoems 排序：推荐保留的排最前
// 1. 优先保留 published > draft > archived
// 2. 同状态下，数据更完整的优先（有 translation / appreciation / category / tags）
// 3. 仍相同，保留创建时间最早的
func sortDedupPoems(poems []repository.DedupPoem) {
	statusOrder := map[string]int{"published": 0, "draft": 1, "archived": 2}

	sort.SliceStable(poems, func(i, j int) bool {
		pi, pj := poems[i], poems[j]

		// 1. 状态优先级
		si, sj := statusOrder[pi.Status], statusOrder[pj.Status]
		if si != sj {
			return si < sj
		}

		// 2. 数据完整度（分数越高越优先）
		scoreI := completenessScore(pi)
		scoreJ := completenessScore(pj)
		if scoreI != scoreJ {
			return scoreI > scoreJ
		}

		// 3. 创建时间最早的优先
		return pi.CreatedAt.Before(pj.CreatedAt)
	})
}

// completenessScore 计算诗文数据完整度分数
func completenessScore(p repository.DedupPoem) int {
	score := 0
	if p.Translation != "" {
		score += 2
	}
	if p.Appreciation != "" {
		score += 2
	}
	if p.CategoryID != nil {
		score += 1
	}
	if len(p.Tags) > 0 {
		score += 1
	}
	return score
}

// toDedupPoem 转换 DedupPoem 为 AdminToolDedupPoem
func toDedupPoem(p repository.DedupPoem) adminmodel.AdminToolDedupPoem {
	resp := adminmodel.AdminToolDedupPoem{
		ID:           p.ID,
		Title:        p.Title,
		TitleSC:      p.TitleSC,
		Author:       p.Author,
		AuthorSC:     p.AuthorSC,
		Dynasty:      p.Dynasty,
		Content:      p.Content,
		ContentSC:    p.ContentSC,
		Translation:  p.Translation,
		Appreciation: p.Appreciation,
		CategoryID:   p.CategoryID,
		Tags:         p.Tags,
		Status:       p.Status,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	if p.CategoryName != nil {
		resp.CategoryName = *p.CategoryName
	}
	return resp
}
