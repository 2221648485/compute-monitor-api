package compat

import (
	"time"

	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ListClusters 返回调度器可见的集群列表。第一阶段只提供 demo 集群。
func (h *Handler) ListClusters(c *gin.Context) {
	response.OK(c, demoClusters)
}

// StaticSummary 返回静态容量信息，后续会从 K8s Node capacity 汇总。
func (h *Handler) StaticSummary(c *gin.Context) {
	response.OK(c, StaticSummary{
		ClusterID:        c.Param("clusterId"),
		NodeCount:        len(demoNodes),
		CPUTotal:         64,
		MemoryTotalBytes: 274877906944,
		GPUTotal:         8,
		PodCapacity:      220,
	})
}

// DynamicSummary 返回动态使用率，后续会从 Prometheus 和 K8s Pod 状态计算。
func (h *Handler) DynamicSummary(c *gin.Context) {
	response.OK(c, DynamicSummary{
		ClusterID:       c.Param("clusterId"),
		CPUUsageRate:    23.5,
		MemoryUsageRate: 61.2,
		GPUUsageRate:    35.8,
		PodRunningCount: 18,
		PodPendingCount: 1,
	})
}

// ListNodes 当前返回 mock 节点，阶段 2 会替换为 client-go 查询结果。
func (h *Handler) ListNodes(c *gin.Context) {
	response.OK(c, demoNodes)
}

// NodeResourceConsumption 当前返回 mock 使用率，阶段 3 会替换为 Prometheus 指标。
func (h *Handler) NodeResourceConsumption(c *gin.Context) {
	response.OK(c, []NodeResourceConsumption{
		{
			NodeName:        "demo-node-01",
			CPUUsageRate:    20.1,
			MemoryUsageRate: 58.4,
			DiskUsageRate:   42.0,
			GPUUsageRate:    36.5,
		},
		{
			NodeName:        "demo-node-02",
			CPUUsageRate:    31.7,
			MemoryUsageRate: 63.2,
			DiskUsageRate:   39.8,
			GPUUsageRate:    28.4,
		},
	})
}

// ListApps 将 Deployment 映射成调度器侧的 app。
func (h *Handler) ListApps(c *gin.Context) {
	response.OK(c, demoApps)
}

// ListInstances 将 Pod 映射成调度器侧的 instance。
func (h *Handler) ListInstances(c *gin.Context) {
	response.OK(c, demoInstances)
}

// ListDeployments 返回原生 Deployment 兼容列表。
func (h *Handler) ListDeployments(c *gin.Context) {
	response.OK(c, demoDeployments)
}

// NodeMetrics 根据 type 返回一段 mock 指标序列，方便先打通调度器指标请求。
func (h *Handler) NodeMetrics(c *gin.Context) {
	metricType := c.Query("type")
	if metricType == "" {
		metricType = "cpu/usage_rate"
	}

	now := time.Now().Unix()
	response.OK(c, NodeMetricSeries{
		NodeName: c.Param("nodeName"),
		Type:     metricType,
		Unit:     metricUnit(metricType),
		Values: []MetricPoint{
			{Timestamp: now - 120, Value: 21.3},
			{Timestamp: now - 60, Value: 22.1},
			{Timestamp: now, Value: 23.4},
		},
	})
}

func metricUnit(metricType string) string {
	// 不同指标单位不同，先在兼容层做最小映射，后续可迁到 metrics 模块统一管理。
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
