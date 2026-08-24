package user

import (
	"context"
	"fmt"

	usermodel "poem-backend/internal/model/user"
	"poem-backend/internal/repository"
)

type FavoriteService struct {
	favoriteRepo *repository.FavoriteRepository
	poemRepo     *repository.PoemRepository
}

func NewFavoriteService(
	favoriteRepo *repository.FavoriteRepository,
	poemRepo *repository.PoemRepository,
) *FavoriteService {
	return &FavoriteService{
		favoriteRepo: favoriteRepo,
		poemRepo:     poemRepo,
	}
}

// AddFavorite 添加收藏
func (s *FavoriteService) AddFavorite(ctx context.Context, userID string, poemID int64) error {
	// 检查诗歌是否存在
	_, err := s.poemRepo.GetByID(ctx, poemID)
	if err != nil {
		return fmt.Errorf("poem not found: %w", err)
	}

	return s.favoriteRepo.Create(ctx, userID, poemID)
}

// RemoveFavorite 取消收藏
func (s *FavoriteService) RemoveFavorite(ctx context.Context, userID string, poemID int64) error {
	return s.favoriteRepo.Delete(ctx, userID, poemID)
}

// ListFavorites 获取收藏列表
func (s *FavoriteService) ListFavorites(ctx context.Context, userID string, page, pageSize int) (*usermodel.FavoriteListResponse, error) {
	favorites, total, err := s.favoriteRepo.List(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list favorites: %w", err)
	}

	var list []usermodel.FavoriteResponse
	for _, f := range favorites {
		poem, err := s.poemRepo.GetByID(ctx, f.PoemID)
		if err != nil {
			continue // 跳过已删除的诗歌
		}
		list = append(list, usermodel.FavoriteResponse{
			Poem: usermodel.PoemListItem{
				ID:       poem.ID,
				Title:    poem.Title,
				Author:   poem.Author,
				Dynasty:  poem.Dynasty,
				CoverURL: poem.CoverURL,
			},
			CreatedAt: f.CreatedAt,
		})
	}

	return &usermodel.FavoriteListResponse{
		Total: int(total),
		List:  list,
	}, nil
}
