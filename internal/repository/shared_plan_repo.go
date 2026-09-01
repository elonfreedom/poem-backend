package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	usermodel "poem-backend/internal/model/user"
)

type SharedPlanRepository struct {
	db *pgxpool.Pool
}

func NewSharedPlanRepository(db *pgxpool.Pool) *SharedPlanRepository {
	return &SharedPlanRepository{db: db}
}

// ==================== 共享计划 CRUD ====================

// Create 创建共享计划
func (r *SharedPlanRepository) Create(ctx context.Context, plan *usermodel.SharedPlan) error {
	query := `
		INSERT INTO shared_plans (creator_id, title, description, tags, poem_ids, daily_count, total_days, subscribe_count, is_published, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		plan.CreatorID, plan.Title, plan.Description, plan.Tags, plan.PoemIDs,
		plan.DailyCount, plan.TotalDays, plan.SubscribeCount, plan.IsPublished,
		plan.CreatedAt, plan.UpdatedAt,
	).Scan(&plan.ID)
}

// GetByID 根据ID获取共享计划
func (r *SharedPlanRepository) GetByID(ctx context.Context, id int64) (*usermodel.SharedPlan, error) {
	query := `
		SELECT sp.id, sp.creator_id, u.nickname as creator_name, sp.title, sp.description, sp.tags,
		       sp.poem_ids, sp.daily_count, sp.total_days, sp.subscribe_count, sp.is_published, sp.created_at, sp.updated_at
		FROM shared_plans sp
		JOIN users u ON sp.creator_id = u.id
		WHERE sp.id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	var plan usermodel.SharedPlan
	err := row.Scan(&plan.ID, &plan.CreatorID, &plan.CreatorName, &plan.Title, &plan.Description,
		&plan.Tags, &plan.PoemIDs, &plan.DailyCount, &plan.TotalDays, &plan.SubscribeCount,
		&plan.IsPublished, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// Update 更新共享计划
func (r *SharedPlanRepository) Update(ctx context.Context, plan *usermodel.SharedPlan) error {
	query := `
		UPDATE shared_plans SET
			title = $1, description = $2, tags = $3, poem_ids = $4,
			daily_count = $5, total_days = $6, updated_at = $7
		WHERE id = $8 AND creator_id = $9
	`
	_, err := r.db.Exec(ctx, query,
		plan.Title, plan.Description, plan.Tags, plan.PoemIDs,
		plan.DailyCount, plan.TotalDays, plan.UpdatedAt, plan.ID, plan.CreatorID)
	return err
}

// UpdatePublishStatus 更新发布状态
func (r *SharedPlanRepository) UpdatePublishStatus(ctx context.Context, id int64, creatorID string, isPublished bool) error {
	query := `UPDATE shared_plans SET is_published = $1, updated_at = NOW() WHERE id = $2 AND creator_id = $3`
	_, err := r.db.Exec(ctx, query, isPublished, id, creatorID)
	return err
}

// Delete 删除共享计划（仅创建者）
func (r *SharedPlanRepository) Delete(ctx context.Context, id int64, creatorID string) error {
	// 检查是否有订阅者，有则不删除仅取消发布
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM plan_subscriptions WHERE shared_plan_id = $1`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		// 有订阅者，仅取消发布
		return r.UpdatePublishStatus(ctx, id, creatorID, false)
	}
	// 无订阅者，直接删除
	_, err = r.db.Exec(ctx, `DELETE FROM shared_plans WHERE id = $1 AND creator_id = $2`, id, creatorID)
	return err
}

// List 获取共享计划列表（分页、搜索、筛选、排序）
func (r *SharedPlanRepository) List(ctx context.Context, page, pageSize int, keyword string, tags []string, sortBy string) ([]usermodel.SharedPlanListItem, int, error) {
	// 构建查询条件
	where := []string{"sp.is_published = true"}
	args := []any{}
	argIdx := 1

	if keyword != "" {
		where = append(where, fmt.Sprintf("(sp.title ILIKE $%d OR sp.description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+keyword+"%")
		argIdx++
	}
	if len(tags) > 0 {
		where = append(where, fmt.Sprintf("sp.tags && $%d", argIdx))
		args = append(args, tags)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// 排序
	orderBy := "sp.created_at DESC"
	switch sortBy {
	case "popular":
		orderBy = "sp.subscribe_count DESC"
	case "newest":
		orderBy = "sp.created_at DESC"
	}

	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM shared_plans sp WHERE %s", whereClause)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := fmt.Sprintf(`
		SELECT sp.id, sp.title, sp.description, sp.tags, sp.daily_count, sp.total_days,
		       sp.subscribe_count, u.nickname as creator_name, sp.created_at
		FROM shared_plans sp
		JOIN users u ON sp.creator_id = u.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	limit := pageSize
	offset := (page - 1) * pageSize
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]usermodel.SharedPlanListItem, 0)
	for rows.Next() {
		var item usermodel.SharedPlanListItem
		err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Tags,
			&item.DailyCount, &item.TotalDays, &item.SubscribeCount, &item.CreatorName, &item.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	return list, total, rows.Err()
}

// GetMyPlans 获取我创建的计划
func (r *SharedPlanRepository) GetMyPlans(ctx context.Context, creatorID string) ([]usermodel.SharedPlanListItem, error) {
	query := `
		SELECT id, title, description, tags, daily_count, total_days, subscribe_count, is_published, created_at
		FROM shared_plans WHERE creator_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]usermodel.SharedPlanListItem, 0)
	for rows.Next() {
		var item usermodel.SharedPlanListItem
		var isPublished bool
		err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Tags,
			&item.DailyCount, &item.TotalDays, &item.SubscribeCount, &isPublished, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

// IncrementSubscribeCount 订阅数+1
func (r *SharedPlanRepository) IncrementSubscribeCount(ctx context.Context, planID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE shared_plans SET subscribe_count = subscribe_count + 1 WHERE id = $1`, planID)
	return err
}

