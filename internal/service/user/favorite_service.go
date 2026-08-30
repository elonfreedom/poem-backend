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
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "poem not found", Detail: fmt.Sprintf("诗歌不存在: id=%d", poemID)}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询诗歌失败: %v", err)}
	}

	if err := s.favoriteRepo.Create(ctx, userID, poemID); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("添加收藏失败: %v", err)}
	}
	return nil
}

// RemoveFavorite 取消收藏
func (s *FavoriteService) RemoveFavorite(ctx context.Context, userID string, poemID int64) error {
	if err := s.favoriteRepo.Delete(ctx, userID, poemID); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("取消收藏失败: %v", err)}
	}
	return nil
}

// ListFavorites 获取收藏列表
func (s *FavoriteService) ListFavorites(ctx context.Context, userID string, page, pageSize int) (*usermodel.FavoriteListResponse, error) {
	favorites, total, err := s.favoriteRepo.List(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询收藏列表失败: %v", err)}
	}

	var list []usermodel.FavoriteResponse
	for _, f := range favorites {
		poem, err := s.poemRepo.GetByID(ctx, f.PoemID)
		if err != nil {
			// 诗歌已删除/下架：仍返回记录，标记 unavailable
			list = append(list, usermodel.FavoriteResponse{
				Poem: usermodel.PoemListItem{
					ID: f.PoemID,
				},
				Available: false,
				CreatedAt: f.CreatedAt,
			})
			continue
		}
		list = append(list, usermodel.FavoriteResponse{
			Poem: usermodel.PoemListItem{
				ID:       poem.ID,
				Title:    poem.Title,
				Author:   poem.Author,
				Dynasty:  poem.Dynasty,
				CoverURL: poem.CoverURL,
			},
			Available: true,
			CreatedAt: f.CreatedAt,
		})
	}

	return &usermodel.FavoriteListResponse{
		Total: int(total),
		List:  list,
	}, nil
}
