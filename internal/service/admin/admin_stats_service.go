package admin

import (
	"context"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
)

type AdminStatsService struct {
	statsRepo *repository.StatsRepository
}

func NewAdminStatsService(statsRepo *repository.StatsRepository) *AdminStatsService {
	return &AdminStatsService{statsRepo: statsRepo}
}

// Overview 总览数据
func (s *AdminStatsService) Overview(ctx context.Context) (*adminmodel.AdminStatsOverview, error) {
	totalPoems, totalUsers, totalViews, todayActive, todayCheckin, err :=
		s.statsRepo.GetOverview(ctx)
	if err != nil {
		return nil, err
	}

	return &adminmodel.AdminStatsOverview{
		TotalUsers:   totalUsers,
		TotalPoems:   totalPoems,
		TotalViews:   totalViews,
		TodayActive:  todayActive,
		TodayCheckin: todayCheckin,
	}, nil
}

// Daily 每日统计
func (s *AdminStatsService) Daily(ctx context.Context, days int) ([]adminmodel.AdminStatsDaily, error) {
	results, err := s.statsRepo.GetDailyStats(ctx, days)
	if err != nil {
		return nil, err
	}

	items := make([]adminmodel.AdminStatsDaily, 0, len(results))
	for _, r := range results {
		items = append(items, adminmodel.AdminStatsDaily{
			Date:  r.Date,
			Views: r.Views,
			Users: r.Users,
		})
	}
	return items, nil
}

// HotPoems 热门诗歌
func (s *AdminStatsService) HotPoems(ctx context.Context, limit int) ([]adminmodel.AdminStatsHotPoem, error) {
	results, err := s.statsRepo.GetHotPoems(ctx, limit)
	if err != nil {
		return nil, err
	}

	items := make([]adminmodel.AdminStatsHotPoem, 0, len(results))
	for _, r := range results {
		items = append(items, adminmodel.AdminStatsHotPoem{
			PoemID:    r.PoemID,
			Title:     r.Title,
			Author:    r.Author,
			ViewCount: r.ViewCount,
		})
	}
	return items, nil
}

// UserGrowth 用户增长
func (s *AdminStatsService) UserGrowth(ctx context.Context, days int) ([]adminmodel.AdminStatsUserGrowth, error) {
	results, err := s.statsRepo.GetUserGrowth(ctx, days)
	if err != nil {
		return nil, err
	}

	items := make([]adminmodel.AdminStatsUserGrowth, 0, len(results))
	for _, r := range results {
		items = append(items, adminmodel.AdminStatsUserGrowth{
			Date:       r.Date,
			NewUsers:   r.NewUsers,
			TotalUsers: r.TotalUsers,
		})
	}
	return items, nil
}
