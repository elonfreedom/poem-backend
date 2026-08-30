package user

import (
	"github.com/go-fuego/fuego"

	usermodel "poem-backend/internal/model/user"
	userservice "poem-backend/internal/service/user"
	"poem-backend/pkg/response"
)

type FavoriteHandler struct {
	favoriteService *userservice.FavoriteService
}

func NewFavoriteHandler(favoriteService *userservice.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

// AddFavorite 添加收藏
func (h *FavoriteHandler) AddFavorite(c fuego.ContextWithBody[usermodel.FavoriteRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.favoriteService.AddFavorite(c.Context(), userID, body.PoemID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusFavorited), nil
}

// RemoveFavorite 取消收藏
func (h *FavoriteHandler) RemoveFavorite(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	poemID, err := ParsePathID(c, "poem_id")
	if err != nil {
		return nil, err
	}

	if err := h.favoriteService.RemoveFavorite(c.Context(), userID, poemID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusUnfavorited), nil
}

// ListFavorites 获取收藏列表
// 响应结构: {code: 0, message: "ok", data: {items: [...], total: N}}
func (h *FavoriteHandler) ListFavorites(c fuego.ContextNoBody) (*response.APIResponse[response.PageData[usermodel.FavoriteResponse]], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	result, err := h.favoriteService.ListFavorites(c.Context(), userID, 1, 10)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.PageOK(result.List, int64(result.Total)), nil
}
