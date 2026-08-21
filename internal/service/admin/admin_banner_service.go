package admin

import (
	"context"
	"time"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
)

type AdminBannerService struct {
	bannerRepo *repository.BannerRepository
}

func NewAdminBannerService(bannerRepo *repository.BannerRepository) *AdminBannerService {
	return &AdminBannerService{bannerRepo: bannerRepo}
}

// List 获取 Banner 列表
func (s *AdminBannerService) List(ctx context.Context) ([]adminmodel.AdminBannerResponse, error) {
	banners, err := s.bannerRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]adminmodel.AdminBannerResponse, 0, len(banners))
	for _, b := range banners {
		items = append(items, toAdminBannerResponse(b))
	}
	return items, nil
}

// GetByID 获取 Banner 详情
func (s *AdminBannerService) GetByID(ctx context.Context, id int64) (*adminmodel.AdminBannerResponse, error) {
	banner, err := s.bannerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toAdminBannerResponse(*banner)
	return &resp, nil
}

// Create 创建 Banner
func (s *AdminBannerService) Create(ctx context.Context, req *adminmodel.AdminBannerCreateRequest) (*adminmodel.AdminBannerResponse, error) {
	now := time.Now()
	banner := &model.Banner{
		Title:     req.Title,
		ImageURL:  req.ImageURL,
		LinkType:  req.LinkType,
		LinkValue: req.LinkValue,
		Sort:      req.Sort,
		Status:    req.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if banner.Status == "" {
		banner.Status = "active"
	}

	if err := s.bannerRepo.Create(ctx, banner); err != nil {
		return nil, err
	}

	resp := toAdminBannerResponse(*banner)
	return &resp, nil
}

// Update 更新 Banner
func (s *AdminBannerService) Update(ctx context.Context, id int64, req *adminmodel.AdminBannerUpdateRequest) error {
	banner, err := s.bannerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	banner.Title = req.Title
	banner.ImageURL = req.ImageURL
	banner.LinkType = req.LinkType
	banner.LinkValue = req.LinkValue
	banner.Sort = req.Sort
	banner.Status = req.Status
	banner.UpdatedAt = time.Now()

	return s.bannerRepo.Update(ctx, banner)
}

// Delete 删除 Banner
func (s *AdminBannerService) Delete(ctx context.Context, id int64) error {
	return s.bannerRepo.Delete(ctx, id)
}

func toAdminBannerResponse(b model.Banner) adminmodel.AdminBannerResponse {
	return adminmodel.AdminBannerResponse{
		ID:        b.ID,
		Title:     b.Title,
		ImageURL:  b.ImageURL,
		LinkType:  b.LinkType,
		LinkValue: b.LinkValue,
		Sort:      b.Sort,
		Status:    b.Status,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}
