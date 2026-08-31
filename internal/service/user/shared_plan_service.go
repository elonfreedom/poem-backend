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

type SharedPlanService struct {
	sharedPlanRepo *repository.SharedPlanRepository
	checkinRepo    *repository.CheckinRepository
	poemRepo       *repository.PoemRepository
}

func NewSharedPlanService(sharedPlanRepo *repository.SharedPlanRepository, checkinRepo *repository.CheckinRepository, poemRepo *repository.PoemRepository) *SharedPlanService {
	return &SharedPlanService{
		sharedPlanRepo: sharedPlanRepo,
		checkinRepo:    checkinRepo,
		poemRepo:       poemRepo,
	}
}

// ==================== 共享计划管理 ====================

// CreateSharedPlan 创建并发布共享计划
func (s *SharedPlanService) CreateSharedPlan(ctx context.Context, userID string, req *usermodel.CreateSharedPlanRequest) (*usermodel.SharedPlan, error) {
	// 计算总天数
	totalDays := len(req.PoemIDs) / req.DailyCount
	if totalDays == 0 {
		return nil, fuego.BadRequestError{Title: "invalid plan", Detail: fmt.Sprintf("诗文数量不足，至少需要 %d 首", req.DailyCount)}
	}

	now := time.Now()
	plan := &usermodel.SharedPlan{
		CreatorID:      userID,
		Title:          req.Title,
		Description:    req.Description,
		Tags:           req.Tags,
		PoemIDs:        req.PoemIDs,
		DailyCount:     req.DailyCount,
		TotalDays:      totalDays,
		SubscribeCount: 0,
		IsPublished:    true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.sharedPlanRepo.Create(ctx, plan); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("创建计划失败: %v", err)}
	}

	return plan, nil
}

// GetSharedPlan 获取共享计划详情（含诗文详情）
func (s *SharedPlanService) GetSharedPlan(ctx context.Context, id int64) (*usermodel.SharedPlanDetail, error) {
	plan, err := s.sharedPlanRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "plan not found", Detail: fmt.Sprintf("计划不存在: id=%d", id)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}

	// 构建列表项
	item := usermodel.SharedPlanListItem{
		ID:             plan.ID,
		Title:          plan.Title,
		Description:    plan.Description,
		Tags:           plan.Tags,
		DailyCount:     plan.DailyCount,
		TotalDays:      plan.TotalDays,
		SubscribeCount: plan.SubscribeCount,
		CreatorName:    plan.CreatorName,
		CreatedAt:      plan.CreatedAt,
	}

	// 获取诗文详情
	var poems []usermodel.PoemBrief
	for _, pid := range plan.PoemIDs {
		p, err := s.poemRepo.GetByID(ctx, pid)
		if err != nil {
			continue
		}
		poems = append(poems, usermodel.PoemBrief{
			ID:      p.ID,
			Title:   p.Title,
			Author:  p.Author,
			Dynasty: p.Dynasty,
		})
	}

	return &usermodel.SharedPlanDetail{
		SharedPlanListItem: item,
		PoemIDs:            plan.PoemIDs,
		Poems:              poems,
	}, nil
}

// UpdateSharedPlan 更新共享计划
func (s *SharedPlanService) UpdateSharedPlan(ctx context.Context, id int64, userID string, req *usermodel.UpdateSharedPlanRequest) error {
	plan, err := s.sharedPlanRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "plan not found", Detail: fmt.Sprintf("计划不存在: id=%d", id)}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}
	if plan.CreatorID != userID {
		return fuego.ForbiddenError{Title: "access denied", Detail: "无权修改他人计划"}
	}

	// 更新字段
	if req.Title != "" {
		plan.Title = req.Title
	}
	if req.Description != "" {
		plan.Description = req.Description
	}
	if req.Tags != nil {
		plan.Tags = req.Tags
	}
	if len(req.PoemIDs) > 0 {
		plan.PoemIDs = req.PoemIDs
		plan.TotalDays = len(req.PoemIDs) / plan.DailyCount
	}
	if req.DailyCount > 0 {
		plan.DailyCount = req.DailyCount
		if len(plan.PoemIDs) > 0 {
			plan.TotalDays = len(plan.PoemIDs) / plan.DailyCount
		}
	}
	plan.UpdatedAt = time.Now()

	if err := s.sharedPlanRepo.Update(ctx, plan); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新计划失败: %v", err)}
	}
	return nil
}

