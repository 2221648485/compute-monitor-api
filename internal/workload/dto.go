package workload

import "compute-monitor-api/internal/page"

// ListWorkloadQuery 是 Pod、Deployment、Service 等工作负载列表的通用查询参数。
// keyword 用于模糊匹配名称、命名空间和部分关键字段，避免集群资源过多时只能全量翻页。
type ListWorkloadQuery struct {
	page.Query
	Keyword string `form:"keyword"`
}

// ApplyYAMLRequest 是通过 YAML 创建或更新 Kubernetes 原生资源的请求体。
type ApplyYAMLRequest struct {
	Namespace string `json:"namespace"`
	YAML      string `json:"yaml"`
}

// PodLogsResponse 是 Pod 日志查询接口响应。
type PodLogsResponse struct {
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	Container  string `json:"container"`
	TailLines  int64  `json:"tailLines"`
	LimitBytes int64  `json:"limitBytes"`
	Logs       string `json:"logs"`
}

// DeleteDeploymentResponse 是删除 Deployment 后的接口响应。
type DeleteDeploymentResponse struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Deleted   bool   `json:"deleted"`
}

// ScaleDeploymentRequest 是 Deployment 扩缩容请求体。
type ScaleDeploymentRequest struct {
	Replicas int `json:"replicas"`
}

// ScaleDeploymentResponse 是 Deployment 扩缩容接口响应。
type ScaleDeploymentResponse struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Replicas  int    `json:"replicas"`
}
