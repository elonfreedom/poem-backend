package admin

import (
	"context"
	"time"

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
func (s *AdminPoemService) List(ctx context.Context, page, pageSize int, categoryID *int64, status, keyword, dynasty string, authorID *int64) (*response.PageData[adminmodel.AdminPoemResponse], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	poems, total, err := s.poemRepo.ListAll(ctx, page, pageSize, categoryID, status, keyword, dynasty, authorID)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	resp := toAdminPoemResponse(*poem, nil)
	return &resp, nil
}

// Create 创建诗歌
func (s *AdminPoemService) Create(ctx context.Context, req *adminmodel.AdminPoemCreateRequest, createdBy *string) (*adminmodel.AdminPoemResponse, error) {
	now := time.Now()

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

	poem := &model.Poem{
		Title:         req.Title,
		Author:        req.Author,
		Dynasty:       req.Dynasty,
		Content:       req.Content,
		Translation:   req.Translation,
		Appreciation:  req.Appreciation,
		Source:        req.Source,
		CategoryID:    req.CategoryID,
		AuthorID:      req.AuthorID,
		Tags:          req.Tags,
		CoverURL:      req.CoverURL,
		Status:        req.Status,
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
		TitlePinyin:   titlePinyin,
		ContentPinyin: contentPinyin,
		TitleSC:       titleSC,
		AuthorSC:      authorSC,
		ContentSC:     contentSC,
	}
	if poem.Status == "" {
		poem.Status = "draft"
	}

	if err := s.poemRepo.Create(ctx, poem); err != nil {
		return nil, err
	}

	resp := toAdminPoemResponse(*poem, nil)
	return &resp, nil
}

// Update 更新诗歌
func (s *AdminPoemService) Update(ctx context.Context, id int64, req *adminmodel.AdminPoemUpdateRequest) error {
	poem, err := s.poemRepo.GetByID(ctx, id)
	if err != nil {
		return err
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

	poem.UpdatedAt = time.Now()

	return s.poemRepo.Update(ctx, poem)
}

// Delete 删除诗歌
func (s *AdminPoemService) Delete(ctx context.Context, id int64) error {
	return s.poemRepo.Delete(ctx, id)
}

// UpdateStatus 更新诗歌状态
func (s *AdminPoemService) UpdateStatus(ctx context.Context, id int64, status string) error {
	return s.poemRepo.UpdateStatus(ctx, id, status)
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
		TitlePinyin:   p.TitlePinyin,
		ContentPinyin: p.ContentPinyin,
		TitleSC:       p.TitleSC,
		AuthorSC:      p.AuthorSC,
		ContentSC:     p.ContentSC,
	}
	if categoryName != nil {
		resp.CategoryName = *categoryName
	}
	return resp
}
