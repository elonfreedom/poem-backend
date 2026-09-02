package admin

import (
	"strconv"

	"github.com/go-fuego/fuego"

	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/service/admin"
	"poem-backend/pkg/response"
)

type ImportRecordHandler struct {
	importRecordService *admin.ImportRecordService
}

func NewImportRecordHandler(importRecordService *admin.ImportRecordService) *ImportRecordHandler {
	return &ImportRecordHandler{importRecordService: importRecordService}
}

// List 获取导入记录列表
func (h *ImportRecordHandler) List(c fuego.ContextNoBody) (*response.APIResponse[response.PageData[adminmodel.ImportRecord]], error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	status := c.QueryParam("status")
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	items, total, err := h.importRecordService.List(c.Context(), page, pageSize, status, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return response.PageOK(items, int64(total)), nil
}

// GetByID 获取导入记录详情
func (h *ImportRecordHandler) GetByID(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.ImportRecord], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "导入记录ID必须是数字"}
	}

	record, err := h.importRecordService.GetByID(c.Context(), id)
	if err != nil {
		return nil, err
	}

	return response.OK(*record), nil
}

// Progress 获取导入进度（轻量，供轮询）
func (h *ImportRecordHandler) Progress(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.ImportProgress], error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "invalid id", Detail: "导入记录ID必须是数字"}
	}

	progress, err := h.importRecordService.GetProgress(c.Context(), id)
	if err != nil {
		return nil, err
	}

	return response.OK(*progress), nil
}

// Stats 获取导入统计
func (h *ImportRecordHandler) Stats(c fuego.ContextNoBody) (*response.APIResponse[adminmodel.ImportRecordStatsResponse], error) {
	status := c.QueryParam("status")
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	stats, err := h.importRecordService.GetStats(c.Context(), status, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return response.OK(*stats), nil
}
