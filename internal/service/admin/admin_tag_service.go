package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
)

type AdminTagService struct {
	tagRepo *repository.TagRepository
}

func NewAdminTagService(tagRepo *repository.TagRepository) *AdminTagService {
	return &AdminTagService{tagRepo: tagRepo}
}

// List 获取标签列表
func (s *AdminTagService) List(ctx context.Context) ([]adminmodel.AdminTagResponse, error) {
	tags, err := s.tagRepo.List(ctx)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询标签列表失败: %v", err)}
	}

	items := make([]adminmodel.AdminTagResponse, 0, len(tags))
	for _, t := range tags {
		items = append(items, adminmodel.AdminTagResponse{
			ID:        t.ID,
			Name:      t.Name,
			CreatedAt: t.CreatedAt,
		})
	}
	return items, nil
}

// Create 创建标签
func (s *AdminTagService) Create(ctx context.Context, req *adminmodel.AdminTagCreateRequest) (*adminmodel.AdminTagResponse, error) {
	tag := &model.Tag{
		Name:      req.Name,
		CreatedAt: time.Now(),
	}
	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("创建标签失败: %v", err)}
	}

	return &adminmodel.AdminTagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt,
	}, nil
}

// Delete 删除标签
func (s *AdminTagService) Delete(ctx context.Context, id int64) error {
	if err := s.tagRepo.Delete(ctx, id); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("删除标签失败: id=%d, error=%v", id, err)}
	}
	return nil
}
