package migrate

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/repository"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// expectedSchema 定义迁移完成后必须存在的关键表/列，用于验证迁移是否真正生效
var expectedSchema = []struct {
	table  string
	column string
}{
	{"poems", "title_pinyin"},
	{"poems", "content_pinyin"},
	{"poems", "title_sc"},
	{"poems", "author_sc"},
	{"poems", "content_sc"},
	{"poems", "author_id"},
	{"authors", "id"},
	{"passkeys", "flags"},
	{"shared_plans", "id"},
	{"plan_subscriptions", "id"},
}

// Run 执行数据库迁移（幂等，可重复调用）
// 使用 golang-migrate + pgx 原生驱动，从 embed.FS 读取 migrations/*.sql
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

	// 创建 migrate 实例
	m, err := migrate.NewWithInstance("iofs", srcDriver, "pgx", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// 修复 dirty 状态（设为前一版本，让 Up() 重新执行）
	if err := fixDirtyState(m); err != nil {
		return fmt.Errorf("failed to fix dirty state: %w", err)
	}

	// 执行迁移（幂等迁移文件可安全重跑）
	if err := runMigrations(m); err != nil {
		return err
	}

	// 验证迁移结果：检查关键表/列是否存在
	if err := verifySchema(sqlDB); err != nil {
		log.Printf("Schema verification failed: %v", err)
		// 验证失败时，强制回退并重试一次
		log.Println("Retrying migrations after verification failure...")
		if retryErr := retryMigrations(m); retryErr != nil {
			return fmt.Errorf("migration retry failed: %w (original error: %v)", retryErr, err)
		}
		// 再次验证
		if verifyErr := verifySchema(sqlDB); verifyErr != nil {
			return fmt.Errorf("schema verification failed after retry: %w", verifyErr)
		}
	}

	log.Println("Database migrations applied and verified")

	// 迁移完成后，为存量诗歌生成拼音数据
	poemRepo := repository.NewPoemRepository(db)
	if count, err := poemRepo.EnsurePinyinForAllPoems(context.Background()); err != nil {
		log.Printf("Pinyin generation warning: %v", err)
		// 拼音生成失败不影响服务启动，admin 可以后续手动补充
	} else if count > 0 {
		log.Printf("Pinyin generation completed: %d poems processed", count)
	}

	return nil
}

// runMigrations 执行迁移并记录版本信息
func runMigrations(m *migrate.Migrate) error {
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	// 记录当前迁移版本
	version, dirty, _ := m.Version()
	log.Printf("Migration completed: version=%d, dirty=%v", version, dirty)
	return nil
}

// retryMigrations 强制回退到前一版本后重新执行迁移
func retryMigrations(m *migrate.Migrate) error {
	version, _, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return err
	}

	// 强制回退到前一版本
	if version > 0 {
		if err := m.Force(int(version) - 1); err != nil {
			return fmt.Errorf("force version %d failed: %w", version-1, err)
		}
	}

	// 重新执行迁移
	return m.Up()
}

// fixDirtyState 修复迁移的 dirty 状态
// 当迁移失败被标记为 dirty 时，强制设为前一版本（version-1），
// 这样 m.Up() 会重新执行当前迁移，而不是跳过它。
func fixDirtyState(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return err
	}

	if !dirty {
		return nil
	}

	log.Printf("Found dirty migration at version %d, forcing to %d", version, version-1)

	// dirty 状态：强制将版本号设为前一版本，让 Up() 重新执行当前迁移
	if err := m.Force(int(version) - 1); err != nil {
		return fmt.Errorf("force version %d failed: %w", version-1, err)
	}

	return nil
}

// verifySchema 验证关键表/列是否存在，确保迁移真正生效
func verifySchema(db *sql.DB) error {
	for _, exp := range expectedSchema {
		exists, err := columnExists(db, exp.table, exp.column)
		if err != nil {
			return fmt.Errorf("check column %s.%s failed: %w", exp.table, exp.column, err)
		}
		if !exists {
			return fmt.Errorf("expected column %s.%s does not exist", exp.table, exp.column)
		}
	}
	return nil
}

// columnExists 检查指定表的指定列是否存在
func columnExists(db *sql.DB, table, column string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = $1 AND column_name = $2
		)
	`
	var exists bool
	err := db.QueryRow(query, table, column).Scan(&exists)
	return exists, err
}
