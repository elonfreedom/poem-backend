package admin

import (
	"context"
	"fmt"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
	"poem-backend/pkg/convert"
	"poem-backend/pkg/response"
)

type AdminAuthorService struct {
	authorRepo *repository.AuthorRepository
}

func NewAdminAuthorService(authorRepo *repository.AuthorRepository) *AdminAuthorService {
	return &AdminAuthorService{authorRepo: authorRepo}
}

// List 分页获取作者列表
func (s *AdminAuthorService) List(ctx context.Context, page, pageSize int, keyword string) (*response.PageData[adminmodel.AdminAuthorResponse], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	authors, total, err := s.authorRepo.List(ctx, page, pageSize, keyword)
	if err != nil {
		return nil, fmt.Errorf("failed to list authors: %w", err)
	}

	items := make([]adminmodel.AdminAuthorResponse, 0, len(authors))
	for _, a := range authors {
		poemCount, _ := s.authorRepo.GetPoemCount(ctx, a.ID)
		items = append(items, toAdminAuthorResponse(a, poemCount))
	}
	return &response.PageData[adminmodel.AdminAuthorResponse]{Items: items, Total: total}, nil
}

// GetByID 获取作者详情
func (s *AdminAuthorService) GetByID(ctx context.Context, id int64) (*adminmodel.AdminAuthorResponse, error) {
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	poemCount, _ := s.authorRepo.GetPoemCount(ctx, author.ID)
	resp := toAdminAuthorResponse(*author, poemCount)
	return &resp, nil
}

// Create 创建作者
func (s *AdminAuthorService) Create(ctx context.Context, req *adminmodel.AdminAuthorCreateRequest) (*adminmodel.AdminAuthorResponse, error) {
	// 处理繁体：如果未提供则尝试从简体转换
	nameTraditional := req.NameTraditional
	if nameTraditional == "" && req.Name != "" {
		nameTraditional = convert.MustSimplifiedToTraditional(req.Name)
	}

	author := &model.Author{
		Name:            req.Name,
		NameTraditional: nameTraditional,
		Dynasty:         req.Dynasty,
		Biography:       req.Biography,
	}
	if author.Dynasty == "" {
		author.Dynasty = "未知"
	}

	if err := s.authorRepo.Create(ctx, author); err != nil {
		return nil, fmt.Errorf("failed to create author: %w", err)
	}

	resp := toAdminAuthorResponse(*author, 0)
	return &resp, nil
}

// Update 更新作者
func (s *AdminAuthorService) Update(ctx context.Context, id int64, req *adminmodel.AdminAuthorUpdateRequest) error {
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get author: %w", err)
	}

	author.Name = req.Name
	author.NameTraditional = req.NameTraditional
	author.Dynasty = req.Dynasty
	author.Biography = req.Biography
	if author.Dynasty == "" {
		author.Dynasty = "未知"
	}

	return s.authorRepo.Update(ctx, author)
}

// Delete 删除作者
func (s *AdminAuthorService) Delete(ctx context.Context, id int64) error {
	return s.authorRepo.Delete(ctx, id)
}

// SearchOptions 搜索作者（用于下拉框）
func (s *AdminAuthorService) SearchOptions(ctx context.Context, keyword string) ([]adminmodel.AdminAuthorOptionResponse, error) {
	authors, err := s.authorRepo.SearchByKeyword(ctx, keyword, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to search authors: %w", err)
	}

	options := make([]adminmodel.AdminAuthorOptionResponse, 0, len(authors))
	for _, a := range authors {
		options = append(options, adminmodel.AdminAuthorOptionResponse{
			ID:      a.ID,
			Name:    a.Name,
			Dynasty: a.Dynasty,
		})
	}
	return options, nil
}

// BatchMatchPoems 批量匹配诗歌关联作者
func (s *AdminAuthorService) BatchMatchPoems(ctx context.Context, poetryIDs []int64) (*adminmodel.AdminAuthorBatchMatchResponse, error) {
	matched, unmatched, err := s.authorRepo.BatchMatchPoems(ctx, poetryIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch match poems: %w", err)
	}
	return &adminmodel.AdminAuthorBatchMatchResponse{
		Total:     int64(len(poetryIDs)),
		Matched:   matched,
		Unmatched: unmatched,
	}, nil
}

// GenerateAuthorsFromPoems 从诗歌中提取不重复的作者名，自动创建作者记录
// 返回统计信息：唯一作者数、新建数、跳过数
func (s *AdminAuthorService) GenerateAuthorsFromPoems(ctx context.Context) (*adminmodel.AdminToolGenerateAuthorsResponse, error) {
	result, err := s.authorRepo.GenerateAuthorsFromPoems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate authors from poems: %w", err)
	}
	return result, nil
}

// toAdminAuthorResponse 转换 Author 为 AdminAuthorResponse
func toAdminAuthorResponse(a model.Author, poemCount int64) adminmodel.AdminAuthorResponse {
	return adminmodel.AdminAuthorResponse{
		ID:              a.ID,
		Name:            a.Name,
		NameTraditional: a.NameTraditional,
		Dynasty:         a.Dynasty,
		Biography:       a.Biography,
		PoemCount:       poemCount,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}
