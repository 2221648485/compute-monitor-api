package k8ssync

import (
	"errors"
	"log"
	"net/http"

	"compute-monitor-api/internal/cluster"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 负责 Kubernetes 资源同步接口。
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SyncAll(c *gin.Context) {
	result, err := h.service.SyncAll(c.Request.Context(), c.Param("clusterId"), c.Query("namespace"))
	if err != nil {
		writeK8sSyncError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) SyncNamespaces(c *gin.Context) {
	count, err := h.service.SyncNamespaces(c.Request.Context(), c.Param("clusterId"))
	if err != nil {
		writeK8sSyncError(c, err)
		return
	}
	response.OK(c, NamespacesResponse{ClusterID: c.Param("clusterId"), Namespaces: count})
}

func (h *Handler) SyncNodes(c *gin.Context) {
	count, err := h.service.SyncNodes(c.Request.Context(), c.Param("clusterId"))
	if err != nil {
		writeK8sSyncError(c, err)
		return
	}
	response.OK(c, NodesResponse{ClusterID: c.Param("clusterId"), Nodes: count})
}

func (h *Handler) SyncPods(c *gin.Context) {
	count, err := h.service.SyncPods(c.Request.Context(), c.Param("clusterId"), c.Query("namespace"))
	if err != nil {
		writeK8sSyncError(c, err)
		return
	}
	response.OK(c, PodsResponse{ClusterID: c.Param("clusterId"), Pods: count})
}

func (h *Handler) SyncDeployments(c *gin.Context) {
	count, err := h.service.SyncDeployments(c.Request.Context(), c.Param("clusterId"), c.Query("namespace"))
	if err != nil {
		writeK8sSyncError(c, err)
		return
	}
	response.OK(c, DeploymentsResponse{ClusterID: c.Param("clusterId"), Deployments: count})
}

func (h *Handler) SyncServices(c *gin.Context) {
	count, err := h.service.SyncServices(c.Request.Context(), c.Param("clusterId"), c.Query("namespace"))
	if err != nil {
		writeK8sSyncError(c, err)
		return
	}
	response.OK(c, ServicesResponse{ClusterID: c.Param("clusterId"), Services: count})
}

func writeK8sSyncError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, cluster.ErrClusterNotFound):
		c.JSON(http.StatusNotFound, response.Body{Code: http.StatusNotFound, Message: cluster.ErrorMessage(err)})
	case cluster.IsClusterError(err):
		response.BadRequest(c, cluster.ErrorMessage(err))
	default:
		log.Printf("k8ssync request failed: error=%v", err)
		response.InternalServerError(c, "failed to sync kubernetes resources")
	}
}