// PublishPlan 发布计划
func (s *SharedPlanService) PublishPlan(ctx context.Context, id int64, userID string) error {
	plan, err := s.sharedPlanRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "plan not found", Detail: fmt.Sprintf("计划不存在: id=%d", id)}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}
	if plan.CreatorID != userID {
		return fuego.ForbiddenError{Title: "access denied", Detail: "无权操作他人计划"}
	}
	if err := s.sharedPlanRepo.UpdatePublishStatus(ctx, id, userID, true); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("发布计划失败: %v", err)}
	}
	return nil
}

// UnpublishPlan 取消发布
func (s *SharedPlanService) UnpublishPlan(ctx context.Context, id int64, userID string) error {
	plan, err := s.sharedPlanRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "plan not found", Detail: fmt.Sprintf("计划不存在: id=%d", id)}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}
	if plan.CreatorID != userID {
		return fuego.ForbiddenError{Title: "access denied", Detail: "无权操作他人计划"}
	}
	if err := s.sharedPlanRepo.UpdatePublishStatus(ctx, id, userID, false); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("取消发布失败: %v", err)}
	}
	return nil
}

// DeleteSharedPlan 删除共享计划
func (s *SharedPlanService) DeleteSharedPlan(ctx context.Context, id int64, userID string) error {
	if err := s.sharedPlanRepo.Delete(ctx, id, userID); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("删除计划失败: id=%d, error=%v", id, err)}
	}
	return nil
}

// ListSharedPlans 浏览共享库
func (s *SharedPlanService) ListSharedPlans(ctx context.Context, page, pageSize int, keyword string, tags []string, sortBy string) ([]usermodel.SharedPlanListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	list, total, err := s.sharedPlanRepo.List(ctx, page, pageSize, keyword, tags, sortBy)
	if err != nil {
		return nil, 0, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划列表失败: %v", err)}
	}
	return list, total, nil
}

// GetMySharedPlans 获取我创建的计划
func (s *SharedPlanService) GetMySharedPlans(ctx context.Context, userID string) ([]usermodel.SharedPlanListItem, error) {
	list, err := s.sharedPlanRepo.GetMyPlans(ctx, userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询我的计划失败: %v", err)}
	}
	return list, nil
}

// ==================== 订阅管理 ====================

// Subscribe 订阅计划（排队机制）
func (s *SharedPlanService) Subscribe(ctx context.Context, userID string, sharedPlanID int64, startDateStr string) (*usermodel.SubscribeResponse, error) {
	// 检查计划是否存在且已发布
	plan, err := s.sharedPlanRepo.GetByID(ctx, sharedPlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "plan not found", Detail: fmt.Sprintf("计划不存在: id=%d", sharedPlanID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}
	if !plan.IsPublished {
		return nil, fuego.BadRequestError{Title: "plan unavailable", Detail: "计划已下架"}
	}

	// 检查是否已订阅（subscribed 状态）
	existing, _ := s.sharedPlanRepo.GetSubscription(ctx, userID, sharedPlanID)
	if existing != nil && existing.Status == "subscribed" {
		return nil, fuego.BadRequestError{Title: "already subscribed", Detail: "已订阅该计划"}
	}

	// 查找用户当前计划（is_current=true）
	activeSub, _ := s.sharedPlanRepo.GetActiveSubscription(ctx, userID)

	now := time.Now()
	sub := &usermodel.PlanSubscription{
		UserID:       userID,
		SharedPlanID: sharedPlanID,
		Status:       "subscribed",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if activeSub == nil {
		// 无当前计划 → 直接激活
		sub.IsCurrent = true
		sub.QueueOrder = 0
		startDate := now
		if startDateStr != "" {
			if d, err := time.Parse("2006-01-02", startDateStr); err == nil {
				startDate = d
			}
		}
		sub.StartDate = &startDate
	} else {
		// 有当前计划 → 进入队列
		sub.IsCurrent = false
		sub.QueueOrder = 0
		sub.StartDate = nil

		// 获取用户所有 subscribed 订阅，计算 queue_order
		subs, _ := s.sharedPlanRepo.GetSubscriptionsByUserIDWithStatus(ctx, userID, "subscribed")
		sub.QueueOrder = len(subs) // 新订阅排在最后（is_current 占 0）
	}

	if err := s.sharedPlanRepo.CreateSubscription(ctx, sub); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("订阅失败: %v", err)}
	}

	// 订阅数+1
	s.sharedPlanRepo.IncrementSubscribeCount(ctx, sharedPlanID)

	// 构建响应
	resp := &usermodel.SubscribeResponse{
		ID:         sub.ID,
		Status:     "subscribed",
		IsCurrent:  sub.IsCurrent,
		QueueOrder: sub.QueueOrder,
	}

	// 排队计划计算预计开始月份
	if !sub.IsCurrent && activeSub != nil {
		activePlan, _ := s.sharedPlanRepo.GetByID(ctx, activeSub.SharedPlanID)
		if activePlan != nil && activeSub.StartDate != nil {
			endDate := activeSub.StartDate.AddDate(0, 0, activePlan.TotalDays)
			resp.EstimatedMonth = endDate.Format("2006-01")
		}
	}

	return resp, nil
}