// DecrementSubscribeCount 订阅数-1
func (r *SharedPlanRepository) DecrementSubscribeCount(ctx context.Context, planID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE shared_plans SET subscribe_count = GREATEST(subscribe_count - 1, 0) WHERE id = $1`, planID)
	return err
}

// ==================== 订阅管理 ====================

// CreateSubscription 创建订阅
func (r *SharedPlanRepository) CreateSubscription(ctx context.Context, sub *usermodel.PlanSubscription) error {
	query := `
		INSERT INTO plan_subscriptions (user_id, shared_plan_id, start_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		sub.UserID, sub.SharedPlanID, sub.StartDate, sub.Status, sub.CreatedAt, sub.UpdatedAt,
	).Scan(&sub.ID)
}

// GetSubscription 获取订阅关系
func (r *SharedPlanRepository) GetSubscription(ctx context.Context, userID string, sharedPlanID int64) (*usermodel.PlanSubscription, error) {
	query := `
		SELECT id, user_id, shared_plan_id, start_date, status, created_at, updated_at
		FROM plan_subscriptions WHERE user_id = $1 AND shared_plan_id = $2
	`
	row := r.db.QueryRow(ctx, query, userID, sharedPlanID)
	var sub usermodel.PlanSubscription
	err := row.Scan(&sub.ID, &sub.UserID, &sub.SharedPlanID, &sub.StartDate, &sub.Status, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetSubscriptionByID 根据订阅ID获取
func (r *SharedPlanRepository) GetSubscriptionByID(ctx context.Context, id int64) (*usermodel.PlanSubscription, error) {
	query := `
		SELECT id, user_id, shared_plan_id, start_date, status, is_current, queue_order, created_at, updated_at
		FROM plan_subscriptions WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	var sub usermodel.PlanSubscription
	err := row.Scan(&sub.ID, &sub.UserID, &sub.SharedPlanID, &sub.StartDate, &sub.Status, &sub.IsCurrent, &sub.QueueOrder, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// DeleteSubscription 取消订阅
func (r *SharedPlanRepository) DeleteSubscription(ctx context.Context, userID string, sharedPlanID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM plan_subscriptions WHERE user_id = $1 AND shared_plan_id = $2`, userID, sharedPlanID)
	return err
}

// UpdateStartDate 更新开始日期
func (r *SharedPlanRepository) UpdateStartDate(ctx context.Context, id int64, userID string, startDate string) error {
	_, err := r.db.Exec(ctx, `UPDATE plan_subscriptions SET start_date = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`, startDate, id, userID)
	return err
}

// UpdateSubscriptionStatus 更新订阅状态
func (r *SharedPlanRepository) UpdateSubscriptionStatus(ctx context.Context, id int64, userID string, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE plan_subscriptions SET status = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`, status, id, userID)
	return err
}

// GetMySubscriptions 获取我的订阅列表
func (r *SharedPlanRepository) GetMySubscriptions(ctx context.Context, userID string) ([]usermodel.SubscribeListResponse, error) {
	query := `
		SELECT ps.id, ps.shared_plan_id, sp.title, sp.tags, sp.daily_count, sp.total_days,
		       ps.start_date, ps.status, ps.created_at
		FROM plan_subscriptions ps
		JOIN shared_plans sp ON ps.shared_plan_id = sp.id
		WHERE ps.user_id = $1
		ORDER BY ps.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]usermodel.SubscribeListResponse, 0)
	for rows.Next() {
		var item usermodel.SubscribeListResponse
		err := rows.Scan(&item.ID, &item.SharedPlanID, &item.Title, &item.Tags,
			&item.DailyCount, &item.TotalDays, &item.StartDate, &item.Status, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

// GetLastCheckinByPoem 从 reading_progress 获取用户对某诗文的最新打卡记录（自建计划）
func (r *SharedPlanRepository) GetLastCheckinByPoem(ctx context.Context, userID string, poemID int64) (date string, planTitle string, daysAgo int, err error) {
	query := `
		SELECT rp.date, sp.title
		FROM reading_progress rp
		CROSS JOIN LATERAL unnest(rp.poem_ids) AS pid
		JOIN reading_plans rpl ON rpl.user_id = rp.user_id
		LEFT JOIN shared_plans sp ON rpl.shared_plan_id = sp.id
		WHERE rp.user_id = $1 AND pid = $2
		ORDER BY rp.date DESC
		LIMIT 1
	`
	var checkinDate string
	var title *string
	err = r.db.QueryRow(ctx, query, userID, poemID).Scan(&checkinDate, &title)
	if err != nil {
		return "", "", 0, err
	}
	if title != nil {
		planTitle = *title
	}
	parsedDate, _ := time.Parse("2006-01-02", checkinDate)
	daysAgo = int(time.Since(parsedDate).Hours() / 24)
	return checkinDate, planTitle, daysAgo, nil
}

// ==================== 排队机制新增方法 ====================

// GetActiveSubscription 获取用户当前 is_current=true 的订阅
func (r *SharedPlanRepository) GetActiveSubscription(ctx context.Context, userID string) (*usermodel.PlanSubscription, error) {
	query := `
		SELECT id, user_id, shared_plan_id, start_date, status, is_current, queue_order, created_at, updated_at
		FROM plan_subscriptions WHERE user_id = $1 AND is_current = true
		LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, userID)
	var sub usermodel.PlanSubscription
	err := row.Scan(&sub.ID, &sub.UserID, &sub.SharedPlanID, &sub.StartDate, &sub.Status, &sub.IsCurrent, &sub.QueueOrder, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetSubscriptionsByUserID 获取用户订阅（按 queue_order 排序）
// status 过滤规则：
//   - "" 或 "active" → 只返回 subscribed + completed（默认）
//   - "all"          → 返回全部（含 cancelled）
//   - "cancelled"    → 只返回 cancelled
//   - "subscribed"   → 只返回 subscribed
//   - "completed"    → 只返回 completed
func (r *SharedPlanRepository) GetSubscriptionsByUserID(ctx context.Context, userID string, status string) ([]usermodel.PlanSubscription, error) {
	query := `
		SELECT id, user_id, shared_plan_id, start_date, status, is_current, queue_order, created_at, updated_at
		FROM plan_subscriptions WHERE user_id = $1
	`
	args := []any{userID}
	argIdx := 2

	switch status {
	case "", "active":
		query += fmt.Sprintf(" AND status IN ($%d, $%d)", argIdx, argIdx+1)
		args = append(args, "subscribed", "completed")
		argIdx += 2
	case "all":
		// 不加过滤
	case "cancelled", "subscribed", "completed":
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	default:
		// 未知值按默认处理
		query += fmt.Sprintf(" AND status IN ($%d, $%d)", argIdx, argIdx+1)
		args = append(args, "subscribed", "completed")
		argIdx += 2
	}

	query += " ORDER BY queue_order ASC, created_at ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]usermodel.PlanSubscription, 0)
	for rows.Next() {
		var sub usermodel.PlanSubscription
		err := rows.Scan(&sub.ID, &sub.UserID, &sub.SharedPlanID, &sub.StartDate, &sub.Status, &sub.IsCurrent, &sub.QueueOrder, &sub.CreatedAt, &sub.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, sub)
	}
	return list, rows.Err()
}

// GetSubscriptionsByUserIDWithStatus 获取用户指定状态的订阅（按 queue_order 排序）
func (r *SharedPlanRepository) GetSubscriptionsByUserIDWithStatus(ctx context.Context, userID string, status string) ([]usermodel.PlanSubscription, error) {
	query := `
		SELECT id, user_id, shared_plan_id, start_date, status, is_current, queue_order, created_at, updated_at
		FROM plan_subscriptions WHERE user_id = $1 AND status = $2
		ORDER BY queue_order ASC, created_at ASC
	`
	rows, err := r.db.Query(ctx, query, userID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]usermodel.PlanSubscription, 0)
	for rows.Next() {
		var sub usermodel.PlanSubscription
		err := rows.Scan(&sub.ID, &sub.UserID, &sub.SharedPlanID, &sub.StartDate, &sub.Status, &sub.IsCurrent, &sub.QueueOrder, &sub.CreatedAt, &sub.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, sub)
	}
	return list, rows.Err()
}

// SoftDeleteSubscription 软删除（status → cancelled）
func (r *SharedPlanRepository) SoftDeleteSubscription(ctx context.Context, userID string, sharedPlanID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE plan_subscriptions SET status = 'cancelled', is_current = false, updated_at = NOW() WHERE user_id = $1 AND shared_plan_id = $2`, userID, sharedPlanID)
	return err
}

// ActivateSubscription 激活订阅
func (r *SharedPlanRepository) ActivateSubscription(ctx context.Context, subID int64, userID string, startDate time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE plan_subscriptions SET is_current = true, status = 'subscribed', start_date = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`, startDate, subID, userID)
	return err
}

// DeactivateCurrentSubscription 取消当前计划标记
func (r *SharedPlanRepository) DeactivateCurrentSubscription(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE plan_subscriptions SET is_current = false, updated_at = NOW() WHERE user_id = $1 AND is_current = true`, userID)
	return err
}

