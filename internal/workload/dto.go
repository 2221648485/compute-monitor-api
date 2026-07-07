package workload

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

// DeleteDeploymentResponse 是删除 Deployment 的接口响应。
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