// Unsubscribe 取消订阅（软删除）
func (s *SharedPlanService) Unsubscribe(ctx context.Context, userID string, sharedPlanID int64) error {
	// 加载订阅
	existing, err := s.sharedPlanRepo.GetSubscription(ctx, userID, sharedPlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "subscription not found", Detail: "订阅不存在"}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询订阅失败: %v", err)}
	}

	// 软删除：status → cancelled
	if err := s.sharedPlanRepo.SoftDeleteSubscription(ctx, userID, sharedPlanID); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("取消订阅失败: %v", err)}
	}

	// 如果取消的是当前计划，后续排队前移
	if existing.IsCurrent {
		s.sharedPlanRepo.ShiftQueueOrders(ctx, userID, 0)
	} else if existing.QueueOrder > 0 {
		s.sharedPlanRepo.ShiftQueueOrders(ctx, userID, existing.QueueOrder)
	}

	// 订阅数-1
	s.sharedPlanRepo.DecrementSubscribeCount(ctx, sharedPlanID)
	return nil
}

// Activate 激活排队计划
func (s *SharedPlanService) Activate(ctx context.Context, subID int64, userID string, startDateStr string) (*usermodel.SubscriptionItem, error) {
	// 加载订阅
	sub, err := s.sharedPlanRepo.GetSubscriptionByID(ctx, subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "subscription not found", Detail: "订阅不存在"}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询订阅失败: %v", err)}
	}
	if sub.UserID != userID {
		return nil, fuego.ForbiddenError{Title: "access denied", Detail: "无权操作"}
	}
	if sub.Status != "subscribed" || sub.IsCurrent {
		return nil, fuego.BadRequestError{Title: "invalid status", Detail: "计划状态不允许激活"}
	}

	// 解析开始日期
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid date", Detail: "无效的日期格式，应为 YYYY-MM-DD"}
	}

	// 取消当前计划标记
	s.sharedPlanRepo.DeactivateCurrentSubscription(ctx, userID)

	// 激活目标订阅
	if err := s.sharedPlanRepo.ActivateSubscription(ctx, subID, userID, startDate); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("激活失败: %v", err)}
	}

	// 重新计算 queue_order
	s.sharedPlanRepo.RecalculateQueueOrders(ctx, userID)

	// 返回更新后的订阅信息
	return s.getSubscriptionItem(ctx, subID, userID)
}

