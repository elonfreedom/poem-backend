package user

import (
	"github.com/go-fuego/fuego"

	usermodel "poem-backend/internal/model/user"
	userservice "poem-backend/internal/service/user"
	"poem-backend/pkg/response"
)

type ReadingPlanHandler struct {
	readingPlanService *userservice.ReadingPlanService
}

func NewReadingPlanHandler(readingPlanService *userservice.ReadingPlanService) *ReadingPlanHandler {
	return &ReadingPlanHandler{readingPlanService: readingPlanService}
}

// CreatePlan 创建阅读计划
func (h *ReadingPlanHandler) CreatePlan(c fuego.ContextWithBody[usermodel.CreatePlanRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.readingPlanService.CreatePlan(c.Context(), userID, &body)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetCurrentPlan 获取当前计划
func (h *ReadingPlanHandler) GetCurrentPlan(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	result, err := h.readingPlanService.GetCurrentPlan(c.Context(), userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// PausePlan 暂停计划
func (h *ReadingPlanHandler) PausePlan(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	planID, err := ParsePathInt(c, "id")
	if err != nil {
		return nil, err
	}

	if err := h.readingPlanService.PausePlan(c.Context(), userID, planID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusPaused), nil
}

// ResumePlan 恢复计划
func (h *ReadingPlanHandler) ResumePlan(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	planID, err := ParsePathInt(c, "id")
	if err != nil {
		return nil, err
	}

	if err := h.readingPlanService.ResumePlan(c.Context(), userID, planID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusResumed), nil
}

// GetPlanProgress 获取计划进度
func (h *ReadingPlanHandler) GetPlanProgress(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	planID, err := ParsePathInt(c, "id")
	if err != nil {
		return nil, err
	}

	result, err := h.readingPlanService.GetPlanProgress(c.Context(), userID, planID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// LogReading 记录阅读
func (h *ReadingPlanHandler) LogReading(c fuego.ContextWithBody[usermodel.LogReadingRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.readingPlanService.LogReading(c.Context(), userID, body.PoemIDs)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}
