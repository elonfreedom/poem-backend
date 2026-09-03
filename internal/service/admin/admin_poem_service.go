package admin

import (
	"context"
	"errors"
	"fmt"
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
func (s *AdminPoemService) List(ctx context.Context, page, pageSize int, categoryID *int64, status, keyword, dynasty string, authorID *int64, searchScope string) (*response.PageData[adminmodel.AdminPoemResponse], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	poems, total, err := s.poemRepo.ListAll(ctx, page, pageSize, categoryID, status, keyword, dynasty, authorID, searchScope)
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