// ReorderQueue 调整订阅队列顺序（上移/下移）
func (s *SharedPlanService) ReorderQueue(ctx context.Context, subID int64, userID string, direction string) error {
	// 加载订阅
	sub, err := s.sharedPlanRepo.GetSubscriptionByID(ctx, subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "subscription not found", Detail: "订阅不存在"}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询订阅失败: %v", err)}
	}
	if sub.UserID != userID {
		return fuego.ForbiddenError{Title: "access denied", Detail: "无权操作"}
	}
	if sub.Status != "subscribed" || sub.IsCurrent {
		return fuego.BadRequestError{Title: "invalid status", Detail: "只能调整排队中的计划顺序"}
	}

	// 计算目标 queue_order
	var targetOrder int
	if direction == "up" {
		targetOrder = sub.QueueOrder - 1
		if targetOrder < 0 {
			return fuego.BadRequestError{Title: "cannot move", Detail: "已经是第一个"}
		}
	} else {
		targetOrder = sub.QueueOrder + 1
	}

	// 找到目标位置的订阅
	targetSub, err := s.sharedPlanRepo.GetSubscriptionByQueueOrder(ctx, userID, targetOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.BadRequestError{Title: "cannot move", Detail: "无法移动，目标位置不存在"}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询目标订阅失败: %v", err)}
	}

	// 交换 queue_order
	if err := s.sharedPlanRepo.SwapQueueOrders(ctx, userID, subID, targetSub.ID); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("调整顺序失败: %v", err)}
	}

	return nil
}

// getSubscriptionItem 获取单个订阅的详情（用于激活后返回）
func (s *SharedPlanService) getSubscriptionItem(ctx context.Context, subID int64, userID string) (*usermodel.SubscriptionItem, error) {
	sub, err := s.sharedPlanRepo.GetSubscriptionByID(ctx, subID)
	if err != nil {
		return nil, err
	}

	plan, err := s.sharedPlanRepo.GetByID(ctx, sub.SharedPlanID)
	if err != nil {
		return nil, err
	}

	completedPoems, _ := s.checkinRepo.GetCheckedPoemsBySubscription(ctx, subID)

	item := &usermodel.SubscriptionItem{
		ID:             sub.ID,
		SharedPlanID:   sub.SharedPlanID,
		Title:          plan.Title,
		Status:         sub.Status,
		IsCurrent:      sub.IsCurrent,
		QueueOrder:     sub.QueueOrder,
		CompletedPoems: completedPoems,
		TotalPoems:     len(plan.PoemIDs),
	}

	if sub.StartDate != nil {
		d := sub.StartDate.Format("2006-01-02")
		item.StartDate = &d
	}

	return item, nil
}

// SetStartDate 设置开始日期
func (s *SharedPlanService) SetStartDate(ctx context.Context, subID int64, userID string, startDateStr string) error {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return fuego.BadRequestError{Title: "invalid date", Detail: "无效的日期格式，应为 YYYY-MM-DD"}
	}
	if err := s.sharedPlanRepo.UpdateStartDate(ctx, subID, userID, startDate.Format("2006-01-02")); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("设置开始日期失败: %v", err)}
	}
	return nil
}

// GetMySubscriptions 获取我的订阅列表（排队机制新版响应）
func (s *SharedPlanService) GetMySubscriptions(ctx context.Context, userID string) (*usermodel.SubscriptionListResponse, error) {
	// 获取用户所有订阅
	subs, err := s.sharedPlanRepo.GetSubscriptionsByUserID(ctx, userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询订阅列表失败: %v", err)}
	}

	// 获取全局已打卡诗文
	globalChecked, _ := s.checkinRepo.GetGlobalCheckedPoems(ctx, userID)
	globalCheckedMap := make(map[int64]bool)
	for _, pid := range globalChecked {
		globalCheckedMap[pid] = true
	}

	// 查找当前计划（用于计算排队计划的预计月份）
	var currentSub *usermodel.PlanSubscription
	for i := range subs {
		if subs[i].IsCurrent {
			currentSub = &subs[i]
			break
		}
	}

	items := make([]usermodel.SubscriptionItem, 0, len(subs))
	for _, sub := range subs {
		// 获取计划信息
		plan, err := s.sharedPlanRepo.GetByID(ctx, sub.SharedPlanID)
		if err != nil {
			continue
		}

		// 获取本计划已打卡诗文
		completedPoems, _ := s.checkinRepo.GetCheckedPoemsBySubscription(ctx, sub.ID)

		// 计算 actual_total
		actualTotal := 0
		for _, pid := range plan.PoemIDs {
			if !globalCheckedMap[pid] {
				actualTotal++
			}
		}

		item := usermodel.SubscriptionItem{
			ID:             sub.ID,
			SharedPlanID:   sub.SharedPlanID,
			Title:          plan.Title,
			Status:         sub.Status,
			IsCurrent:      sub.IsCurrent,
			QueueOrder:     sub.QueueOrder,
			CompletedPoems: completedPoems,
			TotalPoems:     len(plan.PoemIDs),
			ActualTotal:    actualTotal,
		}

		if sub.StartDate != nil {
			d := sub.StartDate.Format("2006-01-02")
			item.StartDate = &d
		}

		// 排队计划计算预计开始月份
		if !sub.IsCurrent && currentSub != nil && currentSub.StartDate != nil && sub.Status == "subscribed" {
			// 找到当前计划在排队序列中的位置，累加 total_days
			estimatedStart := s.calculateEstimatedStart(ctx, subs, currentSub, plan)
			if estimatedStart != "" {
				item.EstimatedMonth = estimatedStart
			}
		}

		items = append(items, item)
	}

	return &usermodel.SubscriptionListResponse{
		Total: len(items),
		Items: items,
	}, nil
}

