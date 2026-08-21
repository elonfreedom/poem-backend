package user

import (
	"strconv"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/middleware"
	"poem-backend/internal/model"
	userservice "poem-backend/internal/service/user"
)

type ReadingPlanHandler struct {
	readingPlanService *userservice.ReadingPlanService
}

func NewReadingPlanHandler(readingPlanService *userservice.ReadingPlanService) *ReadingPlanHandler {
	return &ReadingPlanHandler{readingPlanService: readingPlanService}
}

// CreatePlan 创建阅读计划
func (h *ReadingPlanHandler) CreatePlan(c fuego.ContextWithBody[model.CreatePlanRequest]) (*model.CreatePlanResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.readingPlanService.CreatePlan(c.Context(), userID, &body)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "bad request", Detail: err.Error()}
	}

	return result, nil
}

// GetCurrentPlan 获取当前计划
func (h *ReadingPlanHandler) GetCurrentPlan(c fuego.ContextNoBody) (*model.CurrentPlanResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	result, err := h.readingPlanService.GetCurrentPlan(c.Context(), userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// PausePlan 暂停计划
func (h *ReadingPlanHandler) PausePlan(c fuego.ContextNoBody) (map[string]string, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	planID, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "无效的计划 ID"}
	}

	if err := h.readingPlanService.PausePlan(c.Context(), userID, planID); err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return map[string]string{"status": "paused"}, nil
}

// ResumePlan 恢复计划
func (h *ReadingPlanHandler) ResumePlan(c fuego.ContextNoBody) (map[string]string, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	planID, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "无效的计划 ID"}
	}

	if err := h.readingPlanService.ResumePlan(c.Context(), userID, planID); err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return map[string]string{"status": "resumed"}, nil
}

// GetPlanProgress 获取计划进度
func (h *ReadingPlanHandler) GetPlanProgress(c fuego.ContextNoBody) (*model.PlanProgressResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	planID, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "无效的计划 ID"}
	}

	result, err := h.readingPlanService.GetPlanProgress(c.Context(), userID, planID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}

// LogReading 记录阅读
func (h *ReadingPlanHandler) LogReading(c fuego.ContextWithBody[model.LogReadingRequest]) (*model.LogReadingResponse, error) {
	userID := middleware.GetUserIDFromContext(c.Context())
	if userID == "" {
		return nil, fuego.UnauthorizedError{Title: "unauthorized", Detail: "未登录"}
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.readingPlanService.LogReading(c.Context(), userID, body.PoemIDs)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "internal error", Detail: err.Error()}
	}

	return result, nil
}
