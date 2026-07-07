package server

import (
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 放置平台级辅助接口，例如兼容接口健康状态。
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// CompatEndpointStatus 返回兼容接口实现状态。
func (h *Handler) CompatEndpointStatus(c *gin.Context) {
	response.OK(c, []CompatEndpointStatusResponse{
		{Method: "GET", Path: "/api/v2/clusters/exclusive/list", Priority: "P0", Status: "implemented"},
		{Method: "POST", Path: "/api/v2/clusters/{clusterId}/native", Priority: "P1", Status: "implemented"},
	})
}