// calculateEstimatedStart 计算排队计划的预计开始月份
func (s *SharedPlanService) calculateEstimatedStart(ctx context.Context, subs []usermodel.PlanSubscription, currentSub *usermodel.PlanSubscription, targetPlan *usermodel.SharedPlan) string {
	// 简单估算：当前计划结束月份 + 1
	if currentSub.StartDate == nil {
		return ""
	}
	currentPlan, err := s.sharedPlanRepo.GetByID(ctx, currentSub.SharedPlanID)
	if err != nil {
		return ""
	}
	endDate := currentSub.StartDate.AddDate(0, 0, currentPlan.TotalDays)
	return endDate.Format("2006-01")
}

// ==================== 每日诗文 & 打卡 ====================

// GetTodayPoem 获取今日诗文（含重复诗文过滤）
func (s *SharedPlanService) GetTodayPoem(ctx context.Context, subID int64, userID string) (*usermodel.TodayPoemResponse, error) {
	// 获取订阅信息
	sub, err := s.sharedPlanRepo.GetSubscriptionByID(ctx, subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "subscription not found", Detail: fmt.Sprintf("订阅不存在: id=%d", subID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询订阅失败: %v", err)}
	}
	if sub.UserID != userID {
		return nil, fuego.ForbiddenError{Title: "access denied", Detail: "无权访问"}
	}
	if !sub.IsCurrent {
		return nil, fuego.BadRequestError{Title: "not current plan", Detail: "请先激活该计划"}
	}
	if sub.StartDate == nil {
		return nil, fuego.BadRequestError{Title: "no start date", Detail: "计划未设置开始日期"}
	}

	// 获取计划信息
	plan, err := s.sharedPlanRepo.GetByID(ctx, sub.SharedPlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "plan not found", Detail: "计划不存在"}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}

	// 获取全局已打卡诗文并过滤
	globalChecked, _ := s.checkinRepo.GetGlobalCheckedPoems(ctx, userID)
	globalCheckedMap := make(map[int64]bool)
	for _, pid := range globalChecked {
		globalCheckedMap[pid] = true
	}

	// 过滤后的有效诗文列表
	effectivePoemIDs := make([]int64, 0, len(plan.PoemIDs))
	for _, pid := range plan.PoemIDs {
		if !globalCheckedMap[pid] {
			effectivePoemIDs = append(effectivePoemIDs, pid)
		}
	}

	// 计算今天是第几天
	today := time.Now()
	dayNumber := int(today.Sub(*sub.StartDate).Hours()/24) + 1
	if dayNumber < 1 {
		dayNumber = 1
	}

	// 从有效诗文中获取今日诗文
	startIdx := (dayNumber - 1) * plan.DailyCount
	endIdx := startIdx + plan.DailyCount
	actualTotal := len(effectivePoemIDs)

	if startIdx >= actualTotal {
		// 计划已完成
		return &usermodel.TodayPoemResponse{
			DayNumber:    dayNumber,
			Date:         today.Format("2006-01-02"),
			Poems:        []usermodel.Poem{},
			TotalDays:    plan.TotalDays,
			ProgressRate: 100,
		}, nil
	}
	if endIdx > actualTotal {
		endIdx = actualTotal
	}

	todayPoemIDs := effectivePoemIDs[startIdx:endIdx]

	// 查询诗文详情
	var poems []usermodel.Poem
	for _, pid := range todayPoemIDs {
		p, err := s.poemRepo.GetByID(ctx, pid)
		if err != nil {
			continue // 诗文可能已删除，跳过
		}

		// 查询该诗文的上次打卡信息
		lastCheckin, _ := s.getLastCheckinInfo(ctx, userID, pid)

		poems = append(poems, usermodel.Poem{
			ID:           p.ID,
			Title:        p.Title,
			Author:       p.Author,
			Dynasty:      p.Dynasty,
			Content:      p.Content,
			Translation:  p.Translation,
			Appreciation: p.Appreciation,
			Tags:         p.Tags,
			LastCheckin:  lastCheckin,
		})
	}

	// 检查今日是否已打卡
	checkedIn := false
	var checkedAt *string
	checkin, _ := s.checkinRepo.GetCheckinBySubDay(ctx, subID, dayNumber)
	if checkin != nil {
		checkedIn = true
		t := checkin.CreatedAt.Format("2006-01-02T15:04:05")
		checkedAt = &t
	}

	// 计算完成率
	completedDays, _ := s.checkinRepo.CountCheckinsBySubscription(ctx, subID)
	progressRate := float64(completedDays) / float64(plan.TotalDays) * 100

	return &usermodel.TodayPoemResponse{
		DayNumber:    dayNumber,
		Date:         today.Format("2006-01-02"),
		Poems:        poems,
		IsCheckedIn:  checkedIn,
		CheckedAt:    checkedAt,
		TotalDays:    plan.TotalDays,
		ProgressRate: progressRate,
	}, nil
}

