package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5"

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
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询分类列表失败: %v", err)}
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
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("创建分类失败: %v", err)}
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
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "category not found", Detail: fmt.Sprintf("分类不存在: id=%d", id)}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询分类失败: id=%d, error=%v", id, err)}
	}

	category.Name = req.Name
	category.Sort = req.Sort
	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新分类失败: id=%d, error=%v", id, err)}
	}
	return nil
}

// GetPoemCount 获取分类下的诗歌数量
func (s *AdminCategoryService) GetPoemCount(ctx context.Context, categoryID int64) (int, error) {
	return s.categoryRepo.GetPoemCount(ctx, categoryID)
}

// Delete 删除分类
func (s *AdminCategoryService) Delete(ctx context.Context, id int64) error {
	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("删除分类失败: id=%d, error=%v", id, err)}
	}
	return nil
}
