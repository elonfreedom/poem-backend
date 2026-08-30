package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5"

	usermodel "poem-backend/internal/model/user"
	"poem-backend/internal/repository"
)

type PoemService struct {
	poemRepo *repository.PoemRepository
}

func NewPoemService(poemRepo *repository.PoemRepository) *PoemService {
	return &PoemService{poemRepo: poemRepo}
}

// List 获取诗歌列表
func (s *PoemService) List(ctx context.Context, page, pageSize int, categoryID *int64, dynasty string) (*usermodel.PoemListResponse, error) {
	poems, total, err := s.poemRepo.List(ctx, page, pageSize, categoryID, "published", dynasty)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询诗歌列表失败: %v", err)}
	}

	list := make([]usermodel.PoemListItem, 0, len(poems))
	for _, p := range poems {
		list = append(list, usermodel.PoemListItem{
			ID:       p.ID,
			Title:    p.Title,
			Author:   p.Author,
			Dynasty:  p.Dynasty,
			CoverURL: p.CoverURL,
		})
	}

	return &usermodel.PoemListResponse{
		Total: int(total),
		Items: list,
	}, nil
}

// GetByID 获取诗歌详情
func (s *PoemService) GetByID(ctx context.Context, poemID int64, userID *string) (*usermodel.PoemResponse, error) {
	poem, err := s.poemRepo.GetByID(ctx, poemID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "poem not found", Detail: fmt.Sprintf("诗歌不存在: id=%d", poemID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询诗歌失败: %v", err)}
	}

	// 记录浏览（忽略错误，不影响主流程）
	_ = s.poemRepo.RecordView(ctx, poemID, userID)

	// 检查是否已收藏
	isFavorited := false
	if userID != nil {
		isFavorited, _ = s.poemRepo.IsFavorited(ctx, *userID, poemID)
	}

	resp := usermodel.PoemResponse{
		ID:             poem.ID,
		Title:          poem.Title,
		Author:         poem.Author,
		Dynasty:        poem.Dynasty,
		Content:        poem.Content,
		Translation:    poem.Translation,
		Appreciation:   poem.Appreciation,
		Tags:           poem.Tags,
		CoverURL:       poem.CoverURL,
		IsFavorited:    isFavorited,
		TitlePinyin:    poem.TitlePinyin,
		ContentPinyin:  poem.ContentPinyin,
		TitleSC:        poem.TitleSC,
		AuthorSC:       poem.AuthorSC,
		ContentSC:      poem.ContentSC,
		TranslationSC:  poem.TranslationSC,
		AppreciationSC: poem.AppreciationSC,
	}

	return &resp, nil
}

// Search 搜索诗歌（返回完整诗文数据，支持 PoemCard 渲染）
func (s *PoemService) Search(ctx context.Context, keyword string, page, pageSize int) (*usermodel.SearchResponse, error) {
	poems, total, err := s.poemRepo.Search(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("搜索诗歌失败: %v", err)}
	}

	list := make([]usermodel.PoemResponse, 0, len(poems))
	for _, p := range poems {
		list = append(list, usermodel.PoemResponse{
			ID:             p.ID,
			Title:          p.Title,
			Author:         p.Author,
			Dynasty:        p.Dynasty,
			Content:        p.Content,
			Translation:    p.Translation,
			Appreciation:   p.Appreciation,
			Tags:           p.Tags,
			CoverURL:       p.CoverURL,
			TitlePinyin:    p.TitlePinyin,
			ContentPinyin:  p.ContentPinyin,
			TitleSC:        p.TitleSC,
			AuthorSC:       p.AuthorSC,
			ContentSC:      p.ContentSC,
			TranslationSC:  p.TranslationSC,
			AppreciationSC: p.AppreciationSC,
		})
	}

	return &usermodel.SearchResponse{
		Total: int(total),
		Items: list,
	}, nil
}

// GetDailyRecommendation 获取每日推荐
func (s *PoemService) GetDailyRecommendation(ctx context.Context) (*usermodel.PoemResponse, error) {
	poem, err := s.poemRepo.GetDailyRecommendation(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "no poems available", Detail: "请先通过管理后台导入诗歌数据"}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("获取每日推荐失败: %v", err)}
	}

	resp := usermodel.PoemResponse{
		ID:             poem.ID,
		Title:          poem.Title,
		Author:         poem.Author,
		Dynasty:        poem.Dynasty,
		Content:        poem.Content,
		Translation:    poem.Translation,
		Appreciation:   poem.Appreciation,
		Tags:           poem.Tags,
		CoverURL:       poem.CoverURL,
		TitlePinyin:    poem.TitlePinyin,
		ContentPinyin:  poem.ContentPinyin,
		TitleSC:        poem.TitleSC,
		AuthorSC:       poem.AuthorSC,
		ContentSC:      poem.ContentSC,
		TranslationSC:  poem.TranslationSC,
		AppreciationSC: poem.AppreciationSC,
	}

	return &resp, nil
}
