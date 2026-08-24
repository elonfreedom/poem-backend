// 诗歌数据导入工具
// 数据来源：https://github.com/chinese-poetry/chinese-poetry (MIT License)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/config"
)

// TangPoem 唐诗结构
type TangPoem struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Paragraphs []string `json:"paragraphs"`
}

// CiPoem 宋词结构
type CiPoem struct {
	Author     string   `json:"author"`
	Paragraphs []string `json:"paragraphs"`
	Rhythmic   string   `json:"rhythmic"`
}

const (
	// 唐诗数据文件（全唐诗约 5.5 万首，分多个文件）
	tangPoetryBaseURL = "https://raw.githubusercontent.com/chinese-poetry/chinese-poetry/master/全唐诗"

	// 宋词数据文件（全宋词约 2.1 万首）
	ciPoetryBaseURL = "https://raw.githubusercontent.com/chinese-poetry/chinese-poetry/master/宋词"
)

func main() {
	cfg := config.Load()

	// 连接数据库
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.Database.ConnString())
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	log.Println("数据库连接成功")

	// 导入唐诗
	importTangPoems(ctx, db)

	// 导入宋词
	importCiPoems(ctx, db)

	log.Println("数据导入完成")
}

// importTangPoems 导入唐诗
func importTangPoems(ctx context.Context, db *pgxpool.Pool) {
	log.Println("开始导入唐诗...")

	// 唐诗文件编号：0, 1000, 2000, ...
	fileIndexes := []int{0, 1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000,
		10000, 11000, 12000, 13000, 14000, 15000, 16000, 17000, 18000, 19000,
		20000, 21000, 22000, 23000, 24000, 25000, 26000, 27000, 28000, 29000,
		30000, 31000, 32000, 33000, 34000, 35000, 36000, 37000, 38000, 39000,
		40000, 41000, 42000, 43000, 44000, 45000, 46000, 47000, 48000, 49000,
		50000, 51000, 52000, 53000, 54000, 55000}

	totalCount := 0
	for _, idx := range fileIndexes {
		filename := fmt.Sprintf("poet.tang.%d.json", idx)
		url := fmt.Sprintf("%s/%s", tangPoetryBaseURL, filename)

		poems, err := downloadAndParseTang(url)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				continue
			}
			log.Printf("下载 %s 失败: %v", filename, err)
			continue
		}

		count, err := insertPoems(ctx, db, poems, "唐")
		if err != nil {
			log.Printf("插入 %s 失败: %v", filename, err)
			continue
		}
		totalCount += count
		log.Printf("导入 %s: %d 首", filename, count)
	}

	log.Printf("唐诗导入完成，共 %d 首", totalCount)
}

// importCiPoems 导入宋词
func importCiPoems(ctx context.Context, db *pgxpool.Pool) {
	log.Println("开始导入宋词...")

	fileIndexes := []int{0, 1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000,
		10000, 11000, 12000, 13000, 14000, 15000, 16000, 17000, 18000, 19000,
		20000, 21000}

	totalCount := 0
	for _, idx := range fileIndexes {
		filename := fmt.Sprintf("ci.song.%d.json", idx)
		url := fmt.Sprintf("%s/%s", ciPoetryBaseURL, filename)

		poems, err := downloadAndParseCi(url)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				continue
			}
			log.Printf("下载 %s 失败: %v", filename, err)
			continue
		}

		count, err := insertCiPoems(ctx, db, poems)
		if err != nil {
			log.Printf("插入 %s 失败: %v", filename, err)
			continue
		}
		totalCount += count
		log.Printf("导入 %s: %d 首", filename, count)
	}

	log.Printf("宋词导入完成，共 %d 首", totalCount)
}

// downloadAndParseTang 下载并解析唐诗
func downloadAndParseTang(url string) ([]TangPoem, error) {
	data, err := downloadJSON(url)
	if err != nil {
		return nil, err
	}

	var poems []TangPoem
	if err := json.Unmarshal(data, &poems); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}
	return poems, nil
}

// downloadAndParseCi 下载并解析宋词
func downloadAndParseCi(url string) ([]CiPoem, error) {
	data, err := downloadJSON(url)
	if err != nil {
		return nil, err
	}

	var poems []CiPoem
	if err := json.Unmarshal(data, &poems); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}
	return poems, nil
}

// downloadJSON 下载 JSON 数据
func downloadJSON(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// insertPoems 批量插入唐诗
func insertPoems(ctx context.Context, db *pgxpool.Pool, poems []TangPoem, dynasty string) (int, error) {
	if len(poems) == 0 {
		return 0, nil
	}

	now := time.Now()
	query := `
		INSERT INTO poems (title, author, dynasty, content, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'published', $5, $6)
	`

	batch := &pgx.Batch{}
	for _, p := range poems {
		content := strings.Join(p.Paragraphs, "\n")
		batch.Queue(query, p.Title, p.Author, dynasty, content, now, now)
	}

	br := db.SendBatch(ctx, batch)
	defer func() { _ = br.Close() }()

	count := 0
	for i := 0; i < len(poems); i++ {
		_, err := br.Exec()
		if err != nil {
			return count, err
		}
		count++
	}

	return count, br.Close()
}

// insertCiPoems 批量插入宋词
func insertCiPoems(ctx context.Context, db *pgxpool.Pool, poems []CiPoem) (int, error) {
	if len(poems) == 0 {
		return 0, nil
	}

	now := time.Now()
	query := `
		INSERT INTO poems (title, author, dynasty, content, tags, status, created_at, updated_at)
		VALUES ($1, $2, '宋', $3, $4, 'published', $5, $6)
	`

	batch := &pgx.Batch{}
	for _, p := range poems {
		content := strings.Join(p.Paragraphs, "\n")
		tags := []string{p.Rhythmic}
		batch.Queue(query, p.Rhythmic, p.Author, content, tags, now, now)
	}

	br := db.SendBatch(ctx, batch)
	defer func() { _ = br.Close() }()

	count := 0
	for i := 0; i < len(poems); i++ {
		_, err := br.Exec()
		if err != nil {
			return count, err
		}
		count++
	}

	return count, br.Close()
}
