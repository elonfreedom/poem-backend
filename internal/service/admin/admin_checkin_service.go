package admin

import (
	"context"
	"fmt"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
)

type AdminCheckinService struct {
	checkinRepo *repository.CheckinRepository
}

func NewAdminCheckinService(checkinRepo *repository.CheckinRepository) *AdminCheckinService {
	return &AdminCheckinService{checkinRepo: checkinRepo}
}

// ListCheckins 获取打卡记录列表（管理后台）
func (s *AdminCheckinService) ListCheckins(ctx context.Context, page, pageSize int, keyword, startDate, endDate string) (*adminmodel.AdminCheckinListResponse, error) {
	checkins, total, err := s.checkinRepo.ListAll(ctx, page, pageSize, keyword, startDate, endDate)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询打卡记录列表失败: %v", err)}
	}

	items := make([]adminmodel.AdminCheckinListItem, 0, len(checkins))
	for i, c := range checkins {
		item := adminmodel.AdminCheckinListItem{
			ID:              int64(i + 1) + int64((page-1)*pageSize),
			UserID:          c.UserID,
			Nickname:        c.Nickname,
			CheckinDate:     c.Date.Format("2006-01-02"),
			PoemID:          c.PoemID,
			PoemTitle:       c.PoemTitle,
			ConsecutiveDays: c.ConsecutiveDay,
			CreatedAt:       c.CreatedAt,
		}
		items = append(items, item)
	}

	return &adminmodel.AdminCheckinListResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetCheckinStats 获取打卡数据统计（管理后台）
func (s *AdminCheckinService) GetCheckinStats(ctx context.Context, startDate, endDate string) (*adminmodel.AdminCheckinStats, error) {
	stats, err := s.checkinRepo.GetCheckinStats(ctx, startDate, endDate)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询打卡统计失败: %v", err)}
	}

	hotPoems := make([]adminmodel.AdminCheckinHotPoem, 0, len(stats.HotPoems))
	for _, h := range stats.HotPoems {
		hotPoems = append(hotPoems, adminmodel.AdminCheckinHotPoem{
			PoemID:       h.PoemID,
			PoemTitle:    h.PoemTitle,
			CheckinCount: h.CheckinCount,
		})
	}

	return &adminmodel.AdminCheckinStats{
		DailyAvgRate:  stats.DailyAvgRate,
		Retention7d:   stats.Retention7d,
		TotalCheckins: stats.TotalCheckins,
		TotalUsers:    stats.TotalUsers,
		HotPoems:      hotPoems,
	}, nil
}
