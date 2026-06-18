package compat

import (
	"strconv"
	"time"

	"compute-monitor-api/internal/cluster"
	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/page"
	"compute-monitor-api/internal/prometheus"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

var compatFullPage = page.All()

const (
	defaultPrometheusRangeSeconds int64 = 300
	defaultPrometheusStepSeconds  int64 = 60
)

type prometheusRange struct {
	Start int64
	End   int64
	Step  int64
}

// Handler 负责调度器兼容 API。
// 它不再返回手写假数据，而是读取 K8s/Prometheus，把 K8s 数据缓存到数据库后再返回。
type Handler struct {
	k8sFactory  k8s.ClientFactory
	k8sRepo     k8s.Repository
	promFactory prometheus.ClientFactory
	clusterRepo cluster.Repository
}

func NewHandler(k8sFactory k8s.ClientFactory, k8sRepo k8s.Repository, promFactory prometheus.ClientFactory, clusterRepo cluster.Repository) *Handler {
	return &Handler{k8sFactory: k8sFactory, k8sRepo: k8sRepo, promFactory: promFactory, clusterRepo: clusterRepo}
}

// ListClusters 返回调度器可见的集群列表。
func (h *Handler) ListClusters(c *gin.Context) {
	clusters, _, err := h.clusterRepo.List(c.Request.Context(), cluster.ListQuery{Query: page.All()})
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	result := make([]Cluster, 0, len(clusters))
	for _, item := range clusters {
		// 兼容接口的 name 保持为稳定标识，给旧前端或调度器作为集群 key 使用。
		// 真正用于页面展示的名称放在 displayName 中，来源于后台集群配置的 name 字段。
		result = append(result, Cluster{
			ID:          item.ID,
			Name:        item.ID,
			DisplayName: item.Name,
			Status:      item.Status,
			// 当前系统只纳管 Kubernetes 集群，兼容接口直接固定返回 kubernetes。
			Type: "kubernetes",
		})
	}
	response.OK(c, result)
}

// StaticSummary 从 K8s Node capacity 汇总静态容量。
func (h *Handler) StaticSummary(c *gin.Context) {
	clusterID := c.Param("clusterId")
	nodes, err := h.listNodes(c, clusterID)
	if err != nil {
		return
	}

	result := StaticSummary{ClusterID: clusterID, NodeCount: len(nodes)}
	for _, item := range nodes {
		result.CPUTotal += item.CPUCapacity
		result.MemoryTotalBytes += item.MemoryCapacityBytes
		result.GPUTotal += item.GPUCount
	}
	response.OK(c, result)
}

// DynamicSummary 从 Prometheus 查询动态使用率。
func (h *Handler) DynamicSummary(c *gin.Context) {
	clusterID := c.Param("clusterId")
	queryRange := parsePrometheusRange(c)
	cpu, err := h.latest(c, clusterID, `cluster_cpu_usage_rate`, queryRange)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	memory, err := h.latest(c, clusterID, `cluster_memory_usage_rate`, queryRange)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	gpu, err := h.latest(c, clusterID, `DCGM_FI_DEV_GPU_UTIL`, queryRange)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, DynamicSummary{ClusterID: clusterID, CPUUsageRate: cpu, MemoryUsageRate: memory, GPUUsageRate: gpu})
}

// ListNodes 从数据库中的 Node 缓存返回。
func (h *Handler) ListNodes(c *gin.Context) {
	nodes, err := h.listNodes(c, c.Param("clusterId"))
	if err != nil {
		return
	}
	result := make([]Node, 0, len(nodes))
	for _, item := range nodes {
		role := ""
		if len(item.Roles) > 0 {
			role = item.Roles[0]
		}
		result = append(result, Node{Name: item.Name, InternalIP: item.InternalIP, Status: item.Status, Role: role, CPUCapacity: item.CPUCapacity, MemoryCapacityBytes: item.MemoryCapacityBytes, GPUCount: item.GPUCount, OSImage: item.OSImage, ContainerRuntime: item.ContainerRuntime})
	}
	response.OK(c, result)
}