// getLastCheckinInfo 获取诗文的上次打卡信息
func (s *SharedPlanService) getLastCheckinInfo(ctx context.Context, userID string, poemID int64) (*usermodel.CheckinInfo, error) {
	// 先从 checkins 表查询（含订阅打卡）
	date, planTitle, daysAgo, err := s.checkinRepo.GetLastCheckinByPoem(ctx, userID, poemID)
	if err == nil {
		return &usermodel.CheckinInfo{
			Date:      date,
			PlanTitle: planTitle,
			DaysAgo:   daysAgo,
		}, nil
	}

	// 再从 reading_progress 查询（自建计划阅读打卡）
	date, planTitle, daysAgo, err = s.sharedPlanRepo.GetLastCheckinByPoem(ctx, userID, poemID)
	if err == nil {
		return &usermodel.CheckinInfo{
			Date:      date,
			PlanTitle: planTitle,
			DaysAgo:   daysAgo,
		}, nil
	}

	return nil, nil
}

// SkipDay 跳过当前天，返回下一首未打卡的诗文
func (s *SharedPlanService) SkipDay(ctx context.Context, subID int64, userID string, currentDay int) (*usermodel.SkipDayResponse, error) {
	// 获取订阅信息
	sub, err := s.sharedPlanRepo.GetSubscriptionByID(ctx, subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "subscription not found", Detail: fmt.Sprintf("订阅不存在: id=%d", subID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询订阅失败: %v", err)}
	}
	if sub.UserID != userID {
		return nil, fuego.ForbiddenError{Title: "access denied", Detail: "无权访问"}
	}

	// 获取计划信息
	plan, err := s.sharedPlanRepo.GetByID(ctx, sub.SharedPlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "plan not found", Detail: "计划不存在"}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}

	// 从 currentDay+1 开始找第一个未打卡的天
	nextDay := currentDay + 1
	for nextDay <= plan.TotalDays {
		// 检查该天是否已打卡
		checkin, _ := s.checkinRepo.GetCheckinBySubDay(ctx, subID, nextDay)
		if checkin == nil {
			// 未打卡，获取该天的第一首诗文
			startIdx := (nextDay - 1) * plan.DailyCount
			if startIdx < len(plan.PoemIDs) {
				firstPoemID := plan.PoemIDs[startIdx]
				p, err := s.poemRepo.GetByID(ctx, firstPoemID)
				if err == nil {
					// 查询该诗文的打卡信息
					lastCheckin, _ := s.getLastCheckinInfo(ctx, userID, firstPoemID)
					return &usermodel.SkipDayResponse{
						NextDay: nextDay,
						Poem: usermodel.Poem{
							ID:           p.ID,
							Title:        p.Title,
							Author:       p.Author,
							Dynasty:      p.Dynasty,
							Content:      p.Content,
							Translation:  p.Translation,
							Appreciation: p.Appreciation,
							Tags:         p.Tags,
							LastCheckin:  lastCheckin,
						},
					}, nil
				}
			}
		}
		nextDay++
	}

	// 所有天都已完成打卡
	return nil, fuego.BadRequestError{Title: "all completed", Detail: "所有天都已完成打卡"}
}

