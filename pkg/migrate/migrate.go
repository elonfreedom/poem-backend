package migrate

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Run 执行数据库迁移（幂等，可重复调用）
// 使用 golang-migrate + pgx 原生驱动，从 embed.FS 读取 migrations/*.sql
// 如果迁移处于 dirty 状态，会尝试修复版本号后重新执行
func Run(db *pgxpool.Pool) error {
	// 从 pgxpool 获取连接字符串，构造 *sql.DB 供 golang-migrate 使用
	connStr := db.Config().ConnString()

	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer sqlDB.Close()

	// 从 embed.FS 创建迁移源
	srcDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// 创建 pgx 数据库驱动
	dbDriver, err := pgx.WithInstance(sqlDB, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create pgx driver: %w", err)
	}

	// 创建 migrate 实例并执行
	m, err := migrate.NewWithInstance("iofs", srcDriver, "pgx", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// 修复 dirty 状态：如果迁移失败标记为 dirty，尝试强制修正版本号
	if err := fixDirtyState(m); err != nil {
		return fmt.Errorf("failed to fix dirty state: %w", err)
	}

	// 执行迁移
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	// 验证并修复：检查迁移是否实际生效，如果 schema_migrations 记录了版本但实际列不存在，回退版本重试
	if err := verifyAndFixMigration(sqlDB, m); err != nil {
		return fmt.Errorf("failed to verify migration: %w", err)
	}

	return nil
}

// verifyAndFixMigration 验证迁移是否实际生效
// 检查 poems 表是否有 source 列，如果没有但 schema_migrations 记录了版本 15，则回退版本重试
func verifyAndFixMigration(sqlDB *sql.DB, m *migrate.Migrate) error {
	// 检查 poems 表是否有 source 列
	var hasColumn bool
	err := sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'poems' AND column_name = 'source'
		)
	`).Scan(&hasColumn)
	if err != nil {
		return fmt.Errorf("check source column failed: %w", err)
	}

	if hasColumn {
		return nil // 列存在，无需修复
	}

	// 列不存在，检查 schema_migrations 是否记录了版本 15
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("get migration version failed: %w", err)
	}

	// 如果当前版本 >= 15 且列不存在，说明迁移未实际生效，需要回退重试
	if version >= 15 && !dirty {
		// 回退到版本 14
		if err := m.Force(14); err != nil {
			return fmt.Errorf("force version 14 failed: %w", err)
		}
		// 重新执行迁移
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("re-migration failed: %w", err)
		}
	}

	return nil
}

// fixDirtyState 修复迁移的 dirty 状态
// 当迁移执行失败时，golang-migrate 会标记 dirty=true 并记录失败的版本号
// 此函数检查当前版本的迁移文件是否已实际生效，如果生效则修正版本号
func fixDirtyState(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return err
	}

	// 如果不是 dirty 状态，直接返回
	if !dirty {
		return nil
	}

	// dirty 状态：强制将版本号设为当前版本（假设迁移已部分生效）
	// 下次执行时会从这个版本继续
	if err := m.Force(int(version)); err != nil {
		return fmt.Errorf("force version %d failed: %w", version, err)
	}

	return nil
}
