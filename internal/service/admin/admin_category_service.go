package admin

import (
	"context"
	"time"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
)

type AdminCategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewAdminCategoryService(categoryRepo *repository.CategoryRepository) *AdminCategoryService {
	return &AdminCategoryService{categoryRepo: categoryRepo}
}

// List 获取分类列表（含诗歌数量）
func (s *AdminCategoryService) List(ctx context.Context) ([]adminmodel.AdminCategoryResponse, error) {
	categories, err := s.categoryRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	items := []adminmodel.AdminCategoryResponse{}
	for _, c := range categories {
		items = append(items, adminmodel.AdminCategoryResponse{
			ID:        c.ID,
			Name:      c.Name,
			Sort:      c.Sort,
			PoemCount: c.PoemCount,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}
	return items, nil
}

// Create 创建分类
func (s *AdminCategoryService) Create(ctx context.Context, req *adminmodel.AdminCategoryCreateRequest) (*adminmodel.AdminCategoryResponse, error) {
	now := time.Now()
	category := &model.Category{
		Name:      req.Name,
		Sort:      req.Sort,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return &adminmodel.AdminCategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		Sort:      category.Sort,
		PoemCount: 0,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}, nil
}

// Update 更新分类
func (s *AdminCategoryService) Update(ctx context.Context, id int64, req *adminmodel.AdminCategoryUpdateRequest) error {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	category.Name = req.Name
	category.Sort = req.Sort
	category.UpdatedAt = time.Now()

	return s.categoryRepo.Update(ctx, category)
}

// Delete 删除分类
func (s *AdminCategoryService) Delete(ctx context.Context, id int64) error {
	return s.categoryRepo.Delete(ctx, id)
}
