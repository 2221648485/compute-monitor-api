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

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create 创建集群配置。
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

// CreateByUpload 通过上传 kubeconfig 文件创建集群。
//
// @Summary 上传 kubeconfig 创建集群
// @Tags 集群管理
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id formData string true "集群 ID"
// @Param name formData string true "集群名称"
// @Param kubeconfig formData file true "kubeconfig 文件"
// @Param api_server formData string false "Kubernetes API Server，仅作为元数据保存，不覆盖 kubeconfig"
// @Param prometheus_url formData string false "Prometheus 地址"
// @Param description formData string false "集群描述"
// @Param status formData string false "集群状态"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/clusters/upload [post]
func (h *Handler) CreateByUpload(c *gin.Context) {
	file, err := c.FormFile("kubeconfig")
	if err != nil {
		response.BadRequest(c, ErrorMessage(ErrKubeconfigRequired))
		return
	}
	req := CreateClusterRequest{
		ID:            c.PostForm("id"),
		Name:          c.PostForm("name"),
		APIServer:     c.PostForm("api_server"),
		PrometheusURL: c.PostForm("prometheus_url"),
		Description:   c.PostForm("description"),
		Status:        c.PostForm("status"),
	}
	result, err := h.service.CreateFromUpload(c.Request.Context(), req, file)
	if err != nil {
		writeClusterError(c, err)
		return
	}
	response.Created(c, result)
}

// List 分页查询集群列表。
//
// @Summary 分页查询集群列表
// @Tags 集群管理
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "集群 ID、名称或描述关键字"
// @Param status query string false "集群状态：Running/NotReady/Disabled"
// @Param page query int false "页码"
// @Param size query int false "每页数量"
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

// Get 查询集群详情。
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
