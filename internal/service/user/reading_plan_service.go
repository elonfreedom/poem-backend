package user

import (
	"context"
	"fmt"
	"time"

	usermodel "poem-backend/internal/model/user"
	"poem-backend/internal/repository"
)

type ReadingPlanService struct {
	readingPlanRepo *repository.ReadingPlanRepository
}

func NewReadingPlanService(readingPlanRepo *repository.ReadingPlanRepository) *ReadingPlanService {
	return &ReadingPlanService{readingPlanRepo: readingPlanRepo}
}

// CreatePlan 创建阅读计划
func (s *ReadingPlanService) CreatePlan(ctx context.Context, userID string, req *usermodel.CreatePlanRequest) (*usermodel.CreatePlanResponse, error) {
	// 检查是否有进行中的计划
	existing, _ := s.readingPlanRepo.GetActiveByUserID(ctx, userID)
	if existing != nil {
		return nil, fmt.Errorf("you already have an active plan")
	}

	// 解析开始日期，默认今天
	startDate := time.Now()
	if req.StartDate != "" {
		if d, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = d
		}
	}
	endDate := startDate.AddDate(0, 0, req.Duration-1)

	plan := &usermodel.ReadingPlan{
		UserID:     userID,
		Title:      req.Title,
		DailyCount: req.DailyCount,
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.readingPlanRepo.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	return &usermodel.CreatePlanResponse{
		PlanID:     plan.PlanID,
		Title:      plan.Title,
		DailyCount: plan.DailyCount,
		StartDate:  plan.StartDate,
		EndDate:    plan.EndDate,
		Status:     plan.Status,
	}, nil
}

// GetCurrentPlan 获取当前计划
func (s *ReadingPlanService) GetCurrentPlan(ctx context.Context, userID string) (*usermodel.CurrentPlanResponse, error) {
	plan, err := s.readingPlanRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, nil // 没有进行中的计划
	}

	// 获取今日进度
	today := time.Now()
	progress, _ := s.readingPlanRepo.GetProgress(ctx, userID, today)

	todayCount := 0
	if progress != nil {
		todayCount = progress.ReadCount
	}

	// 计算完成率
	progressList, _ := s.readingPlanRepo.GetProgressByDateRange(ctx, userID, plan.StartDate, plan.EndDate)
	completedDays := 0
	for _, p := range progressList {
		if p.ReadCount >= plan.DailyCount {
			completedDays++
		}
	}

	totalDays := int(plan.EndDate.Sub(plan.StartDate).Hours()/24) + 1
	completionRate := 0.0
	if totalDays > 0 {
		completionRate = float64(completedDays) / float64(totalDays) * 100
	}

	return &usermodel.CurrentPlanResponse{
		Plan: usermodel.CreatePlanResponse{
			PlanID:     plan.PlanID,
			DailyCount: plan.DailyCount,
			StartDate:  plan.StartDate,
			EndDate:    plan.EndDate,
			Status:     plan.Status,
		},
		TodayCount:     todayCount,
		IsTodayFinish:  todayCount >= plan.DailyCount,
		CompletionRate: completionRate,
	}, nil
}

// PausePlan 暂停计划
func (s *ReadingPlanService) PausePlan(ctx context.Context, userID string, planID int) error {
	return s.readingPlanRepo.UpdateStatus(ctx, userID, planID, "paused")
}

// ResumePlan 恢复计划
func (s *ReadingPlanService) ResumePlan(ctx context.Context, userID string, planID int) error {
	return s.readingPlanRepo.UpdateStatus(ctx, userID, planID, "active")
}

// LogReading 记录阅读
func (s *ReadingPlanService) LogReading(ctx context.Context, userID string, poemIDs []int64) (*usermodel.LogReadingResponse, error) {
	// 获取当前计划
	plan, err := s.readingPlanRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("no active plan found")
	}

	today := time.Now()

	// 获取或创建今日进度
	progress, _ := s.readingPlanRepo.GetProgress(ctx, userID, today)
	if progress == nil {
		progress = &usermodel.ReadingProgress{
			UserID:    userID,
			Date:      today,
			CreatedAt: time.Now(),
		}
	}

	// 合并诗歌 ID（去重）
	existingPoems := make(map[int64]bool)
	for _, id := range progress.PoemIDs {
		existingPoems[id] = true
	}
	for _, id := range poemIDs {
		existingPoems[id] = true
	}

	mergedPoems := make([]int64, 0, len(existingPoems))
	for id := range existingPoems {
		mergedPoems = append(mergedPoems, id)
	}

	progress.PoemIDs = mergedPoems
	progress.ReadCount = len(mergedPoems)

	if err := s.readingPlanRepo.UpsertProgress(ctx, progress); err != nil {
		return nil, fmt.Errorf("failed to log reading: %w", err)
	}

	return &usermodel.LogReadingResponse{
		TodayCount:    progress.ReadCount,
		TargetCount:   plan.DailyCount,
		IsTodayFinish: progress.ReadCount >= plan.DailyCount,
	}, nil
}

// GetPlanProgress 获取计划进度
func (s *ReadingPlanService) GetPlanProgress(ctx context.Context, userID string, planID int) (*usermodel.PlanProgressResponse, error) {
	plan, err := s.readingPlanRepo.GetByID(ctx, userID, planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// 获取日期范围内的进度
	progressList, err := s.readingPlanRepo.GetProgressByDateRange(ctx, userID, plan.StartDate, plan.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get progress: %w", err)
	}

	// 构建每日进度
	dailyProgress := make([]usermodel.DailyProgress, 0, len(progressList))
	completedDays := 0
	for _, p := range progressList {
		isReached := p.ReadCount >= plan.DailyCount
		if isReached {
			completedDays++
		}
		dailyProgress = append(dailyProgress, usermodel.DailyProgress{
			Date:      p.Date,
			ReadCount: p.ReadCount,
			Target:    plan.DailyCount,
			IsReached: isReached,
		})
	}

	totalDays := int(plan.EndDate.Sub(plan.StartDate).Hours()/24) + 1
	completionRate := 0.0
	if totalDays > 0 {
		completionRate = float64(completedDays) / float64(totalDays) * 100
	}

	return &usermodel.PlanProgressResponse{
		PlanID:         plan.PlanID,
		DailyCount:     plan.DailyCount,
		StartDate:      plan.StartDate,
		EndDate:        plan.EndDate,
		Status:         plan.Status,
		TotalDays:      totalDays,
		CompletedDays:  completedDays,
		CompletionRate: completionRate,
		DailyProgress:  dailyProgress,
	}, nil
}
