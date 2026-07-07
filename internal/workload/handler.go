package workload

import (
	"strconv"

	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/page"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 负责工作负载 HTTP 接口。
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListPods(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.ListPods(c.Request.Context(), c.Param("clusterId"), c.Query("namespace"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) GetPod(c *gin.Context) {
	result, err := h.service.GetPod(c.Request.Context(), c.Param("clusterId"), c.Param("namespace"), c.Param("name"))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) PodLogs(c *gin.Context) {
	opts, ok := bindLogOptions(c)
	if !ok {
		return
	}
	result, err := h.service.PodLogs(c.Request.Context(), c.Param("clusterId"), c.Param("namespace"), c.Param("name"), opts)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ListDeployments(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.ListDeployments(c.Request.Context(), c.Param("clusterId"), c.Query("namespace"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ListServices(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.ListServices(c.Request.Context(), c.Param("clusterId"), c.Query("namespace"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ApplyYAML(c *gin.Context) {
	var req ApplyYAMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.ApplyYAML(c.Request.Context(), c.Param("clusterId"), req.Namespace, req.YAML)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) DeleteDeployment(c *gin.Context) {
	if err := h.service.DeleteDeployment(c.Request.Context(), c.Param("clusterId"), c.Param("namespace"), c.Param("name")); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, DeleteDeploymentResponse{Namespace: c.Param("namespace"), Name: c.Param("name"), Deleted: true})
}

func (h *Handler) ScaleDeployment(c *gin.Context) {
	var req ScaleDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.ScaleDeployment(c.Request.Context(), c.Param("clusterId"), c.Param("namespace"), c.Param("name"), req.Replicas); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, ScaleDeploymentResponse{Namespace: c.Param("namespace"), Name: c.Param("name"), Replicas: req.Replicas})
}

func bindPage(c *gin.Context) (page.Query, bool) {
	var query page.Query
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return page.Query{}, false
	}
	return query, true
}

func bindLogOptions(c *gin.Context) (k8s.LogOptions, bool) {
	tailLines, ok := parseInt64Query(c, "tailLines", 500)
	if !ok {
		return k8s.LogOptions{}, false
	}
	limitBytes, ok := parseInt64Query(c, "limitBytes", 1024*1024)
	if !ok {
		return k8s.LogOptions{}, false
	}
	return k8s.LogOptions{
		Container:  c.Query("container"),
		TailLines:  tailLines,
		LimitBytes: limitBytes,
	}, true
}

func parseInt64Query(c *gin.Context, name string, defaultValue int64) (int64, bool) {
	rawValue := c.Query(name)
	if rawValue == "" {
		return defaultValue, true
	}
	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil || value <= 0 {
		response.BadRequest(c, name+" must be a positive integer")
		return 0, false
	}
	return value, true
}
