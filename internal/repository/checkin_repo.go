package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/model"
)

type CheckinRepository struct {
	db *pgxpool.Pool
}

func NewCheckinRepository(db *pgxpool.Pool) *CheckinRepository {
	return &CheckinRepository{db: db}
}

// Create 创建打卡记录
func (r *CheckinRepository) Create(ctx context.Context, checkin *model.CheckIn) error {
	query := `
		INSERT INTO checkins (user_id, date, consecutive_day, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, date) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, checkin.UserID, checkin.Date, checkin.ConsecutiveDay, checkin.CreatedAt)
	return err
}

// GetByDate 获取指定日期的打卡记录
func (r *CheckinRepository) GetByDate(ctx context.Context, userID string, date time.Time) (*model.CheckIn, error) {
	query := `SELECT user_id, date, consecutive_day, created_at FROM checkins WHERE user_id = $1 AND date = $2`
	row := r.db.QueryRow(ctx, query, userID, date)
	var c model.CheckIn
	err := row.Scan(&c.UserID, &c.Date, &c.ConsecutiveDay, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetLastCheckIn 获取最近一次打卡记录
func (r *CheckinRepository) GetLastCheckIn(ctx context.Context, userID string) (*model.CheckIn, error) {
	query := `
		SELECT user_id, date, consecutive_day, created_at
		FROM checkins WHERE user_id = $1
		ORDER BY date DESC LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, userID)
	var c model.CheckIn
	err := row.Scan(&c.UserID, &c.Date, &c.ConsecutiveDay, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// List 获取打卡记录列表
func (r *CheckinRepository) List(ctx context.Context, userID string, page, pageSize int) ([]model.CheckIn, int64, error) {
	// 获取总数
	countQuery := `SELECT COUNT(*) FROM checkins WHERE user_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT user_id, date, consecutive_day, created_at
		FROM checkins WHERE user_id = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var checkins []model.CheckIn
	for rows.Next() {
		var c model.CheckIn
		err := rows.Scan(&c.UserID, &c.Date, &c.ConsecutiveDay, &c.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		checkins = append(checkins, c)
	}
	return checkins, total, rows.Err()
}

// GetStats 获取打卡统计
func (r *CheckinRepository) GetStats(ctx context.Context, userID string) (*model.CheckInStats, error) {
	query := `
		SELECT user_id, total_days, consecutive_day, max_consecutive, last_check_in
		FROM checkin_stats WHERE user_id = $1
	`
	row := r.db.QueryRow(ctx, query, userID)
	var s model.CheckInStats
	err := row.Scan(&s.UserID, &s.TotalDays, &s.ConsecutiveDay, &s.MaxConsecutive, &s.LastCheckIn)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpsertStats 更新或插入打卡统计
func (r *CheckinRepository) UpsertStats(ctx context.Context, stats *model.CheckInStats) error {
	query := `
		INSERT INTO checkin_stats (user_id, total_days, consecutive_day, max_consecutive, last_check_in)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			total_days = EXCLUDED.total_days,
			consecutive_day = EXCLUDED.consecutive_day,
			max_consecutive = EXCLUDED.max_consecutive,
			last_check_in = EXCLUDED.last_check_in
	`
	_, err := r.db.Exec(ctx, query,
		stats.UserID, stats.TotalDays, stats.ConsecutiveDay, stats.MaxConsecutive, stats.LastCheckIn)
	return err
}

// GetCheckInDates 获取指定月份的打卡日期
func (r *CheckinRepository) GetCheckInDates(ctx context.Context, userID string, year, month int) ([]int, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	query := `
		SELECT EXTRACT(DAY FROM date)::int
		FROM checkins
		WHERE user_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date
	`
	rows, err := r.db.Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []int
	for rows.Next() {
		var day int
		err := rows.Scan(&day)
		if err != nil {
			return nil, err
		}
		days = append(days, day)
	}
	return days, rows.Err()
}

// GetRanking 获取排行榜
func (r *CheckinRepository) GetRanking(ctx context.Context, limit int) ([]model.RankingItem, error) {
	query := `
		SELECT u.nickname, cs.consecutive_day
		FROM checkin_stats cs
		JOIN users u ON cs.user_id = u.id
		ORDER BY cs.consecutive_day DESC, cs.last_check_in DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.RankingItem
	rank := 1
	for rows.Next() {
		var item model.RankingItem
		err := rows.Scan(&item.Nickname, &item.ConsecutiveDay)
		if err != nil {
			return nil, err
		}
		item.Rank = rank
		items = append(items, item)
		rank++
	}
	return items, rows.Err()
}