// Checkin 打卡
func (s *SharedPlanService) Checkin(ctx context.Context, subID int64, userID string, poemIDs []int64) (*usermodel.CheckinResponse, error) {
	// 获取订阅信息
	sub, err := s.sharedPlanRepo.GetSubscriptionByID(ctx, subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "subscription not found", Detail: fmt.Sprintf("订阅不存在: id=%d", subID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询订阅失败: %v", err)}
	}
	if sub.UserID != userID {
		return nil, fuego.ForbiddenError{Title: "access denied", Detail: "无权访问"}
	}
	if !sub.IsCurrent {
		return nil, fuego.BadRequestError{Title: "not current plan", Detail: "请先激活该计划"}
	}
	if sub.StartDate == nil {
		return nil, fuego.BadRequestError{Title: "no start date", Detail: "计划未设置开始日期"}
	}

	// 获取计划信息
	plan, err := s.sharedPlanRepo.GetByID(ctx, sub.SharedPlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "plan not found", Detail: "计划不存在"}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}

	// 计算今天是第几天
	today := time.Now()
	dayNumber := int(today.Sub(*sub.StartDate).Hours()/24) + 1
	if dayNumber < 1 {
		dayNumber = 1
	}

	// 创建打卡记录（写入统一的 checkins 表）
	if err := s.checkinRepo.CreateSubscriptionCheckin(ctx, userID, subID, dayNumber, today, poemIDs, today); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("打卡失败: %v", err)}
	}

	// 计算完成率
	completedDays, _ := s.checkinRepo.CountCheckinsBySubscription(ctx, subID)
	progressRate := float64(completedDays) / float64(plan.TotalDays) * 100

	// 检查是否全部完成（actual_total）
	globalChecked, _ := s.checkinRepo.GetGlobalCheckedPoems(ctx, userID)
	globalCheckedMap := make(map[int64]bool)
	for _, pid := range globalChecked {
		globalCheckedMap[pid] = true
	}
	actualTotal := 0
	for _, pid := range plan.PoemIDs {
		if !globalCheckedMap[pid] {
			actualTotal++
		}
	}
	actualTotalDays := actualTotal / plan.DailyCount
	if actualTotalDays < 1 {
		actualTotalDays = 1
	}

	resp := &usermodel.CheckinResponse{
		DayNumber:     dayNumber,
		IsTodayFinish: true,
		CompletedDays: completedDays,
		TotalDays:     plan.TotalDays,
		ProgressRate:  progressRate,
	}

	// 如果 actual_total 全部完成 → 标记 completed
	if completedDays >= actualTotalDays {
		s.sharedPlanRepo.UpdateSubscriptionStatus(ctx, subID, userID, "completed")
		s.sharedPlanRepo.DeactivateCurrentSubscription(ctx, userID)
		s.sharedPlanRepo.ShiftQueueOrders(ctx, userID, 0)
		_ = resp
	}

	return resp, nil
}

