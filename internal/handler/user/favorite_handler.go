package user

import (
	"strconv"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	usermodel "poem-backend/internal/model/user"
	userservice "poem-backend/internal/service/user"
)

type FavoriteHandler struct {
	favoriteService *userservice.FavoriteService
}

func NewFavoriteHandler(favoriteService *userservice.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

// AddFavorite 添加收藏
func (h *FavoriteHandler) AddFavorite(c fuego.ContextWithBody[usermodel.FavoriteRequest]) (map[string]string, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.favoriteService.AddFavorite(c.Context(), userID, body.PoemID); err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return map[string]string{"status": "favorited"}, nil
}

// RemoveFavorite 取消收藏
func (h *FavoriteHandler) RemoveFavorite(c fuego.ContextNoBody) (map[string]string, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	poemID, err := strconv.ParseInt(c.PathParam("poem_id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "无效的诗歌 ID"}
	}

	if err := h.favoriteService.RemoveFavorite(c.Context(), userID, poemID); err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return map[string]string{"status": "unfavorited"}, nil
}

// ListFavorites 获取收藏列表
func (h *FavoriteHandler) ListFavorites(c fuego.ContextNoBody) (*usermodel.FavoriteListResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	result, err := h.favoriteService.ListFavorites(c.Context(), userID, 1, 10)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}
