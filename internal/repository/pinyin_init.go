package repository

import (
	"context"
	"log"

	"poem-backend/pkg/convert"
	"poem-backend/pkg/pinyin"
)

// EnsurePinyinForAllPoems 为所有缺少拼音的诗歌生成拼音
// 应在迁移完成后调用，确保存量数据也有拼音
// 返回成功处理的记录数
func (r *PoemRepository) EnsurePinyinForAllPoems(ctx context.Context) (int, error) {
	// 查询需要生成拼音的记录
	rows, err := r.db.Query(ctx, `
		SELECT id, title, author, content FROM poems
		WHERE title_pinyin = '' OR title_pinyin IS NULL
		ORDER BY id
	`)
	if err != nil {
		return 0, err
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
			return 0, err
		}
		poems = append(poems, p)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	if len(poems) == 0 {
		return 0, nil
	}

	log.Printf("正在为 %d 首诗歌生成拼音...", len(poems))

	successCount := 0
	for _, p := range poems {
		titlePinyin := pinyin.ToPinyin(p.title)
		contentPinyin := pinyin.ToPinyinLines(p.content)

		_, err := r.db.Exec(ctx, `
			UPDATE poems SET title_pinyin = $1, content_pinyin = $2, updated_at = NOW()
			WHERE id = $3
		`, titlePinyin, contentPinyin, p.id)
		if err != nil {
			log.Printf("更新 ID=%d 失败: %v", p.id, err)
			continue
		}
		successCount++
	}

	log.Printf("拼音生成完成: 成功 %d/%d", successCount, len(poems))
	return successCount, nil
}

// EnsureSimplifiedForAllPoems 为所有缺少简体的诗歌生成简体（繁体 → 简体）
// 应在迁移完成后调用，确保存量数据也有简体字段
func (r *PoemRepository) EnsureSimplifiedForAllPoems(ctx context.Context) (int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, author, content, translation, appreciation FROM poems
		WHERE title_sc = '' OR title_sc IS NULL
		   OR author_sc = '' OR author_sc IS NULL
		   OR content_sc = '' OR content_sc IS NULL
		   OR translation_sc = '' OR translation_sc IS NULL
		   OR appreciation_sc = '' OR appreciation_sc IS NULL
		ORDER BY id
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type poemData struct {
		id            int64
		title         string
		author        string
		content       string
		translation   string
		appreciation  string
	}
	var poems []poemData
	for rows.Next() {
		var p poemData
		if err := rows.Scan(&p.id, &p.title, &p.author, &p.content, &p.translation, &p.appreciation); err != nil {
			return 0, err
		}
		poems = append(poems, p)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	if len(poems) == 0 {
		return 0, nil
	}

	log.Printf("正在为 %d 首诗歌生成简体...", len(poems))

	successCount := 0
	for _, p := range poems {
		titleSC := convert.MustTraditionalToSimplified(p.title)
		authorSC := convert.MustTraditionalToSimplified(p.author)
		contentSC := convert.MustTraditionalToSimplified(p.content)
		translationSC := convert.MustTraditionalToSimplified(p.translation)
		appreciationSC := convert.MustTraditionalToSimplified(p.appreciation)

		_, err := r.db.Exec(ctx, `
			UPDATE poems SET title_sc = $1, author_sc = $2, content_sc = $3,
			                   translation_sc = $4, appreciation_sc = $5, updated_at = NOW()
			WHERE id = $6
		`, titleSC, authorSC, contentSC, translationSC, appreciationSC, p.id)
		if err != nil {
			log.Printf("更新 ID=%d 失败: %v", p.id, err)
			continue
		}
		successCount++
	}

	log.Printf("简体生成完成: 成功 %d/%d", successCount, len(poems))
	return successCount, nil
}
