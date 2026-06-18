package gpu

import "compute-monitor-api/internal/prometheus"

// GPUResponse 是 GPU 列表接口返回的单个 GPU 视图对象。
// 当前系统没有独立 GPU 表，GPU 数据由 Node 缓存中的 GPUCount 展开得到。
type GPUResponse struct {
	NodeName        string  `json:"nodeName"`
	GPUIndex        int     `json:"gpuIndex"`
	GPUUUID         string  `json:"gpuUUID"`
	Utilization     float64 `json:"utilization"`
	MemoryUsageRate float64 `json:"memoryUsageRate"`
	Temperature     float64 `json:"temperature"`
}

// SummaryResponse 是集群 GPU 汇总接口响应。
type SummaryResponse struct {
	ClusterID string `json:"clusterId"`
	Total     int64  `json:"total"`
	Source    string `json:"source"`
}

// MetricResponse 是节点 GPU 指标接口响应。
type MetricResponse struct {
	NodeName string             `json:"nodeName"`
	Metric   string             `json:"metric"`
	Unit     string             `json:"unit"`
	Values   []prometheus.Point `json:"values"`
}
