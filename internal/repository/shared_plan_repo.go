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
		SELECT id, user_id, shared_plan_id, start_date, status, created_at, updated_at
		FROM plan_subscriptions WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	var sub usermodel.PlanSubscription
	err := row.Scan(&sub.ID, &sub.UserID, &sub.SharedPlanID, &sub.StartDate, &sub.Status, &sub.CreatedAt, &sub.UpdatedAt)
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