// NodeResourceConsumption 从 Prometheus 查询节点资源使用率。
func (h *Handler) NodeResourceConsumption(c *gin.Context) {
	nodes, _, err := h.k8sRepo.ListNodes(c.Request.Context(), c.Param("clusterId"), compatFullPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	queryRange := parsePrometheusRange(c)
	result := make([]NodeResourceConsumption, 0, len(nodes))
	for _, item := range nodes {
		cpu, _ := h.latest(c, c.Param("clusterId"), `node_cpu_usage_rate{node="`+item.Name+`"}`, queryRange)
		memory, _ := h.latest(c, c.Param("clusterId"), `node_memory_usage_rate{node="`+item.Name+`"}`, queryRange)
		gpu, _ := h.latest(c, c.Param("clusterId"), `DCGM_FI_DEV_GPU_UTIL{node="`+item.Name+`"}`, queryRange)
		result = append(result, NodeResourceConsumption{NodeName: item.Name, CPUUsageRate: cpu, MemoryUsageRate: memory, GPUUsageRate: gpu})
	}
	response.OK(c, result)
}

// ListApps 将 K8s Deployment 映射为调度器侧 app。
func (h *Handler) ListApps(c *gin.Context) {
	clusterID := c.Param("clusterId")
	saved, _, err := h.k8sRepo.ListDeployments(c.Request.Context(), clusterID, c.Query("namespace"), compatFullPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	result := make([]App, 0, len(saved))
	for _, item := range saved {
		result = append(result, App{ID: item.Namespace + "/" + item.Name, Name: item.Name, Namespace: item.Namespace, Kind: "Deployment", Status: "Running", Replicas: item.Replicas, ReadyReplicas: item.ReadyReplicas})
	}
	response.OK(c, result)
}

// ListInstances 将 K8s Pod 映射为调度器侧 instance。
func (h *Handler) ListInstances(c *gin.Context) {
	clusterID := c.Param("clusterId")
	saved, _, err := h.k8sRepo.ListPods(c.Request.Context(), clusterID, c.Query("namespace"), compatFullPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	result := make([]Instance, 0, len(saved))
	for _, item := range saved {
		result = append(result, Instance{ID: item.Namespace + "/" + item.Name, Name: item.Name, Namespace: item.Namespace, NodeName: item.NodeName, Phase: item.Phase, PodIP: item.PodIP, HostIP: item.HostIP, RestartCount: item.RestartCount})
	}
	response.OK(c, result)
}

// ListDeployments 返回数据库中的 Deployment 缓存。
func (h *Handler) ListDeployments(c *gin.Context) {
	clusterID := c.Param("clusterId")
	result, _, err := h.k8sRepo.ListDeployments(c.Request.Context(), clusterID, c.Query("namespace"), compatFullPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, result)
}

// NodeMetrics 从 Prometheus 查询节点指标。
func (h *Handler) NodeMetrics(c *gin.Context) {
	metricType := c.Query("type")
	if metricType == "" {
		metricType = "cpu/usage_rate"
	}
	promClient, err := h.promFactory.ForCluster(c.Request.Context(), c.Param("clusterId"))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	queryRange := parsePrometheusRange(c)
	values, err := promClient.QueryRange(c.Request.Context(), compatMetricQuery(c.Param("nodeName"), metricType), queryRange.Start, queryRange.End, queryRange.Step)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, NodeMetricSeries{NodeName: c.Param("nodeName"), Type: metricType, Unit: metricUnit(metricType), Values: toMetricPoints(values)})
}

// CreateNativeResource 通过 K8s dynamic client 创建 YAML 资源。
func (h *Handler) CreateNativeResource(c *gin.Context) {
	var req NativeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	client, err := h.k8sFactory.ForCluster(c.Request.Context(), c.Param("clusterId"))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	result, err := client.ApplyYAML(c.Request.Context(), req.Namespace, req.YAML)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, NativeCreateResponse{Kind: result.Kind, Namespace: result.Namespace, Name: result.Name})
}

// DeleteDeployment 删除 K8s Deployment。
func (h *Handler) DeleteDeployment(c *gin.Context) {
	namespace := c.Query("namespace")
	client, err := h.k8sFactory.ForCluster(c.Request.Context(), c.Param("clusterId"))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	if err := client.DeleteDeployment(c.Request.Context(), namespace, c.Param("name")); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, NativeDeleteResponse{Name: c.Param("name"), Deleted: true})
}

