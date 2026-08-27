package repository

import (
	"context"
	"log"

	"poem-backend/pkg/pinyin"
)

// EnsurePinyinForAllPoems 为所有缺少拼音的诗歌生成拼音
// 应在迁移完成后调用，确保存量数据也有拼音
func (r *PoemRepository) EnsurePinyinForAllPoems(ctx context.Context) error {
	// 查询需要生成拼音的记录
	rows, err := r.db.Query(ctx, `
		SELECT id, title, author, content FROM poems
		WHERE title_pinyin = '' OR title_pinyin IS NULL
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type poemData struct {
		id      int64
		title   string
		author  string
		content string
	}
	var poems []poemData
	for rows.Next() {
		var p poemData
		if err := rows.Scan(&p.id, &p.title, &p.author, &p.content); err != nil {
			return err
		}
		poems = append(poems, p)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(poems) == 0 {
		return nil
	}

	log.Printf("正在为 %d 首诗歌生成拼音...", len(poems))

	successCount := 0
	for _, p := range poems {
		titlePinyin := pinyin.ToPinyin(p.title)
		contentPinyin := pinyin.ToPinyinLines(p.content)
		authorPinyin := pinyin.ToPinyin(p.author)

		_, err := r.db.Exec(ctx, `
			UPDATE poems SET title_pinyin = $1, content_pinyin = $2, author_pinyin = $3, updated_at = NOW()
			WHERE id = $4
		`, titlePinyin, contentPinyin, authorPinyin, p.id)
		if err != nil {
			log.Printf("更新 ID=%d 失败: %v", p.id, err)
			continue
		}
		successCount++
	}

	log.Printf("拼音生成完成: 成功 %d/%d", successCount, len(poems))
	return nil
}
