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
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询配置列表失败: %v", err)}
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "config not found", Detail: fmt.Sprintf("配置不存在: key=%s", key)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询配置失败: key=%s, error=%v", key, err)}
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
		if err := s.configRepo.Create(ctx, config); err != nil {
			return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("创建配置失败: key=%s, error=%v", key, err)}
		}
		return nil
	}

	config.Value = req.Value
	config.Remark = req.Remark
	config.UpdatedAt = time.Now()
	if err := s.configRepo.Update(ctx, config); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新配置失败: key=%s, error=%v", key, err)}
	}
	return nil
}

func toAdminConfigResponse(c model.SystemConfig) adminmodel.AdminConfigResponse {
	return adminmodel.AdminConfigResponse{
		Key:       c.Key,
		Value:     c.Value,
		Remark:    c.Remark,
		UpdatedAt: c.UpdatedAt,
	}
}
