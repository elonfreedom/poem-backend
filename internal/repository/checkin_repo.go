package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	usermodel "poem-backend/internal/model/user"
)

type CheckinRepository struct {
	db *pgxpool.Pool
}

func NewCheckinRepository(db *pgxpool.Pool) *CheckinRepository {
	return &CheckinRepository{db: db}
}

// Create 创建打卡记录（旧版，无订阅）
// 使用 ON CONFLICT (column_list) WHERE predicate 匹配部分唯一索引 uk_checkins_user_date_no_sub
func (r *CheckinRepository) Create(ctx context.Context, checkin *usermodel.CheckIn) error {
	query := `
		INSERT INTO checkins (user_id, date, consecutive_day, poem_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, date) WHERE subscription_id IS NULL DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, checkin.UserID, checkin.Date, checkin.ConsecutiveDay, checkin.PoemID, checkin.CreatedAt)
	return err
}

// ==================== 订阅打卡（合并自 plan_checkins） ====================

// CreateSubscriptionCheckin 创建订阅打卡记录
// 使用 ON CONFLICT (column_list) WHERE predicate 匹配部分唯一索引 uk_checkins_sub_day
func (r *CheckinRepository) CreateSubscriptionCheckin(ctx context.Context, userID string, subscriptionID int64, dayNumber int, date time.Time, poemIDs []int64, createdAt time.Time) error {
	query := `
		INSERT INTO checkins (user_id, date, consecutive_day, poem_id, poem_ids, subscription_id, day_number, created_at)
		VALUES ($1, $2, 1, $3, $4, $5, $6, $7)
		ON CONFLICT (subscription_id, day_number) WHERE subscription_id IS NOT NULL DO UPDATE SET
			poem_ids = EXCLUDED.poem_ids,
			poem_id = EXCLUDED.poem_id,
			date = EXCLUDED.date
	`
	var poemID *int64
	if len(poemIDs) > 0 {
		poemID = &poemIDs[0]
	}
	_, err := r.db.Exec(ctx, query, userID, date, poemID, poemIDs, subscriptionID, dayNumber, createdAt)
	return err
}

// GetCheckinBySubDay 按订阅+天数查询打卡
func (r *CheckinRepository) GetCheckinBySubDay(ctx context.Context, subscriptionID int64, dayNumber int) (*usermodel.CheckIn, error) {
	query := `
		SELECT id, user_id, date, consecutive_day, poem_id, poem_ids, subscription_id, day_number, created_at
		FROM checkins WHERE subscription_id = $1 AND day_number = $2
	`
	row := r.db.QueryRow(ctx, query, subscriptionID, dayNumber)
	var c usermodel.CheckIn
	err := row.Scan(&c.ID, &c.UserID, &c.Date, &c.ConsecutiveDay, &c.PoemID, &c.PoemIDs, &c.SubscriptionID, &c.DayNumber, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCheckinsBySubscription 获取订阅的所有打卡记录
func (r *CheckinRepository) GetCheckinsBySubscription(ctx context.Context, subscriptionID int64) ([]usermodel.CheckIn, error) {
	query := `
		SELECT id, user_id, date, consecutive_day, poem_id, poem_ids, subscription_id, day_number, created_at
		FROM checkins WHERE subscription_id = $1
		ORDER BY day_number
	`
	rows, err := r.db.Query(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []usermodel.CheckIn
	for rows.Next() {
		var c usermodel.CheckIn
		err := rows.Scan(&c.ID, &c.UserID, &c.Date, &c.ConsecutiveDay, &c.PoemID, &c.PoemIDs, &c.SubscriptionID, &c.DayNumber, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// CountCheckinsBySubscription 统计订阅打卡天数
func (r *CheckinRepository) CountCheckinsBySubscription(ctx context.Context, subscriptionID int64) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM checkins WHERE subscription_id = $1`, subscriptionID).Scan(&count)
	return count, err
}

