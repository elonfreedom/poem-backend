package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsRepository struct {
	db *pgxpool.Pool
}

func NewStatsRepository(db *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{db: db}
}

// GetOverview 获取总览统计
func (r *StatsRepository) GetOverview(ctx context.Context) (totalPoems, totalUsers, totalViews, todayViews, todayUsers, totalCategories, totalTags int64, err error) {
	// 诗歌总数
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM poems").Scan(&totalPoems)
	if err != nil {
		return
	}
	// 用户总数
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return
	}
	// 总浏览量
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM poem_views").Scan(&totalViews)
	if err != nil {
		return
	}
	// 今日浏览量
	today := time.Now().Format("2006-01-02")
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM poem_views WHERE created_at::date = $1::date", today).Scan(&todayViews)
	if err != nil {
		return
	}
	// 今日新增用户
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE created_at::date = $1::date", today).Scan(&todayUsers)
	if err != nil {
		return
	}
	// 分类总数
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM categories").Scan(&totalCategories)
	if err != nil {
		return
	}
	// 标签总数
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM tags").Scan(&totalTags)
	return
}

// GetDailyStats 获取每日统计
func (r *StatsRepository) GetDailyStats(ctx context.Context, days int) ([]struct {
	Date  string
	Views int64
	Users int64
}, error) {
	query := `
		SELECT d.date,
		       COALESCE(v.views, 0) AS views,
		       COALESCE(u.users, 0) AS users
		FROM generate_series(
		    CURRENT_DATE - ($1 || ' days')::interval,
		    CURRENT_DATE,
		    '1 day'::interval
		) AS d(date)
		LEFT JOIN (
		    SELECT created_at::date AS date, COUNT(*) AS views
		    FROM poem_views
		    WHERE created_at::date >= CURRENT_DATE - ($1 || ' days')::interval
		    GROUP BY created_at::date
		) v ON d.date = v.date
		LEFT JOIN (
		    SELECT created_at::date AS date, COUNT(*) AS users
		    FROM users
		    WHERE created_at::date >= CURRENT_DATE - ($1 || ' days')::interval
		    GROUP BY created_at::date
		) u ON d.date = u.date
		ORDER BY d.date
	`
	rows, err := r.db.Query(ctx, query, "30")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Date  string
		Views int64
		Users int64
	}
	for rows.Next() {
		var r struct {
			Date  string
			Views int64
			Users int64
		}
		if err := rows.Scan(&r.Date, &r.Views, &r.Users); err != nil {
			return nil, err
		}
		// 格式化日期
		if t, err := time.Parse("2006-01-02", r.Date[:10]); err == nil {
			r.Date = t.Format("2006-01-02")
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetHotPoems 获取热门诗歌
func (r *StatsRepository) GetHotPoems(ctx context.Context, limit int) ([]struct {
	PoemID    int64
	Title     string
	Author    string
	ViewCount int64
}, error) {
	query := `
		SELECT p.id AS poem_id, p.title, p.author, COUNT(pv.id) AS view_count
		FROM poems p
		LEFT JOIN poem_views pv ON p.id = pv.poem_id
		GROUP BY p.id, p.title, p.author
		ORDER BY view_count DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		PoemID    int64
		Title     string
		Author    string
		ViewCount int64
	}
	for rows.Next() {
		var r struct {
			PoemID    int64
			Title     string
			Author    string
			ViewCount int64
		}
		if err := rows.Scan(&r.PoemID, &r.Title, &r.Author, &r.ViewCount); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetUserGrowth 获取用户增长
func (r *StatsRepository) GetUserGrowth(ctx context.Context, days int) ([]struct {
	Date       string
	NewUsers   int64
	TotalUsers int64
}, error) {
	query := `
		SELECT d.date,
		       COALESCE(n.new_users, 0) AS new_users,
		       (SELECT COUNT(*) FROM users WHERE created_at::date <= d.date) AS total_users
		FROM generate_series(
		    CURRENT_DATE - ($1 || ' days')::interval,
		    CURRENT_DATE,
		    '1 day'::interval
		) AS d(date)
		LEFT JOIN (
		    SELECT created_at::date AS date, COUNT(*) AS new_users
		    FROM users
		    WHERE created_at::date >= CURRENT_DATE - ($1 || ' days')::interval
		    GROUP BY created_at::date
		) n ON d.date = n.date
		ORDER BY d.date
	`
	rows, err := r.db.Query(ctx, query, "30")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Date       string
		NewUsers   int64
		TotalUsers int64
	}
	for rows.Next() {
		var r struct {
			Date       string
			NewUsers   int64
			TotalUsers int64
		}
		if err := rows.Scan(&r.Date, &r.NewUsers, &r.TotalUsers); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02", r.Date[:10]); err == nil {
			r.Date = t.Format("2006-01-02")
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
