package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	adminmodel "poem-backend/internal/model/admin"
)

type ImportRecordRepository struct {
	db *pgxpool.Pool
}

func NewImportRecordRepository(db *pgxpool.Pool) *ImportRecordRepository {
	return &ImportRecordRepository{db: db}
}

// Create 创建导入记录
func (r *ImportRecordRepository) Create(ctx context.Context, record *adminmodel.ImportRecord) (int64, error) {
	query := `
		INSERT INTO import_records (file_name, source, total, success, failed, status, errors, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query,
		record.FileName, record.Source, record.Total, record.Success, record.Failed,
		record.Status, record.Errors, record.CreatedBy, record.CreatedAt,
	).Scan(&id)
	return id, err
}

// UpdateStatus 更新导入记录状态（用于导入过程中增量更新）
func (r *ImportRecordRepository) UpdateStatus(ctx context.Context, id int64, success, failed int, status string, errors []adminmodel.ImportError) error {
	query := `
		UPDATE import_records
		SET success = $1, failed = $2, status = $3, errors = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, success, failed, status, errors, id)
	return err
}

// List 分页查询导入记录（支持筛选）
func (r *ImportRecordRepository) List(ctx context.Context, page, pageSize int, status, startDate, endDate string) ([]adminmodel.ImportRecord, int, error) {
	// 构建查询条件
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if startDate != "" {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		where += fmt.Sprintf(" AND created_at < $%d::date + interval '1 day'", argIdx)
		args = append(args, endDate)
		argIdx++
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM import_records " + where
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询列表
	listQuery := fmt.Sprintf(`
		SELECT id, file_name, source, total, success, failed, status, errors, created_by, created_at
		FROM import_records %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	// 复制 args 并添加分页参数
	listArgs := make([]any, len(args))
	copy(listArgs, args)
	listArgs = append(listArgs, pageSize, (page-1)*pageSize)

	rows, err := r.db.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []adminmodel.ImportRecord
	for rows.Next() {
		var rec adminmodel.ImportRecord
		if err := rows.Scan(
			&rec.ID, &rec.FileName, &rec.Source, &rec.Total, &rec.Success, &rec.Failed,
			&rec.Status, &rec.Errors, &rec.CreatedBy, &rec.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetByID 获取单条导入记录
func (r *ImportRecordRepository) GetByID(ctx context.Context, id int64) (*adminmodel.ImportRecord, error) {
	query := `
		SELECT id, file_name, source, total, success, failed, status, errors, created_by, created_at
		FROM import_records WHERE id = $1
	`
	var rec adminmodel.ImportRecord
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&rec.ID, &rec.FileName, &rec.Source, &rec.Total, &rec.Success, &rec.Failed,
		&rec.Status, &rec.Errors, &rec.CreatedBy, &rec.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &rec, nil
}

// GetStats 汇总统计
func (r *ImportRecordRepository) GetStats(ctx context.Context, status, startDate, endDate string) (*adminmodel.ImportRecordStatsResponse, error) {
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if startDate != "" {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		where += fmt.Sprintf(" AND created_at < $%d::date + interval '1 day'", argIdx)
		args = append(args, endDate)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_imports,
			COALESCE(SUM(total), 0) as total_poems,
			COALESCE(SUM(success), 0) as total_success,
			COALESCE(SUM(failed), 0) as total_failed
		FROM import_records %s
	`, where)

	var stats adminmodel.ImportRecordStatsResponse
	if err := r.db.QueryRow(ctx, query, args...).Scan(
		&stats.TotalImports, &stats.TotalPoems, &stats.TotalSuccess, &stats.TotalFailed,
	); err != nil {
		return nil, err
	}

	// 计算成功率（四舍五入保留 1 位小数）
	if stats.TotalPoems > 0 {
		rate := float64(stats.TotalSuccess) / float64(stats.TotalPoems) * 100
		stats.SuccessRate = float64(int(rate*10+0.5)) / 10
	}

	return &stats, nil
}
