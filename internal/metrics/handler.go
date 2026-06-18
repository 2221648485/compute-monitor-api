package metrics

import (
	"compute-monitor-api/internal/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CPU(c *gin.Context) {
	h.write(c, "cpu_usage_rate", "%")
}
func (h *Handler) Memory(c *gin.Context) {
	h.write(c, "memory_usage_rate", "%")
}
func (h *Handler) Disk(c *gin.Context) {
	h.write(c, "disk_usage_rate", "%")
}
func (h *Handler) Network(c *gin.Context) {
	h.write(c, "network_usage_rate", "%")
}

func (h *Handler) write(c *gin.Context, metric string, unit string) {
	start, _ := strconv.ParseInt(c.Query("start"), 10, 64)
	end, _ := strconv.ParseInt(c.Query("end"), 10, 64)
	step, _ := strconv.ParseInt(c.Query("step"), 10, 64)
	result, err := h.service.NodeMetric(c.Request.Context(), c.Param("clusterId"), c.Param("nodeName"), metric, unit, start, end, step)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}
