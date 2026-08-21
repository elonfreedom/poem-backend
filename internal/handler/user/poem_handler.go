package user

import (
	"strconv"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	"poem-backend/internal/model"
	userservice "poem-backend/internal/service/user"
)

type PoemHandler struct {
	poemService *userservice.PoemService
}

func NewPoemHandler(poemService *userservice.PoemService) *PoemHandler {
	return &PoemHandler{poemService: poemService}
}

// List 获取诗歌列表
func (h *PoemHandler) List(c fuego.ContextNoBody) (*model.PoemListResponse, error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	var categoryID *int64
	if cid := c.QueryParam("category_id"); cid != "" {
		id, err := strconv.ParseInt(cid, 10, 64)
		if err == nil {
			categoryID = &id
		}
	}

	result, err := h.poemService.List(c.Context(), page, pageSize, categoryID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// GetByID 获取诗歌详情
func (h *PoemHandler) GetByID(c fuego.ContextNoBody) (*model.PoemResponse, error) {
	poemID, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "无效的诗歌 ID"}
	}

	var userID *string
	uid := middleware.GetUserIDFromContext(c.Context())
	if uid != "" {
		userID = &uid
	}

	result, err := h.poemService.GetByID(c.Context(), poemID, userID)
	if err != nil {
		return nil, fuego.NotFoundError{Title: "not found", Detail: err.Error()}
	}

	return result, nil
}

// Search 搜索诗歌
func (h *PoemHandler) Search(c fuego.ContextNoBody) (*model.PoemListResponse, error) {
	keyword := c.QueryParam("keyword")
	if keyword == "" {
		return nil, fuego.BadRequestError{Title: "missing keyword", Detail: "搜索关键词不能为空"}
	}

	result, err := h.poemService.Search(c.Context(), keyword, 1, 10)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// GetDaily 获取每日推荐
func (h *PoemHandler) GetDaily(c fuego.ContextNoBody) (*model.PoemResponse, error) {
	result, err := h.poemService.GetDailyRecommendation(c.Context())
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}