// UpdateQueueOrder 更新 queue_order
func (r *SharedPlanRepository) UpdateQueueOrder(ctx context.Context, subID int64, queueOrder int) error {
	_, err := r.db.Exec(ctx, `UPDATE plan_subscriptions SET queue_order = $1, updated_at = NOW() WHERE id = $2`, queueOrder, subID)
	return err
}

// ShiftQueueOrders 批量前移 queue_order（大于 afterOrder 的全部 -1）
func (r *SharedPlanRepository) ShiftQueueOrders(ctx context.Context, userID string, afterOrder int) error {
	_, err := r.db.Exec(ctx, `UPDATE plan_subscriptions SET queue_order = queue_order - 1, updated_at = NOW() WHERE user_id = $1 AND queue_order > $2`, userID, afterOrder)
	return err
}

// SwapQueueOrders 交换两个订阅的 queue_order（用于上下移动）
func (r *SharedPlanRepository) SwapQueueOrders(ctx context.Context, userID string, subID1, subID2 int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE plan_subscriptions SET queue_order = sub.new_order, updated_at = NOW()
		FROM (
			SELECT id, CASE
				WHEN id = $1 THEN (SELECT queue_order FROM plan_subscriptions WHERE id = $2)
				ELSE (SELECT queue_order FROM plan_subscriptions WHERE id = $1)
			END AS new_order
			FROM plan_subscriptions
			WHERE user_id = $3 AND id IN ($1, $2) AND status = 'subscribed'
		) sub
		WHERE plan_subscriptions.id = sub.id
	`, subID1, subID2, userID)
	return err
}

// GetSubscriptionByQueueOrder 根据 queue_order 查找订阅
func (r *SharedPlanRepository) GetSubscriptionByQueueOrder(ctx context.Context, userID string, queueOrder int) (*usermodel.PlanSubscription, error) {
	var sub usermodel.PlanSubscription
	query := `SELECT id, user_id, shared_plan_id, start_date, status, is_current, queue_order, created_at, updated_at
	          FROM plan_subscriptions WHERE user_id = $1 AND queue_order = $2 AND status = 'subscribed'`
	err := r.db.QueryRow(ctx, query, userID, queueOrder).Scan(
		&sub.ID, &sub.UserID, &sub.SharedPlanID, &sub.StartDate, &sub.Status,
		&sub.IsCurrent, &sub.QueueOrder, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// RecalculateQueueOrders 重新计算用户所有 subscribed 订阅的 queue_order
func (r *SharedPlanRepository) RecalculateQueueOrders(ctx context.Context, userID string) error {
	query := `
		UPDATE plan_subscriptions SET queue_order = sub.new_order, updated_at = NOW()
		FROM (
			SELECT id, ROW_NUMBER() OVER (ORDER BY created_at ASC) - 1 AS new_order
			FROM plan_subscriptions
			WHERE user_id = $1 AND status = 'subscribed'
		) sub
		WHERE plan_subscriptions.id = sub.id
	`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
