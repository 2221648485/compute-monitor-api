package alert

import (
	"compute-monitor-api/internal/page"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 负责告警规则和告警事件 HTTP 接口。
type Handler struct {
	service *Service
}

// NewHandler 创建告警 handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateRule 创建告警规则。
func (h *Handler) CreateRule(c *gin.Context) {
	var req RuleRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.CreateRule(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Created(c, result)
}

// ListRules 分页查询告警规则列表。
func (h *Handler) ListRules(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.ListRules(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

// UpdateRule 更新告警规则。
func (h *Handler) UpdateRule(c *gin.Context) {
	var req RuleRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.UpdateRule(c.Request.Context(), c.Param("ruleId"), req); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, RuleOperationResponse{ID: c.Param("ruleId"), Updated: true})
}

// DeleteRule 删除告警规则。
func (h *Handler) DeleteRule(c *gin.Context) {
	if err := h.service.DeleteRule(c.Request.Context(), c.Param("ruleId")); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, RuleOperationResponse{ID: c.Param("ruleId"), Deleted: true})
}

// ListEvents 分页查询告警事件列表。
func (h *Handler) ListEvents(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.ListEvents(c.Request.Context(), c.Query("clusterId"), c.Query("status"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func bindPage(c *gin.Context) (page.Query, bool) {
	var query page.Query
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return page.Query{}, false
	}
	return query, true
}
