package repository

import (
	"context"
	"fmt"
	"log"

	adminmodel "poem-backend/internal/model/admin"
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
// 同时补充 authors 表中 name_traditional 为空的记录
// 应在迁移完成后调用，确保存量数据也有简体字段
func (r *PoemRepository) EnsureSimplifiedForAllPoems(ctx context.Context) (int, error) {
	successCount := 0

	// 1. 为诗歌生成简体字段
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
			rows.Close()
			return 0, err
		}
		poems = append(poems, p)
	}
	rows.Close()

	if len(poems) > 0 {
		log.Printf("正在为 %d 首诗歌生成简体...", len(poems))

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
				log.Printf("更新诗歌 ID=%d 失败: %v", p.id, err)
				continue
			}
			successCount++
		}
		log.Printf("诗歌简体生成完成: 成功 %d/%d", successCount, len(poems))
	}

	// 2. 为 authors 表补充 name_traditional
	authorRows, err := r.db.Query(ctx, `
		SELECT id, name FROM authors
		WHERE name_traditional = '' OR name_traditional IS NULL
		ORDER BY id
	`)
	if err != nil {
		return successCount, err
	}

	type authorData struct {
		id   int64
		name string
	}
	var authors []authorData
	for authorRows.Next() {
		var a authorData
		if err := authorRows.Scan(&a.id, &a.name); err != nil {
			authorRows.Close()
			return successCount, err
		}
		authors = append(authors, a)
	}
	authorRows.Close()

	if len(authors) > 0 {
		log.Printf("正在为 %d 个作者生成繁体名...", len(authors))

		authorSuccess := 0
		for _, a := range authors {
			nameTraditional := convert.MustSimplifiedToTraditional(a.name)
			_, err := r.db.Exec(ctx, `
				UPDATE authors SET name_traditional = $1, updated_at = NOW()
				WHERE id = $2
			`, nameTraditional, a.id)
			if err != nil {
				log.Printf("更新作者 ID=%d 失败: %v", a.id, err)
				continue
			}
			authorSuccess++
		}
		log.Printf("作者繁体名生成完成: 成功 %d/%d", authorSuccess, len(authors))
		successCount += authorSuccess
	}

	return successCount, nil
}

// BatchConvertChars 批量转换指定诗歌的字符类型
// target: "simplified" 或 "traditional"
// 转换逻辑：
//   - target=simplified: 将原文视为繁体，转换为简体，填入 sc 字段
//   - target=traditional: 将原文视为简体，转换为繁体，覆盖原文字段
func (r *PoemRepository) BatchConvertChars(ctx context.Context, poetryIDs []int64, target string) (*adminmodel.AdminToolBatchConvertCharsResponse, error) {
	if len(poetryIDs) == 0 {
		return &adminmodel.AdminToolBatchConvertCharsResponse{
			Total: 0, Converted: 0, Message: "未指定诗歌ID",
		}, nil
	}

	// 查询指定 ID 的诗歌
	rows, err := r.db.Query(ctx, `
		SELECT id, title, author, content, translation, appreciation
		FROM poems WHERE id = ANY($1)
		ORDER BY id
	`, poetryIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type poemData struct {
		id           int64
		title        string
		author       string
		content      string
		translation  string
		appreciation string
	}
	var poems []poemData
	for rows.Next() {
		var p poemData
		if err := rows.Scan(&p.id, &p.title, &p.author, &p.content, &p.translation, &p.appreciation); err != nil {
			return nil, err
		}
		poems = append(poems, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(poems) == 0 {
		return &adminmodel.AdminToolBatchConvertCharsResponse{
			Total: 0, Converted: 0, Message: "未找到指定的诗歌",
		}, nil
	}

	log.Printf("正在为 %d 首诗歌批量转换为 %s...", len(poems), target)

	successCount := 0
	for _, p := range poems {
		var err error
		switch target {
		case "simplified":
			// 原文视为繁体 → 生成简体填入 sc 字段
			titleSC := convert.MustTraditionalToSimplified(p.title)
			authorSC := convert.MustTraditionalToSimplified(p.author)
			contentSC := convert.MustTraditionalToSimplified(p.content)
			translationSC := convert.MustTraditionalToSimplified(p.translation)
			appreciationSC := convert.MustTraditionalToSimplified(p.appreciation)
			_, err = r.db.Exec(ctx, `
				UPDATE poems SET title_sc = $1, author_sc = $2, content_sc = $3,
				                   translation_sc = $4, appreciation_sc = $5, updated_at = NOW()
				WHERE id = $6
			`, titleSC, authorSC, contentSC, translationSC, appreciationSC, p.id)
		case "traditional":
			// 原文视为简体 → 生成繁体覆盖原文字段
			titleTC, e1 := convert.SimplifiedToTraditional(p.title)
			authorTC, e2 := convert.SimplifiedToTraditional(p.author)
			contentTC, e3 := convert.SimplifiedToTraditional(p.content)
			translationTC, e4 := convert.SimplifiedToTraditional(p.translation)
			appreciationTC, e5 := convert.SimplifiedToTraditional(p.appreciation)
			if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
				log.Printf("转换 ID=%d 失败，跳过", p.id)
				continue
			}
			_, err = r.db.Exec(ctx, `
				UPDATE poems SET title = $1, author = $2, content = $3,
				                   translation = $4, appreciation = $5, updated_at = NOW()
				WHERE id = $6
			`, titleTC, authorTC, contentTC, translationTC, appreciationTC, p.id)
		}
		if err != nil {
			log.Printf("更新 ID=%d 失败: %v", p.id, err)
			continue
		}
		successCount++
	}

	msg := fmt.Sprintf("字符转换完成: 成功 %d/%d，目标类型: %s", successCount, len(poems), target)
	log.Print(msg)

	return &adminmodel.AdminToolBatchConvertCharsResponse{
		Total:     len(poems),
		Converted: successCount,
		Message:   msg,
	}, nil
}
