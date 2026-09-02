package admin

import (
	"context"
	"time"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
)

type ImportRecordService struct {
	importRecordRepo *repository.ImportRecordRepository
}

func NewImportRecordService(importRecordRepo *repository.ImportRecordRepository) *ImportRecordService {
	return &ImportRecordService{importRecordRepo: importRecordRepo}
}

// Create 创建导入记录
func (s *ImportRecordService) Create(ctx context.Context, fileName, source string, total, success, failed int, errors []adminmodel.ImportError, createdBy *string) (int64, error) {
	// 判断状态
	status := computeStatus(success, failed)

	record := &adminmodel.ImportRecord{
		FileName:  fileName,
		Source:    source,
		Total:     total,
		Success:   success,
		Failed:    failed,
		Status:    status,
		Errors:    errors,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	return s.importRecordRepo.Create(ctx, record)
}

// UpdateStatus 更新导入记录状态
func (s *ImportRecordService) UpdateStatus(ctx context.Context, id int64, success, failed int, status string, errors []adminmodel.ImportError) error {
	return s.importRecordRepo.UpdateStatus(ctx, id, success, failed, status, errors)
}

// List 分页列表
func (s *ImportRecordService) List(ctx context.Context, page, pageSize int, status, startDate, endDate string) ([]adminmodel.ImportRecord, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.importRecordRepo.List(ctx, page, pageSize, status, startDate, endDate)
}

// GetByID 单条详情
func (s *ImportRecordService) GetByID(ctx context.Context, id int64) (*adminmodel.ImportRecord, error) {
	return s.importRecordRepo.GetByID(ctx, id)
}

// GetStats 统计
func (s *ImportRecordService) GetStats(ctx context.Context, status, startDate, endDate string) (*adminmodel.ImportRecordStatsResponse, error) {
	return s.importRecordRepo.GetStats(ctx, status, startDate, endDate)
}

// computeStatus 根据成功/失败数计算导入状态
func computeStatus(success, failed int) string {
	switch {
	case failed == 0:
		return "success"
	case success == 0:
		return "failed"
	default:
		return "partial"
	}
}
