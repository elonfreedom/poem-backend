package admin

import (
	"context"
	"time"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
	"poem-backend/pkg/response"
)

type AdminPoemService struct {
	poemRepo *repository.PoemRepository
}

func NewAdminPoemService(poemRepo *repository.PoemRepository) *AdminPoemService {
	return &AdminPoemService{poemRepo: poemRepo}
}

// List 分页获取诗歌列表
func (s *AdminPoemService) List(ctx context.Context, page, pageSize int, categoryID *int64, status, keyword string) (*response.PageData[adminmodel.AdminPoemResponse], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	poems, total, err := s.poemRepo.ListAll(ctx, page, pageSize, categoryID, status, keyword)
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
	poem := &model.Poem{
		Title:        req.Title,
		Author:       req.Author,
		Dynasty:      req.Dynasty,
		Content:      req.Content,
		Translation:  req.Translation,
		Appreciation: req.Appreciation,
		CategoryID:   req.CategoryID,
		Tags:         req.Tags,
		CoverURL:     req.CoverURL,
		Status:       req.Status,
		CreatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
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
	poem.CategoryID = req.CategoryID
	poem.Tags = req.Tags
	poem.CoverURL = req.CoverURL
	if req.Status != "" {
		poem.Status = req.Status
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

// toAdminPoemResponse 转换 Poem 为 AdminPoemResponse
func toAdminPoemResponse(p model.Poem, categoryName *string) adminmodel.AdminPoemResponse {
	resp := adminmodel.AdminPoemResponse{
		ID:           p.ID,
		Title:        p.Title,
		Author:       p.Author,
		Dynasty:      p.Dynasty,
		Content:      p.Content,
		Translation:  p.Translation,
		Appreciation: p.Appreciation,
		CategoryID:   p.CategoryID,
		Tags:         p.Tags,
		CoverURL:     p.CoverURL,
		Status:       p.Status,
		CreatedBy:    p.CreatedBy,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	if categoryName != nil {
		resp.CategoryName = *categoryName
	}
	return resp
}
