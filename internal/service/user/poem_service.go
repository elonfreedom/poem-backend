package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-fuego/fuego"

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
		return nil, fmt.Errorf("failed to list poems: %w", err)
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
		return nil, fmt.Errorf("poem not found: %w", err)
	}

	// 记录浏览（忽略错误，不影响主流程）
	_ = s.poemRepo.RecordView(ctx, poemID, userID)

	// 检查是否已收藏
	isFavorited := false
	if userID != nil {
		isFavorited, _ = s.poemRepo.IsFavorited(ctx, *userID, poemID)
	}

	resp := usermodel.PoemResponse{
		ID:            poem.ID,
		Title:         poem.Title,
		Author:        poem.Author,
		Dynasty:       poem.Dynasty,
		Content:       poem.Content,
		Translation:   poem.Translation,
		Appreciation:  poem.Appreciation,
		Tags:          poem.Tags,
		CoverURL:      poem.CoverURL,
		IsFavorited:   isFavorited,
		TitlePinyin:   poem.TitlePinyin,
		ContentPinyin: poem.ContentPinyin,
		AuthorPinyin:  poem.AuthorPinyin,
	}

	return &resp, nil
}

// Search 搜索诗歌
func (s *PoemService) Search(ctx context.Context, keyword string, page, pageSize int) (*usermodel.PoemListResponse, error) {
	poems, total, err := s.poemRepo.Search(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search poems: %w", err)
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

// GetDailyRecommendation 获取每日推荐
func (s *PoemService) GetDailyRecommendation(ctx context.Context) (*usermodel.PoemResponse, error) {
	poem, err := s.poemRepo.GetDailyRecommendation(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "暂无诗歌", Detail: "请先通过管理后台导入诗歌数据"}
		}
		return nil, fmt.Errorf("failed to get daily recommendation: %w", err)
	}

	resp := usermodel.PoemResponse{
		ID:            poem.ID,
		Title:         poem.Title,
		Author:        poem.Author,
		Dynasty:       poem.Dynasty,
		Content:       poem.Content,
		Translation:   poem.Translation,
		Appreciation:  poem.Appreciation,
		Tags:          poem.Tags,
		CoverURL:      poem.CoverURL,
		TitlePinyin:   poem.TitlePinyin,
		ContentPinyin: poem.ContentPinyin,
		AuthorPinyin:  poem.AuthorPinyin,
	}

	return &resp, nil
}
