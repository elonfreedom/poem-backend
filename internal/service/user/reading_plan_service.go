package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5"

	usermodel "poem-backend/internal/model/user"
	"poem-backend/internal/repository"
)

type ReadingPlanService struct {
	readingPlanRepo *repository.ReadingPlanRepository
}

func NewReadingPlanService(readingPlanRepo *repository.ReadingPlanRepository) *ReadingPlanService {
	return &ReadingPlanService{readingPlanRepo: readingPlanRepo}
}

// validDurations 允许的计划时长（天）
var validDurations = map[int]bool{7: true, 14: true, 30: true, 90: true}

// CreatePlan 创建阅读计划
func (s *ReadingPlanService) CreatePlan(ctx context.Context, userID string, req *usermodel.CreatePlanRequest) (*usermodel.CreatePlanResponse, error) {
	// 校验时长参数
	if !validDurations[req.Duration] {
		return nil, fuego.BadRequestError{Title: "invalid duration", Detail: "计划时长仅限 7、14、30、90 天"}
	}

	// 检查是否有进行中的计划
	existing, _ := s.readingPlanRepo.GetActiveByUserID(ctx, userID)
	if existing != nil {
		return nil, fuego.BadRequestError{Title: "plan conflict", Detail: "已有一个进行中的计划"}
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
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("创建计划失败: %v", err)}
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
	if err := s.readingPlanRepo.UpdateStatus(ctx, userID, planID, "paused"); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("暂停计划失败: %v", err)}
	}
	return nil
}

// ResumePlan 恢复计划
func (s *ReadingPlanService) ResumePlan(ctx context.Context, userID string, planID int) error {
	if err := s.readingPlanRepo.UpdateStatus(ctx, userID, planID, "active"); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("恢复计划失败: %v", err)}
	}
	return nil
}

// LogReading 记录阅读（自动检测计划完成）
func (s *ReadingPlanService) LogReading(ctx context.Context, userID string, poemIDs []int64) (*usermodel.LogReadingResponse, error) {
	// 获取当前计划
	plan, err := s.readingPlanRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "no active plan", Detail: "没有进行中的阅读计划"}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
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
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("记录阅读失败: %v", err)}
	}

	// 自动检测：计划是否已完成（所有天数达标 或 已到结束日期）
	_ = s.CheckAndCompletePlan(ctx, userID, plan)

	return &usermodel.LogReadingResponse{
		TodayCount:    progress.ReadCount,
		TargetCount:   plan.DailyCount,
		IsTodayFinish: progress.ReadCount >= plan.DailyCount,
	}, nil
}

// CheckAndCompletePlan 检查并标记计划为完成
// 触发条件：所有天数的阅读都达标，或已过结束日期
func (s *ReadingPlanService) CheckAndCompletePlan(ctx context.Context, userID string, plan *usermodel.ReadingPlan) error {
	if plan.Status != "active" && plan.Status != "paused" {
		return nil
	}

	// 条件1：已过结束日期
	if time.Now().After(plan.EndDate) {
		return s.readingPlanRepo.UpdateStatus(ctx, userID, plan.PlanID, "completed")
	}

	// 条件2：所有天数都达标
	progressList, err := s.readingPlanRepo.GetProgressByDateRange(ctx, userID, plan.StartDate, plan.EndDate)
	if err != nil {
		return nil // 不阻塞主流程
	}

	totalDays := int(plan.EndDate.Sub(plan.StartDate).Hours()/24) + 1
	completedDays := 0
	for _, p := range progressList {
		if p.ReadCount >= plan.DailyCount {
			completedDays++
		}
	}

	if completedDays >= totalDays {
		return s.readingPlanRepo.UpdateStatus(ctx, userID, plan.PlanID, "completed")
	}

	return nil
}

// GetPlanProgress 获取计划进度
func (s *ReadingPlanService) GetPlanProgress(ctx context.Context, userID string, planID int) (*usermodel.PlanProgressResponse, error) {
	plan, err := s.readingPlanRepo.GetByID(ctx, userID, planID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "plan not found", Detail: fmt.Sprintf("计划不存在: id=%d", planID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}

	// 获取日期范围内的进度
	progressList, err := s.readingPlanRepo.GetProgressByDateRange(ctx, userID, plan.StartDate, plan.EndDate)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询进度失败: %v", err)}
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
