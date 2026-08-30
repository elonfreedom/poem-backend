package user

import (
	"strings"

	"github.com/go-fuego/fuego"

	usermodel "poem-backend/internal/model/user"
	userservice "poem-backend/internal/service/user"
	"poem-backend/pkg/response"
)

type SharedPlanHandler struct {
	sharedPlanService *userservice.SharedPlanService
}

func NewSharedPlanHandler(sharedPlanService *userservice.SharedPlanService) *SharedPlanHandler {
	return &SharedPlanHandler{sharedPlanService: sharedPlanService}
}

// ==================== 共享计划管理 ====================

// CreateSharedPlan 创建并发布共享计划
func (h *SharedPlanHandler) CreateSharedPlan(c fuego.ContextWithBody[usermodel.CreateSharedPlanRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.sharedPlanService.CreateSharedPlan(c.Context(), userID, &body)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetSharedPlan 获取共享计划详情（含诗文列表）
func (h *SharedPlanHandler) GetSharedPlan(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	id, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	result, err := h.sharedPlanService.GetSharedPlan(c.Context(), id)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(map[string]any{
		"id":              result.ID,
		"title":           result.Title,
		"description":     result.Description,
		"tags":            result.Tags,
		"total_days":      result.TotalDays,
		"daily_count":     result.DailyCount,
		"subscribe_count": result.SubscribeCount,
		"creator_name":    result.CreatorName,
		"created_at":      result.CreatedAt,
		"poem_ids":        result.PoemIDs,
		"poems":           result.Poems,
	}), nil
}

// UpdateSharedPlan 更新共享计划
func (h *SharedPlanHandler) UpdateSharedPlan(c fuego.ContextWithBody[usermodel.UpdateSharedPlanRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	id, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.sharedPlanService.UpdateSharedPlan(c.Context(), id, userID, &body); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusUpdated), nil
}

// PublishPlan 发布计划
func (h *SharedPlanHandler) PublishPlan(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	id, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	if err := h.sharedPlanService.PublishPlan(c.Context(), id, userID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusPublished), nil
}

// UnpublishPlan 取消发布
func (h *SharedPlanHandler) UnpublishPlan(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	id, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	if err := h.sharedPlanService.UnpublishPlan(c.Context(), id, userID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusUnpublished), nil
}

// DeleteSharedPlan 删除共享计划
func (h *SharedPlanHandler) DeleteSharedPlan(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	id, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	if err := h.sharedPlanService.DeleteSharedPlan(c.Context(), id, userID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusDeleted), nil
}

// ListSharedPlans 浏览共享库
func (h *SharedPlanHandler) ListSharedPlans(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	page, pageSize := ParsePageParams(c)
	keyword := c.QueryParam("q")
	sortBy := c.QueryParam("sort")

	// 解析标签筛选
	var tags []string
	if tagsStr := c.QueryParam("tags"); tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	list, total, err := h.sharedPlanService.ListSharedPlans(c.Context(), page, pageSize, keyword, tags, sortBy)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(PageResponse[usermodel.SharedPlanListItem]{
		Items: list,
		Total: int64(total),
	}), nil
}

// GetMySharedPlans 获取我创建的计划
func (h *SharedPlanHandler) GetMySharedPlans(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	list, err := h.sharedPlanService.GetMySharedPlans(c.Context(), userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(list), nil
}

// ==================== 订阅管理 ====================

// Subscribe 订阅计划
func (h *SharedPlanHandler) Subscribe(c fuego.ContextWithBody[usermodel.SubscribeRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	id, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		body = usermodel.SubscribeRequest{}
	}

	result, err := h.sharedPlanService.Subscribe(c.Context(), userID, id, body.StartDate)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// Unsubscribe 取消订阅
func (h *SharedPlanHandler) Unsubscribe(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	id, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	if err := h.sharedPlanService.Unsubscribe(c.Context(), userID, id); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusUnsubscribed), nil
}

// SetStartDate 设置开始日期
func (h *SharedPlanHandler) SetStartDate(c fuego.ContextWithBody[usermodel.SetStartDateRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	id, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	if err := h.sharedPlanService.SetStartDate(c.Context(), id, userID, body.StartDate); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusUpdated), nil
}

// GetMySubscriptions 获取我的订阅列表
func (h *SharedPlanHandler) GetMySubscriptions(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	list, err := h.sharedPlanService.GetMySubscriptions(c.Context(), userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(PageResponse[usermodel.SubscribeListResponse]{
		Items: list,
		Total: int64(len(list)),
	}), nil
}

// ==================== 每日诗文 & 打卡 ====================

// GetTodayPoem 获取今日诗文
func (h *SharedPlanHandler) GetTodayPoem(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	subID, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	result, err := h.sharedPlanService.GetTodayPoem(c.Context(), subID, userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// Checkin 打卡
func (h *SharedPlanHandler) Checkin(c fuego.ContextWithBody[usermodel.CheckinRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	subID, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.sharedPlanService.Checkin(c.Context(), subID, userID, body.PoemIDs)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// SkipDay 跳过当前天，返回下一首未打卡的诗文
func (h *SharedPlanHandler) SkipDay(c fuego.ContextWithBody[usermodel.SkipDayRequest]) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	subID, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid body", Detail: err.Error()}
	}

	result, err := h.sharedPlanService.SkipDay(c.Context(), subID, userID, body.CurrentDay)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetCheckins 获取订阅的打卡记录列表（用于热力图）
func (h *SharedPlanHandler) GetCheckins(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	subID, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	result, err := h.sharedPlanService.GetCheckins(c.Context(), subID, userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// GetSubscriptionProgress 获取订阅进度
func (h *SharedPlanHandler) GetSubscriptionProgress(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	subID, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	result, err := h.sharedPlanService.GetSubscriptionProgress(c.Context(), subID, userID)
	if err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(result), nil
}

// PauseSubscription 暂停订阅
func (h *SharedPlanHandler) PauseSubscription(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	subID, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	if err := h.sharedPlanService.PauseSubscription(c.Context(), subID, userID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusPaused), nil
}

// ResumeSubscription 恢复订阅
func (h *SharedPlanHandler) ResumeSubscription(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return nil, err
	}

	subID, err := ParsePathID(c, "id")
	if err != nil {
		return nil, err
	}

	if err := h.sharedPlanService.ResumeSubscription(c.Context(), subID, userID); err != nil {
		return nil, err // 透传 Service 错误
	}

	return response.Success(StatusResumed), nil
}

