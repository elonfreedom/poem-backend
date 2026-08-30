package user

import (
	"context"
	"fmt"
	"time"

	"github.com/go-fuego/fuego"

	usermodel "poem-backend/internal/model/user"
	"poem-backend/internal/repository"
)

type CheckinService struct {
	checkinRepo     *repository.CheckinRepository
	readingPlanRepo *repository.ReadingPlanRepository
}

func NewCheckinService(checkinRepo *repository.CheckinRepository, readingPlanRepo *repository.ReadingPlanRepository) *CheckinService {
	return &CheckinService{
		checkinRepo:     checkinRepo,
		readingPlanRepo: readingPlanRepo,
	}
}

// Checkin 打卡（需有活跃阅读计划）
func (s *CheckinService) Checkin(ctx context.Context, userID string, date string, poemID *int64) (*usermodel.CheckInResponse, error) {
	// 校验：用户必须有活跃的阅读计划才能打卡
	activePlan, err := s.readingPlanRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询阅读计划失败: %v", err)}
	}
	if activePlan == nil {
		return nil, fuego.BadRequestError{Title: "no active plan", Detail: "请先创建阅读计划后再打卡"}
	}

	// 解析日期，默认今天
	checkinDate := time.Now()
	if date != "" {
		if d, err := time.Parse("2006-01-02", date); err == nil {
			checkinDate = d
		}
	}

	// 检查当天是否已打卡
	existing, _ := s.checkinRepo.GetByDate(ctx, userID, checkinDate)
	if existing != nil {
		return &usermodel.CheckInResponse{
			Date:           existing.Date,
			ConsecutiveDay: existing.ConsecutiveDay,
		}, nil
	}

	// 计算连续天数
	consecutiveDay := 1
	lastCheckin, _ := s.checkinRepo.GetLastCheckIn(ctx, userID)
	if lastCheckin != nil {
		yesterday := checkinDate.AddDate(0, 0, -1)
		if lastCheckin.Date.Equal(yesterday) || lastCheckin.Date.After(yesterday) {
			consecutiveDay = lastCheckin.ConsecutiveDay + 1
		}
	}

	// 创建打卡记录
	checkin := &usermodel.CheckIn{
		UserID:         userID,
		Date:           checkinDate,
		ConsecutiveDay: consecutiveDay,
		CreatedAt:      time.Now(),
	}
	if poemID != nil {
		checkin.PoemID = poemID
	}
	if err := s.checkinRepo.Create(ctx, checkin); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("打卡失败: %v", err)}
	}

	// 更新统计
	stats, _ := s.checkinRepo.GetStats(ctx, userID)
	if stats == nil {
		stats = &usermodel.CheckInStats{
			UserID: userID,
		}
	}
	stats.TotalDays++
	stats.ConsecutiveDay = consecutiveDay
	if consecutiveDay > stats.MaxConsecutive {
		stats.MaxConsecutive = consecutiveDay
	}
	stats.LastCheckIn = checkinDate

	if err := s.checkinRepo.UpsertStats(ctx, stats); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新打卡统计失败: %v", err)}
	}

	return &usermodel.CheckInResponse{
		Date:           checkin.Date,
		ConsecutiveDay: checkin.ConsecutiveDay,
	}, nil
}

// GetStats 获取打卡统计
func (s *CheckinService) GetStats(ctx context.Context, userID string) (*usermodel.CheckInStatsResponse, error) {
	stats, err := s.checkinRepo.GetStats(ctx, userID)
	if err != nil {
		// 返回空统计
		return &usermodel.CheckInStatsResponse{
			TotalDays:      0,
			ConsecutiveDay: 0,
			MaxConsecutive: 0,
		}, nil
	}

	return &usermodel.CheckInStatsResponse{
		TotalDays:      stats.TotalDays,
		ConsecutiveDay: stats.ConsecutiveDay,
		MaxConsecutive: stats.MaxConsecutive,
		LastCheckIn:    stats.LastCheckIn,
	}, nil
}

// GetCheckinList 获取打卡记录列表
func (s *CheckinService) GetCheckinList(ctx context.Context, userID string, page, pageSize int, startDate, endDate string) (*usermodel.CheckInListResponse, error) {
	// 解析日期范围
	var start, end time.Time
	if startDate != "" {
		if d, err := time.Parse("2006-01-02", startDate); err == nil {
			start = d
		}
	}
	if endDate != "" {
		if d, err := time.Parse("2006-01-02", endDate); err == nil {
			end = d
		}
	}

	checkins, total, err := s.checkinRepo.List(ctx, userID, page, pageSize, start, end)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询打卡记录失败: %v", err)}
	}

	var list []usermodel.CheckInResponse
	for _, c := range checkins {
		list = append(list, usermodel.CheckInResponse{
			Date:           c.Date,
			ConsecutiveDay: c.ConsecutiveDay,
			PoemID:         c.PoemID,
			PoemTitle:      c.PoemTitle,
		})
	}

	return &usermodel.CheckInListResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetCalendar 获取打卡日历
func (s *CheckinService) GetCalendar(ctx context.Context, userID string, year, month int) (*usermodel.CheckInCalendarResponse, error) {
	days, err := s.checkinRepo.GetCheckInDates(ctx, userID, year, month)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询打卡日历失败: %v", err)}
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

	calendarDays := make([]usermodel.CalendarDay, 0, daysInMonth)
	for i := 1; i <= daysInMonth; i++ {
		calendarDays = append(calendarDays, usermodel.CalendarDay{
			Day:       i,
			IsChecked: checkinDays[i],
		})
	}

	return &usermodel.CheckInCalendarResponse{
		Year:  year,
		Month: month,
		Days:  calendarDays,
	}, nil
}

// GetRanking 获取排行榜
func (s *CheckinService) GetRanking(ctx context.Context, userID string) (*usermodel.RankingResponse, error) {
	items, err := s.checkinRepo.GetRanking(ctx, 100)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询排行榜失败: %v", err)}
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

	return &usermodel.RankingResponse{
		Total:            len(items),
		MyRank:           myRank,
		MyConsecutiveDay: myConsecutiveDay,
		List:             items,
	}, nil
}