// GetLastCheckinByPoem 从 checkins 表查询用户对某诗文的最新打卡（含订阅打卡）
func (r *CheckinRepository) GetLastCheckinByPoem(ctx context.Context, userID string, poemID int64) (date string, planTitle string, daysAgo int, err error) {
	query := `
		SELECT c.date::text, sp.title
		FROM checkins c
		JOIN plan_subscriptions ps ON c.subscription_id = ps.id
		JOIN shared_plans sp ON ps.shared_plan_id = sp.id
		CROSS JOIN LATERAL unnest(c.poem_ids) AS pid
		WHERE c.user_id = $1 AND pid = $2
		ORDER BY c.date DESC
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

// GetByDate 获取指定日期的打卡记录
func (r *CheckinRepository) GetByDate(ctx context.Context, userID string, date time.Time) (*usermodel.CheckIn, error) {
	query := `SELECT user_id, date, consecutive_day, created_at FROM checkins WHERE user_id = $1 AND date = $2`
	row := r.db.QueryRow(ctx, query, userID, date)
	var c usermodel.CheckIn
	err := row.Scan(&c.UserID, &c.Date, &c.ConsecutiveDay, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetLastCheckIn 获取最近一次打卡记录
func (r *CheckinRepository) GetLastCheckIn(ctx context.Context, userID string) (*usermodel.CheckIn, error) {
	query := `
		SELECT user_id, date, consecutive_day, created_at
		FROM checkins WHERE user_id = $1
		ORDER BY date DESC LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, userID)
	var c usermodel.CheckIn
	err := row.Scan(&c.UserID, &c.Date, &c.ConsecutiveDay, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// List 获取打卡记录列表（JOIN poems 表获取诗文标题，用于热力图）
func (r *CheckinRepository) List(ctx context.Context, userID string, page, pageSize int, start, end time.Time) ([]usermodel.CheckIn, int64, error) {
	where := "WHERE c.user_id = $1"
	args := []interface{}{userID}
	argIdx := 2

	if !start.IsZero() {
		where += " AND c.date >= $" + string(rune('0'+argIdx))
		args = append(args, start)
		argIdx++
	}
	if !end.IsZero() {
		where += " AND c.date <= $" + string(rune('0'+argIdx))
		args = append(args, end)
		argIdx++
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM checkins c " + where
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表（LEFT JOIN poems 获取诗文标题）
	query := `
		SELECT c.user_id, c.date, c.consecutive_day, c.poem_id, p.title, c.created_at
		FROM checkins c
		LEFT JOIN poems p ON c.poem_id = p.id
		` + where + `
		ORDER BY c.date DESC
		LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var checkins []usermodel.CheckIn
	for rows.Next() {
		var c usermodel.CheckIn
		err := rows.Scan(&c.UserID, &c.Date, &c.ConsecutiveDay, &c.PoemID, &c.PoemTitle, &c.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		checkins = append(checkins, c)
	}
	return checkins, total, rows.Err()
}

// GetStats 获取打卡统计
func (r *CheckinRepository) GetStats(ctx context.Context, userID string) (*usermodel.CheckInStats, error) {
	query := `
		SELECT user_id, total_days, consecutive_day, max_consecutive, last_check_in
		FROM checkin_stats WHERE user_id = $1
	`
	row := r.db.QueryRow(ctx, query, userID)
	var s usermodel.CheckInStats
	err := row.Scan(&s.UserID, &s.TotalDays, &s.ConsecutiveDay, &s.MaxConsecutive, &s.LastCheckIn)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpsertStats 更新或插入打卡统计
func (r *CheckinRepository) UpsertStats(ctx context.Context, stats *usermodel.CheckInStats) error {
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

// ListAll 获取所有用户的打卡记录列表（管理后台用，JOIN users 表获取昵称，支持昵称搜索和日期范围筛选）
func (r *CheckinRepository) ListAll(ctx context.Context, page, pageSize int, keyword, startDate, endDate string) ([]usermodel.CheckIn, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if keyword != "" {
		where += " AND u.nickname ILIKE $" + string(rune('0'+argIdx))
		args = append(args, "%"+keyword+"%")
		argIdx++
	}
	if startDate != "" {
		where += " AND c.date >= $" + string(rune('0'+argIdx))
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		where += " AND c.date <= $" + string(rune('0'+argIdx))
		args = append(args, endDate)
		argIdx++
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM checkins c JOIN users u ON c.user_id = u.id " + where
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT c.user_id, u.nickname, c.date, c.consecutive_day, c.poem_id, p.title, c.created_at
		FROM checkins c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN poems p ON c.poem_id = p.id
		` + where + `
		ORDER BY c.date DESC, c.created_at DESC
		LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var checkins []usermodel.CheckIn
	for rows.Next() {
		var c usermodel.CheckIn
		err := rows.Scan(&c.UserID, &c.Nickname, &c.Date, &c.ConsecutiveDay, &c.PoemID, &c.PoemTitle, &c.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		checkins = append(checkins, c)
	}
	return checkins, total, rows.Err()
}

// CheckinHotPoem 打卡热门诗歌
type CheckinHotPoem struct {
	PoemID       int64  `db:"poem_id"`
	PoemTitle    string `db:"title"`
	CheckinCount int64  `db:"checkin_count"`
}

// CheckinStats 打卡统计（管理后台用）
type CheckinStats struct {
	TotalCheckins int64
	TotalUsers    int64
	DailyAvgRate  float64
	Retention7d   float64
	HotPoems      []CheckinHotPoem
}

// GetCheckinStats 获取打卡数据统计（管理后台用，支持日期范围）
func (r *CheckinRepository) GetCheckinStats(ctx context.Context, startDate, endDate string) (*CheckinStats, error) {
	stats := &CheckinStats{}

	// 日期范围条件
	dateWhere := ""
	args := []interface{}{}
	argIdx := 1
	if startDate != "" {
		dateWhere += " AND date >= $" + string(rune('0'+argIdx))
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		dateWhere += " AND date <= $" + string(rune('0'+argIdx))
		args = append(args, endDate)
		argIdx++
	}

	// 总打卡数
	totalQuery := "SELECT COUNT(*) FROM checkins WHERE 1=1" + dateWhere
	if err := r.db.QueryRow(ctx, totalQuery, args...).Scan(&stats.TotalCheckins); err != nil {
		return nil, err
	}

	// 总打卡用户数（去重）
	usersQuery := "SELECT COUNT(DISTINCT user_id) FROM checkins WHERE 1=1" + dateWhere
	if err := r.db.QueryRow(ctx, usersQuery, args...).Scan(&stats.TotalUsers); err != nil {
		return nil, err
	}

	// 日均打卡率 = 总打卡数 / 总用户数（简化计算）
	if stats.TotalUsers > 0 {
		stats.DailyAvgRate = float64(int(float64(stats.TotalCheckins)/float64(stats.TotalUsers)*100+0.5)) / 100
	}

	// 7 日留存率：7 天前有打卡的用户中，近 7 天仍有打卡的比例
	retentionQuery := `
		SELECT
			COUNT(DISTINCT old.user_id) AS old_users,
			COUNT(DISTINCT recent.user_id) AS retained_users
		FROM checkins old
		LEFT JOIN checkins recent ON old.user_id = recent.user_id
			AND recent.date >= CURRENT_DATE - INTERVAL '7 days'
			AND recent.date <= CURRENT_DATE
		WHERE old.date >= CURRENT_DATE - INTERVAL '14 days'
			AND old.date < CURRENT_DATE - INTERVAL '7 days'
	`
	var oldUsers, retainedUsers int64
	if err := r.db.QueryRow(ctx, retentionQuery).Scan(&oldUsers, &retainedUsers); err != nil {
		return nil, err
	}
	if oldUsers > 0 {
		stats.Retention7d = float64(int(float64(retainedUsers)/float64(oldUsers)*100+0.5)) / 100
	}

	// 热门诗歌 TOP 10（按打卡次数）
	hotQuery := `
		SELECT c.poem_id, p.title, COUNT(*) AS checkin_count
		FROM checkins c
		JOIN poems p ON c.poem_id = p.id
		WHERE c.poem_id IS NOT NULL` + dateWhere + `
		GROUP BY c.poem_id, p.title
		ORDER BY checkin_count DESC
		LIMIT 10
	`
	hotRows, err := r.db.Query(ctx, hotQuery, args...)
	if err != nil {
		return nil, err
	}
	defer hotRows.Close()

	for hotRows.Next() {
		var item CheckinHotPoem
		if err := hotRows.Scan(&item.PoemID, &item.PoemTitle, &item.CheckinCount); err != nil {
			return nil, err
		}
		stats.HotPoems = append(stats.HotPoems, item)
	}
	if err := hotRows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRanking 获取排行榜
func (r *CheckinRepository) GetRanking(ctx context.Context, limit int) ([]usermodel.RankingItem, error) {
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

	var items []usermodel.RankingItem
	rank := 1
	for rows.Next() {
		var item usermodel.RankingItem
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
