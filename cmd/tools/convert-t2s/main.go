// convert-t2s 批量为已有诗歌生成简体（繁体 → 简体）
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

	"poem-backend/pkg/convert"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "预览模式，不修改数据库")
	batchSize := flag.Int("batch-size", 100, "每批处理数量")
	sqlOnly := flag.Bool("sql-only", false, "仅输出 SQL 语句到 stdout")
	dsn := flag.String("dsn", "", "数据库连接字符串（默认从环境变量读取）")
	flag.Parse()

	ctx := context.Background()

	if *sqlOnly {
		if err := generateSQL(ctx, *dsn, *dryRun, *batchSize); err != nil {
			log.Fatalf("生成 SQL 失败: %v", err)
		}
		return
	}

	if err := updateDatabase(ctx, *dsn, *dryRun, *batchSize); err != nil {
		log.Fatalf("更新失败: %v", err)
	}
}

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
		fmt.Println("没有需要生成简体的记录")
		return nil
	}

	fmt.Printf("找到 %d 条需要生成简体的记录\n", len(poems))

	if dryRun {
		fmt.Println("【预览模式】不修改数据库")
		for _, p := range poems {
			titleSC, _ := convert.TraditionalToSimplified(p.title)
			authorSC, _ := convert.TraditionalToSimplified(p.author)
			contentSC, _ := convert.TraditionalToSimplified(p.content)
			fmt.Printf("ID=%d: %s → %s\n", p.id, p.title, titleSC)
			if authorSC != "" {
				fmt.Printf("  作者: %s → %s\n", p.author, authorSC)
			}
			if contentSC != "" {
				fmt.Printf("  正文: %.30s → %.30s\n", p.content, contentSC)
			}
		}
		return nil
	}

	successCount := 0
	for _, p := range poems {
		titleSC, err := convert.TraditionalToSimplified(p.title)
		if err != nil {
			log.Printf("转换 ID=%d 标题失败: %v", p.id, err)
			continue
		}
		authorSC, err := convert.TraditionalToSimplified(p.author)
		if err != nil {
			log.Printf("转换 ID=%d 作者失败: %v", p.id, err)
			continue
		}
		contentSC, err := convert.TraditionalToSimplified(p.content)
		if err != nil {
			log.Printf("转换 ID=%d 正文失败: %v", p.id, err)
			continue
		}

		_, err = db.Exec(ctx, `
			UPDATE poems SET title_sc = $1, author_sc = $2, content_sc = $3, updated_at = NOW()
			WHERE id = $4
		`, titleSC, authorSC, contentSC, p.id)
		if err != nil {
			log.Printf("更新 ID=%d 失败: %v", p.id, err)
			continue
		}
		successCount++
	}

	fmt.Printf("成功更新 %d/%d 条记录\n", successCount, len(poems))
	return nil
}

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
		fmt.Println("-- 没有需要生成简体的记录")
		return nil
	}

	fmt.Printf("-- 共 %d 条记录需要生成简体\n", len(poems))

	if dryRun {
		fmt.Println("-- 【预览模式】")
		for _, p := range poems {
			titleSC, _ := convert.TraditionalToSimplified(p.title)
		authorSC, _ := convert.TraditionalToSimplified(p.author)
			fmt.Printf("-- ID=%d: %s → %s, 作者: %s → %s\n", p.id, p.title, titleSC, p.author, authorSC)
		}
		return nil
	}

	for _, p := range poems {
		titleSC, err := convert.TraditionalToSimplified(p.title)
		if err != nil {
			log.Printf("-- 转换 ID=%d 标题失败: %v", p.id, err)
			continue
		}
		authorSC, err := convert.TraditionalToSimplified(p.author)
		if err != nil {
			log.Printf("-- 转换 ID=%d 作者失败: %v", p.id, err)
			continue
		}
		contentSC, err := convert.TraditionalToSimplified(p.content)
		if err != nil {
			log.Printf("-- 转换 ID=%d 正文失败: %v", p.id, err)
			continue
		}

		fmt.Printf("UPDATE poems SET title_sc = '%s', author_sc = '%s', content_sc = '%s', updated_at = NOW() WHERE id = %d;\n",
			escapeSQL(titleSC), escapeSQL(authorSC), escapeSQL(contentSC), p.id)
	}

	return nil
}

type poemData struct {
	id      int64
	title   string
	author  string
	content string
}

func fetchPoems(ctx context.Context, db *pgxpool.Pool, batchSize int) ([]poemData, error) {
	rows, err := db.Query(ctx, `
		SELECT id, title, author, content FROM poems
		WHERE title_sc = '' OR title_sc IS NULL
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
