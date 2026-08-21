package user

import (
	"context"
	"fmt"

	"poem-backend/internal/model"
	"poem-backend/internal/repository"
)

type PoemService struct {
	poemRepo *repository.PoemRepository
}

func NewPoemService(poemRepo *repository.PoemRepository) *PoemService {
	return &PoemService{poemRepo: poemRepo}
}

// List 获取诗歌列表
func (s *PoemService) List(ctx context.Context, page, pageSize int, categoryID *int64) (*model.PoemListResponse, error) {
	poems, total, err := s.poemRepo.List(ctx, page, pageSize, categoryID, "published")
	if err != nil {
		return nil, fmt.Errorf("failed to list poems: %w", err)
	}

	list := make([]model.PoemListItem, 0, len(poems))
	for _, p := range poems {
		list = append(list, model.PoemListItem{
			ID:    p.ID,
			Title: p.Title,
			Author: p.Author,
			Dynasty: p.Dynasty,
			CoverURL: p.CoverURL,
		})
	}

	return &model.PoemListResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetByID 获取诗歌详情
func (s *PoemService) GetByID(ctx context.Context, poemID int64, userID *string) (*model.PoemResponse, error) {
	poem, err := s.poemRepo.GetByID(ctx, poemID)
	if err != nil {
		return nil, fmt.Errorf("poem not found: %w", err)
	}

	// 记录浏览
	s.poemRepo.RecordView(ctx, poemID, userID)

	// 检查是否已收藏
	isFavorited := false
	if userID != nil {
		isFavorited, _ = s.poemRepo.IsFavorited(ctx, *userID, poemID)
	}

	resp := model.PoemResponse{
		ID:          poem.ID,
		Title:       poem.Title,
		Author:      poem.Author,
		Dynasty:     poem.Dynasty,
		Content:     poem.Content,
		Translation: poem.Translation,
		Appreciation: poem.Appreciation,
		Tags:        poem.Tags,
		CoverURL:    poem.CoverURL,
		IsFavorited: isFavorited,
	}

	return &resp, nil
}

// Search 搜索诗歌
func (s *PoemService) Search(ctx context.Context, keyword string, page, pageSize int) (*model.PoemListResponse, error) {
	poems, total, err := s.poemRepo.Search(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search poems: %w", err)
	}

	list := make([]model.PoemListItem, 0, len(poems))
	for _, p := range poems {
		list = append(list, model.PoemListItem{
			ID:    p.ID,
			Title: p.Title,
			Author: p.Author,
			Dynasty: p.Dynasty,
			CoverURL: p.CoverURL,
		})
	}

	return &model.PoemListResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetDailyRecommendation 获取每日推荐
func (s *PoemService) GetDailyRecommendation(ctx context.Context) (*model.PoemResponse, error) {
	poem, err := s.poemRepo.GetDailyRecommendation(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily recommendation: %w", err)
	}

	resp := model.PoemResponse{
		ID:          poem.ID,
		Title:       poem.Title,
		Author:      poem.Author,
		Dynasty:     poem.Dynasty,
		Content:     poem.Content,
		Translation: poem.Translation,
		Appreciation: poem.Appreciation,
		Tags:        poem.Tags,
		CoverURL:    poem.CoverURL,
	}

	return &resp, nil
}
