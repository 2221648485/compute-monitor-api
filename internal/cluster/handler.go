package cluster

import (
	"errors"
	"log"
	"net/http"

	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 负责集群管理 HTTP 接口。
type Handler struct {
	service *Service
}

// NewHandler 创建集群 handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create 创建集群。
//
// @Summary 创建集群配置
// @Tags 集群管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateClusterRequest true "创建集群请求"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/clusters [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		writeClusterError(c, err)
		return
	}
	response.Created(c, result)
}

// List 查询集群列表。
//
// @Summary 分页查询集群列表
// @Tags 集群管理
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "集群 ID、名称或描述关键字"
// @Param status query string false "集群状态：Running/NotReady/Disabled"
// @Param page query int false "页码，默认 1"
// @Param size query int false "每页数量，默认 20，最大 100"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/clusters [get]
func (h *Handler) List(c *gin.Context) {
	var query ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		writeClusterError(c, err)
		return
	}
	response.OK(c, result)
}

// Get 集群详情。
//
// @Summary 查询集群详情
// @Tags 集群管理
// @Produce json
// @Security BearerAuth
// @Param clusterId path string true "集群 ID"
// @Success 200 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 404 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/clusters/{clusterId} [get]
func (h *Handler) Get(c *gin.Context) {
	result, err := h.service.Get(c.Request.Context(), c.Param("clusterId"))
	if err != nil {
		writeClusterError(c, err)
		return
	}
	response.OK(c, result)
}

// Update 修改集群配置。
//
// @Summary 修改集群配置
// @Tags 集群管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param clusterId path string true "集群 ID"
// @Param request body UpdateClusterRequest true "修改集群请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 404 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/clusters/{clusterId} [put]
func (h *Handler) Update(c *gin.Context) {
	var req UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.Update(c.Request.Context(), c.Param("clusterId"), req)
	if err != nil {
		writeClusterError(c, err)
		return
	}
	response.OK(c, result)
}

// Test 测试集群连接。
//
// @Summary 测试集群连接
// @Tags 集群管理
// @Produce json
// @Security BearerAuth
// @Param clusterId path string true "集群 ID"
// @Success 200 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 404 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/clusters/{clusterId}/test [post]
func (h *Handler) Test(c *gin.Context) {
	result, err := h.service.TestConnection(c.Request.Context(), c.Param("clusterId"))
	if err != nil {
		writeClusterError(c, err)
		return
	}
	response.OK(c, result)
}

// Delete 删除集群配置。
//
// @Summary 删除集群配置
// @Tags 集群管理
// @Produce json
// @Security BearerAuth
// @Param clusterId path string true "集群 ID"
// @Success 200 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 404 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/clusters/{clusterId} [delete]
func (h *Handler) Delete(c *gin.Context) {
	clusterID := c.Param("clusterId")
	if err := h.service.Delete(c.Request.Context(), clusterID); err != nil {
		writeClusterError(c, err)
		return
	}
	response.OK(c, DeleteClusterResponse{ClusterID: clusterID, Deleted: true})
}

func writeClusterError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrClusterNotFound):
		c.JSON(http.StatusNotFound, response.Body{Code: http.StatusNotFound, Message: ErrorMessage(err)})
	case IsClusterError(err):
		response.BadRequest(c, ErrorMessage(err))
	default:
		log.Printf("cluster request failed: error=%v", err)
		response.InternalServerError(c, "failed to process cluster request")
	}
}
