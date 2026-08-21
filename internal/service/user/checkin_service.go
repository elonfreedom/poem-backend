package user

import (
	"context"
	"fmt"
	"time"

	"poem-backend/internal/model"
	"poem-backend/internal/repository"
)

type CheckinService struct {
	checkinRepo *repository.CheckinRepository
}

func NewCheckinService(checkinRepo *repository.CheckinRepository) *CheckinService {
	return &CheckinService{checkinRepo: checkinRepo}
}

// Checkin 打卡
func (s *CheckinService) Checkin(ctx context.Context, userID string) (*model.CheckInResponse, error) {
	today := time.Now()

	// 检查今天是否已打卡
	existing, _ := s.checkinRepo.GetByDate(ctx, userID, today)
	if existing != nil {
		return &model.CheckInResponse{
			Date:           existing.Date,
			ConsecutiveDay: existing.ConsecutiveDay,
		}, nil
	}

	// 计算连续天数
	consecutiveDay := 1
	lastCheckin, _ := s.checkinRepo.GetLastCheckIn(ctx, userID)
	if lastCheckin != nil {
		yesterday := today.AddDate(0, 0, -1)
		if lastCheckin.Date.Equal(yesterday) || lastCheckin.Date.After(yesterday) {
			consecutiveDay = lastCheckin.ConsecutiveDay + 1
		}
	}

	// 创建打卡记录
	checkin := &model.CheckIn{
		UserID:         userID,
		Date:           today,
		ConsecutiveDay: consecutiveDay,
		CreatedAt:      time.Now(),
	}
	if err := s.checkinRepo.Create(ctx, checkin); err != nil {
		return nil, fmt.Errorf("failed to checkin: %w", err)
	}

	// 更新统计
	stats, _ := s.checkinRepo.GetStats(ctx, userID)
	if stats == nil {
		stats = &model.CheckInStats{
			UserID: userID,
		}
	}
	stats.TotalDays++
	stats.ConsecutiveDay = consecutiveDay
	if consecutiveDay > stats.MaxConsecutive {
		stats.MaxConsecutive = consecutiveDay
	}
	stats.LastCheckIn = today

	if err := s.checkinRepo.UpsertStats(ctx, stats); err != nil {
		return nil, fmt.Errorf("failed to update stats: %w", err)
	}

	return &model.CheckInResponse{
		Date:           checkin.Date,
		ConsecutiveDay: checkin.ConsecutiveDay,
	}, nil
}

// GetStats 获取打卡统计
func (s *CheckinService) GetStats(ctx context.Context, userID string) (*model.CheckInStatsResponse, error) {
	stats, err := s.checkinRepo.GetStats(ctx, userID)
	if err != nil {
		// 返回空统计
		return &model.CheckInStatsResponse{
			TotalDays:      0,
			ConsecutiveDay: 0,
			MaxConsecutive: 0,
		}, nil
	}

	return &model.CheckInStatsResponse{
		TotalDays:      stats.TotalDays,
		ConsecutiveDay: stats.ConsecutiveDay,
		MaxConsecutive: stats.MaxConsecutive,
		LastCheckIn:    stats.LastCheckIn,
	}, nil
}

// GetCheckinList 获取打卡记录列表
func (s *CheckinService) GetCheckinList(ctx context.Context, userID string, page, pageSize int) (*model.CheckInListResponse, error) {
	checkins, total, err := s.checkinRepo.List(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkins: %w", err)
	}

	var list []model.CheckInResponse
	for _, c := range checkins {
		list = append(list, model.CheckInResponse{
			Date:           c.Date,
			ConsecutiveDay: c.ConsecutiveDay,
		})
	}

	return &model.CheckInListResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetCalendar 获取打卡日历
func (s *CheckinService) GetCalendar(ctx context.Context, userID string, year, month int) (*model.CheckInCalendarResponse, error) {
	days, err := s.checkinRepo.GetCheckInDates(ctx, userID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}

	// 构建打卡日期集合
	checkinDays := make(map[int]bool)
	for _, d := range days {
		checkinDays[d] = true
	}

	// 获取当月天数
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	daysInMonth := lastDay.Day()

	calendarDays := make([]model.CalendarDay, 0, daysInMonth)
	for i := 1; i <= daysInMonth; i++ {
		calendarDays = append(calendarDays, model.CalendarDay{
			Day:       i,
			IsChecked: checkinDays[i],
		})
	}

	return &model.CheckInCalendarResponse{
		Year:  year,
		Month: month,
		Days:  calendarDays,
	}, nil
}

// GetRanking 获取排行榜
func (s *CheckinService) GetRanking(ctx context.Context, userID string) (*model.RankingResponse, error) {
	items, err := s.checkinRepo.GetRanking(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get ranking: %w", err)
	}

	// 获取用户自己的排名和连续天数
	stats, _ := s.checkinRepo.GetStats(ctx, userID)
	myConsecutiveDay := 0
	if stats != nil {
		myConsecutiveDay = stats.ConsecutiveDay
	}

	// 计算用户排名
	myRank := 0
	for i, item := range items {
		if item.ConsecutiveDay <= myConsecutiveDay {
			myRank = i + 1
			break
		}
	}
	if myRank == 0 && myConsecutiveDay > 0 {
		myRank = len(items) + 1
	}

	return &model.RankingResponse{
		Total:            len(items),
		MyRank:           myRank,
		MyConsecutiveDay: myConsecutiveDay,
		List:             items,
	}, nil
}
