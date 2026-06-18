package gpu

import (
	"compute-monitor-api/internal/page"
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

func (h *Handler) List(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.List(c.Request.Context(), c.Param("clusterId"), "", query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ListByNode(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.List(c.Request.Context(), c.Param("clusterId"), c.Param("nodeName"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Summary(c *gin.Context) {
	result, err := h.service.Summary(c.Request.Context(), c.Param("clusterId"))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Top(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.Top(c.Request.Context(), c.Param("clusterId"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Metrics(c *gin.Context) {
	start, _ := strconv.ParseInt(c.Query("start"), 10, 64)
	end, _ := strconv.ParseInt(c.Query("end"), 10, 64)
	step, _ := strconv.ParseInt(c.Query("step"), 10, 64)
	points, err := h.service.Metric(c.Request.Context(), c.Param("clusterId"), c.Param("nodeName"), start, end, step)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, MetricResponse{
		NodeName: c.Param("nodeName"),
		Metric:   "DCGM_FI_DEV_GPU_UTIL",
		Unit:     "%",
		Values:   points,
	})
}

func bindPage(c *gin.Context) (page.Query, bool) {
	var query page.Query
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, err.Error())
		return page.Query{}, false
	}
	return query, true
}
