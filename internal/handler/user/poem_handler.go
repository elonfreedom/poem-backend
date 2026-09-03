package user

import (
	"strconv"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	userservice "poem-backend/internal/service/user"
	"poem-backend/pkg/response"
)

type PoemHandler struct {
	poemService *userservice.PoemService
}

func NewPoemHandler(poemService *userservice.PoemService) *PoemHandler {
	return &PoemHandler{poemService: poemService}
}

// List 获取诗歌列表
func (h *PoemHandler) List(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}

	var categoryID *int64
	if cid := c.QueryParam("category"); cid != "" {
		id, err := strconv.ParseInt(cid, 10, 64)
		if err == nil {
			categoryID = &id
		}
	}

	dynasty := c.QueryParam("dynasty")

	result, err := h.poemService.List(c.Context(), page, perPage, categoryID, dynasty)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetByID 获取诗歌详情
func (h *PoemHandler) GetByID(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	poemID, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	var userID *string
	uid := middleware.GetUserIDFromContext(c.Context())
	if uid != "" {
		userID = &uid
	}

	result, err := h.poemService.GetByID(c.Context(), poemID, userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// Search 搜索诗歌
func (h *PoemHandler) Search(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	keyword := c.QueryParam("q")
	if keyword == "" {
		return nil, fuego.BadRequestError{Title: "missing keyword", Detail: "搜索关键词不能为空"}
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}

	searchScope := c.QueryParam("search_scope")
	result, err := h.poemService.Search(c.Context(), keyword, page, perPage, searchScope)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetDaily 获取每日推荐
func (h *PoemHandler) GetDaily(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	result, err := h.poemService.GetDailyRecommendation(c.Context())
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}
