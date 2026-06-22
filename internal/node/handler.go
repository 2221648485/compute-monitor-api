package node

import (
	"compute-monitor-api/internal/page"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListNamespaces(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.ListNamespaces(c.Request.Context(), c.Param("clusterId"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ListNodes(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.ListNodes(c.Request.Context(), c.Param("clusterId"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) GetNode(c *gin.Context) {
	result, err := h.service.GetNode(c.Request.Context(), c.Param("clusterId"), c.Param("nodeName"))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ListNodePods(c *gin.Context) {
	query, ok := bindPage(c)
	if !ok {
		return
	}
	result, err := h.service.ListPods(c.Request.Context(), c.Param("clusterId"), c.Query("namespace"), query)
	if err != nil {
		response.InternalServerError(c, err.Error())
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
