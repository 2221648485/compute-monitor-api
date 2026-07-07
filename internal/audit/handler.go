package audit

import (
	"compute-monitor-api/internal/page"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 负责审计日志 HTTP 接口。
type Handler struct {
	service *Service
}

// NewHandler 创建审计日志 handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// List 分页查询审计日志列表。
func (h *Handler) List(c *gin.Context) {
	var query page.Query
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.List(c.Request.Context(), c.Query("clusterId"), c.Query("action"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}
