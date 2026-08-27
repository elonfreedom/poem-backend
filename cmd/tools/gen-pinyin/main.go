// gen-pinyin 批量为已有诗歌生成拼音
// 支持两种模式：
//   - 默认：直接连接数据库并更新
//   - --sql-only：输出 SQL 脚本到 stdout（用于无法直接执行二进制的情况）
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/pkg/pinyin"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "预览模式，不修改数据库")
	batchSize := flag.Int("batch-size", 100, "每批处理数量")
	sqlOnly := flag.Bool("sql-only", false, "仅输出 SQL 语句到 stdout")
	dsn := flag.String("dsn", "", "数据库连接字符串（默认从环境变量读取）")
	flag.Parse()

	ctx := context.Background()

	// sql-only 模式：输出 SQL 并退出
	if *sqlOnly {
		if err := generateSQL(ctx, *dsn, *dryRun, *batchSize); err != nil {
			log.Fatalf("生成 SQL 失败: %v", err)
		}
		return
	}

	// 默认模式：直接连接数据库并更新
	if err := updateDatabase(ctx, *dsn, *dryRun, *batchSize); err != nil {
		log.Fatalf("更新失败: %v", err)
	}
}

// getDSN 获取数据库连接字符串
func getDSN(dsn string) string {
	if dsn != "" {
		return dsn
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "poem"),
		getEnv("DB_SSLMODE", "disable"),
	)
}

// updateDatabase 直接连接数据库并更新拼音字段
func updateDatabase(ctx context.Context, dsn string, dryRun bool, batchSize int) error {
	db, err := pgxpool.New(ctx, getDSN(dsn))
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	poems, err := fetchPoems(ctx, db, batchSize)
	if err != nil {
		return err
	}

	if len(poems) == 0 {
		fmt.Println("没有需要生成拼音的记录")
		return nil
	}

	fmt.Printf("找到 %d 条需要生成拼音的记录\n", len(poems))

	if dryRun {
		fmt.Println("【预览模式】不修改数据库")
		for _, p := range poems {
			fmt.Printf("ID=%d: %s → %s\n", p.id, p.title, pinyin.ToPinyin(p.title))
		}
		return nil
	}

	successCount := 0
	for _, p := range poems {
		titlePinyin := pinyin.ToPinyin(p.title)
		contentPinyin := pinyin.ToPinyinLines(p.content)
		authorPinyin := pinyin.ToPinyin(p.author)

		_, err := db.Exec(ctx, `
			UPDATE poems SET title_pinyin = $1, content_pinyin = $2, author_pinyin = $3, updated_at = NOW()
			WHERE id = $4
		`, titlePinyin, contentPinyin, authorPinyin, p.id)
		if err != nil {
			log.Printf("更新 ID=%d 失败: %v", p.id, err)
			continue
		}
		successCount++
	}

	fmt.Printf("成功更新 %d/%d 条记录\n", successCount, len(poems))
	return nil
}

// generateSQL 输出 SQL 更新语句到 stdout
func generateSQL(ctx context.Context, dsn string, dryRun bool, batchSize int) error {
	db, err := pgxpool.New(ctx, getDSN(dsn))
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	poems, err := fetchPoems(ctx, db, batchSize)
	if err != nil {
		return err
	}

	if len(poems) == 0 {
		fmt.Println("-- 没有需要生成拼音的记录")
		return nil
	}

	fmt.Printf("-- 共 %d 条记录需要生成拼音\n", len(poems))

	if dryRun {
		fmt.Println("-- 【预览模式】")
		for _, p := range poems {
			fmt.Printf("-- ID=%d: %s → %s\n", p.id, p.title, pinyin.ToPinyin(p.title))
		}
		return nil
	}

	for _, p := range poems {
		titlePinyin := pinyin.ToPinyin(p.title)
		contentPinyin := pinyin.ToPinyinLines(p.content)
		authorPinyin := pinyin.ToPinyin(p.author)

		fmt.Printf("UPDATE poems SET title_pinyin = '%s', content_pinyin = '%s', author_pinyin = '%s', updated_at = NOW() WHERE id = %d;\n",
			escapeSQL(titlePinyin), escapeSQL(contentPinyin), escapeSQL(authorPinyin), p.id)
	}

	return nil
}

// poemData 诗歌数据
type poemData struct {
	id      int64
	title   string
	author  string
	content string
}

// fetchPoems 查询需要生成拼音的记录
func fetchPoems(ctx context.Context, db *pgxpool.Pool, batchSize int) ([]poemData, error) {
	rows, err := db.Query(ctx, `
		SELECT id, title, author, content FROM poems
		WHERE title_pinyin = '' OR title_pinyin IS NULL
		ORDER BY id
		LIMIT $1
	`, batchSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	defer rows.Close()

	var poems []poemData
	for rows.Next() {
		var p poemData
		if err := rows.Scan(&p.id, &p.title, &p.author, &p.content); err != nil {
			return nil, fmt.Errorf("扫描失败: %w", err)
		}
		poems = append(poems, p)
	}
	return poems, rows.Err()
}

// escapeSQL 转义 SQL 字符串中的单引号
func escapeSQL(s string) string {
	result := ""
	for _, c := range s {
		if c == '\'' {
			result += "''"
		} else {
			result += string(c)
		}
	}
	return result
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