func (h *Handler) BatchStartApps(c *gin.Context) { h.batchScale(c, 1) }
func (h *Handler) BatchStopApps(c *gin.Context)  { h.batchScale(c, 0) }

func (h *Handler) BatchDeleteApps(c *gin.Context) {
	var req BatchAppsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	for _, app := range req.Apps {
		client, err := h.k8sFactory.ForCluster(c.Request.Context(), c.Param("clusterId"))
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		if err := client.DeleteDeployment(c.Request.Context(), app.Namespace, app.Name); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
	}
	response.OK(c, BatchAppsResponse{ClusterID: c.Param("clusterId"), Action: "delete", Total: len(req.Apps), Items: req.Apps})
}

func (h *Handler) batchScale(c *gin.Context, replicas int) {
	var req BatchAppsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	for _, app := range req.Apps {
		client, err := h.k8sFactory.ForCluster(c.Request.Context(), c.Param("clusterId"))
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		if err := client.ScaleDeployment(c.Request.Context(), app.Namespace, app.Name, replicas); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
	}
	action := "start"
	if replicas == 0 {
		action = "stop"
	}
	response.OK(c, BatchAppsResponse{ClusterID: c.Param("clusterId"), Action: action, Total: len(req.Apps), Items: req.Apps})
}

func (h *Handler) listNodes(c *gin.Context, clusterID string) ([]k8s.Node, error) {
	nodes, _, err := h.k8sRepo.ListNodes(c.Request.Context(), clusterID, compatFullPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
	}
	return nodes, err
}

func (h *Handler) latest(c *gin.Context, clusterID string, query string, queryRange prometheusRange) (float64, error) {
	promClient, err := h.promFactory.ForCluster(c.Request.Context(), clusterID)
	if err != nil {
		return 0, err
	}
	points, err := promClient.QueryRange(c.Request.Context(), query, queryRange.Start, queryRange.End, queryRange.Step)
	if err != nil || len(points) == 0 {
		return 0, err
	}
	return points[len(points)-1].Value, nil
}

func parsePrometheusRange(c *gin.Context) prometheusRange {
	end, _ := strconv.ParseInt(c.Query("end"), 10, 64)
	if end <= 0 {
		end = time.Now().Unix()
	}

	start, _ := strconv.ParseInt(c.Query("start"), 10, 64)
	if start <= 0 {
		start = end - defaultPrometheusRangeSeconds
	}

	step, _ := strconv.ParseInt(c.Query("step"), 10, 64)
	if step <= 0 {
		step = defaultPrometheusStepSeconds
	}

	return prometheusRange{
		Start: start,
		End:   end,
		Step:  step,
	}
}

func compatMetricQuery(nodeName string, metricType string) string {
	switch metricType {
	case "memory/usage":
		return `node_memory_usage_bytes{node="` + nodeName + `"}`
	case "gpu/node_ai_core_usage_rate":
		return `DCGM_FI_DEV_GPU_UTIL{node="` + nodeName + `"}`
	default:
		return `node_cpu_usage_rate{node="` + nodeName + `"}`
	}
}

func toMetricPoints(points []prometheus.Point) []MetricPoint {
	result := make([]MetricPoint, 0, len(points))
	for _, item := range points {
		result = append(result, MetricPoint{Timestamp: item.Timestamp, Value: item.Value})
	}
	return result
}

func metricUnit(metricType string) string {
	switch metricType {
	case "network/rx_rate", "network/tx_rate", "disk/node_readio", "disk/node_writeio":
		return "B/s"
	case "gpu/node_temperature":
		return "celsius"
	case "gpu/node_power":
		return "watt"
	default:
		return "%"
	}
}
