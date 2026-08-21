package admin

import (
	"context"
	"time"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
)

type AdminAnnouncementService struct {
	announcementRepo *repository.AnnouncementRepository
}

func NewAdminAnnouncementService(announcementRepo *repository.AnnouncementRepository) *AdminAnnouncementService {
	return &AdminAnnouncementService{announcementRepo: announcementRepo}
}

// List 获取公告列表
func (s *AdminAnnouncementService) List(ctx context.Context) ([]adminmodel.AdminAnnouncementResponse, error) {
	announcements, err := s.announcementRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]adminmodel.AdminAnnouncementResponse, 0, len(announcements))
	for _, a := range announcements {
		items = append(items, toAdminAnnouncementResponse(a))
	}
	return items, nil
}

// GetByID 获取公告详情
func (s *AdminAnnouncementService) GetByID(ctx context.Context, id int64) (*adminmodel.AdminAnnouncementResponse, error) {
	announcement, err := s.announcementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toAdminAnnouncementResponse(*announcement)
	return &resp, nil
}

// Create 创建公告
func (s *AdminAnnouncementService) Create(ctx context.Context, req *adminmodel.AdminAnnouncementCreateRequest) (*adminmodel.AdminAnnouncementResponse, error) {
	now := time.Now()
	announcement := &model.Announcement{
		Title:     req.Title,
		Content:   req.Content,
		Status:    req.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if announcement.Status == "" {
		announcement.Status = "draft"
	}

	if err := s.announcementRepo.Create(ctx, announcement); err != nil {
		return nil, err
	}

	resp := toAdminAnnouncementResponse(*announcement)
	return &resp, nil
}

// Update 更新公告
func (s *AdminAnnouncementService) Update(ctx context.Context, id int64, req *adminmodel.AdminAnnouncementUpdateRequest) error {
	announcement, err := s.announcementRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	announcement.Title = req.Title
	announcement.Content = req.Content
	announcement.Status = req.Status
	announcement.UpdatedAt = time.Now()

	return s.announcementRepo.Update(ctx, announcement)
}

// Delete 删除公告
func (s *AdminAnnouncementService) Delete(ctx context.Context, id int64) error {
	return s.announcementRepo.Delete(ctx, id)
}

func toAdminAnnouncementResponse(a model.Announcement) adminmodel.AdminAnnouncementResponse {
	return adminmodel.AdminAnnouncementResponse{
		ID:        a.ID,
		Title:     a.Title,
		Content:   a.Content,
		Status:    a.Status,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