// GetSubscriptionProgress 获取订阅进度（含每日打卡诗文标题，支持热力图）
func (s *SharedPlanService) GetSubscriptionProgress(ctx context.Context, subID int64, userID string) (*usermodel.PlanProgressResponse, error) {
	// 获取订阅信息
	sub, err := s.sharedPlanRepo.GetSubscriptionByID(ctx, subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "subscription not found", Detail: fmt.Sprintf("订阅不存在: id=%d", subID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询订阅失败: %v", err)}
	}
	if sub.UserID != userID {
		return nil, fuego.ForbiddenError{Title: "access denied", Detail: "无权访问"}
	}

	// 获取计划信息
	plan, err := s.sharedPlanRepo.GetByID(ctx, sub.SharedPlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "plan not found", Detail: "计划不存在"}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询计划失败: %v", err)}
	}

	// 获取打卡记录
	checkins, err := s.checkinRepo.GetCheckinsBySubscription(ctx, subID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("获取打卡记录失败: %v", err)}
	}

	// 构建打卡映射：day_number -> 第一首诗文标题
	checkinMap := make(map[int]string)
	for _, c := range checkins {
		if pid := c.PoemIDOrFirst(); pid != nil {
			if p, err := s.poemRepo.GetByID(ctx, *pid); err == nil {
				checkinMap[*c.DayNumber] = p.Title
			}
		}
	}

	// 构建每日进度
	dailyProgress := make([]usermodel.DailyProgress, 0, plan.TotalDays)
	baseDate := time.Now()
	if sub.StartDate != nil {
		baseDate = *sub.StartDate
	}
	for day := 1; day <= plan.TotalDays; day++ {
		poemTitle := ""
		isReached := false
		if title, ok := checkinMap[day]; ok {
			isReached = true
			poemTitle = title
		}
		dailyProgress = append(dailyProgress, usermodel.DailyProgress{
			Date:      baseDate.AddDate(0, 0, day-1),
			ReadCount: 0,
			Target:    plan.DailyCount,
			IsReached: isReached,
			PoemTitle: poemTitle,
		})
	}

	completedDays := len(checkins)
	progressRate := float64(completedDays) / float64(plan.TotalDays) * 100

	// 处理可空的 start_date
	var startDate, endDate time.Time
	if sub.StartDate != nil {
		startDate = *sub.StartDate
		endDate = sub.StartDate.AddDate(0, 0, plan.TotalDays-1)
	}

	return &usermodel.PlanProgressResponse{
		PlanID:         int(sub.SharedPlanID),
		DailyCount:     plan.DailyCount,
		StartDate:      startDate,
		EndDate:        endDate,
		Status:         sub.Status,
		TotalDays:      plan.TotalDays,
		CompletedDays:  completedDays,
		CompletionRate: progressRate,
		DailyProgress:  dailyProgress,
	}, nil
}

// GetCheckins 获取订阅的打卡记录列表
func (s *SharedPlanService) GetCheckins(ctx context.Context, subID int64, userID string) (*usermodel.CheckinsResponse, error) {
	// 获取订阅信息
	sub, err := s.sharedPlanRepo.GetSubscriptionByID(ctx, subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "subscription not found", Detail: fmt.Sprintf("订阅不存在: id=%d", subID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询订阅失败: %v", err)}
	}
	if sub.UserID != userID {
		return nil, fuego.ForbiddenError{Title: "access denied", Detail: "无权访问"}
	}

	// 获取打卡记录
	checkins, err := s.checkinRepo.GetCheckinsBySubscription(ctx, subID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("获取打卡记录失败: %v", err)}
	}

	// 构建响应
	items := make([]usermodel.CheckinRecord, 0, len(checkins))
	for _, c := range checkins {
		poemTitle := ""
		if pid := c.PoemIDOrFirst(); pid != nil {
			if p, err := s.poemRepo.GetByID(ctx, *pid); err == nil {
				poemTitle = p.Title
			}
		}
		dayNumber := 0
		if c.DayNumber != nil {
			dayNumber = *c.DayNumber
		}
		items = append(items, usermodel.CheckinRecord{
			Date:      c.Date.Format("2006-01-02"),
			DayNumber: dayNumber,
			PoemTitle: poemTitle,
		})
	}

	return &usermodel.CheckinsResponse{
		Total: len(items),
		Items: items,
	}, nil
}

// PauseSubscription 暂停订阅（已废弃，新模型使用取消+重新订阅）
// Deprecated: 排队机制不再支持暂停，用户取消后重新订阅
func (s *SharedPlanService) PauseSubscription(ctx context.Context, subID int64, userID string) error {
	return fuego.BadRequestError{Title: "deprecated", Detail: "暂停功能已废弃，请取消订阅后重新订阅"}
}

// ResumeSubscription 恢复订阅（已废弃）
// Deprecated: 排队机制不再支持恢复，用户取消后重新订阅
func (s *SharedPlanService) ResumeSubscription(ctx context.Context, subID int64, userID string) error {
	return fuego.BadRequestError{Title: "deprecated", Detail: "恢复功能已废弃，请取消订阅后重新订阅"}
}
