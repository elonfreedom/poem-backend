package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	usermodel "poem-backend/internal/model/user"
)

type ReadingPlanRepository struct {
	db *pgxpool.Pool
}

func NewReadingPlanRepository(db *pgxpool.Pool) *ReadingPlanRepository {
	return &ReadingPlanRepository{db: db}
}

// Create 创建阅读计划
func (r *ReadingPlanRepository) Create(ctx context.Context, plan *usermodel.ReadingPlan) error {
	// 获取用户当前最大 plan_id
	var maxPlanID int
	err := r.db.QueryRow(ctx, `SELECT COALESCE(MAX(plan_id), 0) FROM reading_plans WHERE user_id = $1`, plan.UserID).Scan(&maxPlanID)
	if err != nil {
		return err
	}
	plan.PlanID = maxPlanID + 1

	query := `
		INSERT INTO reading_plans (user_id, plan_id, title, daily_count, start_date, end_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = r.db.Exec(ctx, query,
		plan.UserID, plan.PlanID, plan.Title, plan.DailyCount, plan.StartDate, plan.EndDate,
		plan.Status, plan.CreatedAt, plan.UpdatedAt)
	return err
}

// GetByID 根据 user_id 和 plan_id 获取计划
func (r *ReadingPlanRepository) GetByID(ctx context.Context, userID string, planID int) (*usermodel.ReadingPlan, error) {
	query := `
		SELECT user_id, plan_id, title, daily_count, start_date, end_date, status, created_at, updated_at
		FROM reading_plans WHERE user_id = $1 AND plan_id = $2
	`
	row := r.db.QueryRow(ctx, query, userID, planID)
	var plan usermodel.ReadingPlan
	err := row.Scan(&plan.UserID, &plan.PlanID, &plan.Title, &plan.DailyCount, &plan.StartDate,
		&plan.EndDate, &plan.Status, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// GetActiveByUserID 获取用户当前进行中的计划
func (r *ReadingPlanRepository) GetActiveByUserID(ctx context.Context, userID string) (*usermodel.ReadingPlan, error) {
	query := `
		SELECT user_id, plan_id, title, daily_count, start_date, end_date, status, created_at, updated_at
		FROM reading_plans WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, userID)
	var plan usermodel.ReadingPlan
	err := row.Scan(&plan.UserID, &plan.PlanID, &plan.Title, &plan.DailyCount, &plan.StartDate,
		&plan.EndDate, &plan.Status, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// UpdateStatus 更新计划状态
func (r *ReadingPlanRepository) UpdateStatus(ctx context.Context, userID string, planID int, status string) error {
	query := `UPDATE reading_plans SET status = $1, updated_at = NOW() WHERE user_id = $2 AND plan_id = $3`
	_, err := r.db.Exec(ctx, query, status, userID, planID)
	return err
}

// CountByUserID 获取用户阅读计划数量
func (r *ReadingPlanRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM reading_plans WHERE user_id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// GetProgress 获取阅读进度
func (r *ReadingPlanRepository) GetProgress(ctx context.Context, userID string, date time.Time) (*usermodel.ReadingProgress, error) {
	query := `
		SELECT user_id, date, read_count, poem_ids, created_at
		FROM reading_progress WHERE user_id = $1 AND date = $2
	`
	row := r.db.QueryRow(ctx, query, userID, date)
	var p usermodel.ReadingProgress
	err := row.Scan(&p.UserID, &p.Date, &p.ReadCount, &p.PoemIDs, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpsertProgress 更新或插入阅读进度
func (r *ReadingPlanRepository) UpsertProgress(ctx context.Context, progress *usermodel.ReadingProgress) error {
	query := `
		INSERT INTO reading_progress (user_id, date, read_count, poem_ids, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, date) DO UPDATE SET
			read_count = EXCLUDED.read_count,
			poem_ids = EXCLUDED.poem_ids
	`
	_, err := r.db.Exec(ctx, query,
		progress.UserID, progress.Date, progress.ReadCount, progress.PoemIDs, progress.CreatedAt)
	return err
}

// GetProgressByDateRange 获取日期范围内的进度
func (r *ReadingPlanRepository) GetProgressByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]usermodel.ReadingProgress, error) {
	query := `
		SELECT user_id, date, read_count, poem_ids, created_at
		FROM reading_progress
		WHERE user_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date
	`
	rows, err := r.db.Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progressList []usermodel.ReadingProgress
	for rows.Next() {
		var p usermodel.ReadingProgress
		err := rows.Scan(&p.UserID, &p.Date, &p.ReadCount, &p.PoemIDs, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		progressList = append(progressList, p)
	}
	return progressList, rows.Err()
}
