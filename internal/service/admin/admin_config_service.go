package admin

import (
	"context"
	"time"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
)

type AdminConfigService struct {
	configRepo *repository.ConfigRepository
}

func NewAdminConfigService(configRepo *repository.ConfigRepository) *AdminConfigService {
	return &AdminConfigService{configRepo: configRepo}
}

// List 获取配置列表
func (s *AdminConfigService) List(ctx context.Context) ([]adminmodel.AdminConfigResponse, error) {
	configs, err := s.configRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]adminmodel.AdminConfigResponse, 0, len(configs))
	for _, c := range configs {
		items = append(items, toAdminConfigResponse(c))
	}
	return items, nil
}

// GetByKey 获取单个配置
func (s *AdminConfigService) GetByKey(ctx context.Context, key string) (*adminmodel.AdminConfigResponse, error) {
	config, err := s.configRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	resp := toAdminConfigResponse(*config)
	return &resp, nil
}

// Update 更新配置
func (s *AdminConfigService) Update(ctx context.Context, key string, req *adminmodel.AdminConfigUpdateRequest) error {
	config, err := s.configRepo.GetByKey(ctx, key)
	if err != nil {
		// 不存在则创建
		config = &model.SystemConfig{
			Key:       key,
			Value:     req.Value,
			Remark:    req.Remark,
			UpdatedAt: time.Now(),
		}
		return s.configRepo.Create(ctx, config)
	}

	config.Value = req.Value
	config.Remark = req.Remark
	config.UpdatedAt = time.Now()
	return s.configRepo.Update(ctx, config)
}

func toAdminConfigResponse(c model.SystemConfig) adminmodel.AdminConfigResponse {
	return adminmodel.AdminConfigResponse{
		Key:       c.Key,
		Value:     c.Value,
		Remark:    c.Remark,
		UpdatedAt: c.UpdatedAt,
	}
}
