package cluster

import "compute-monitor-api/internal/page"

// ListQuery 是集群列表查询参数。
type ListQuery struct {
	page.Query
	Keyword string `form:"keyword"`
	Status  string `form:"status"`
}

// CreateClusterRequest 是创建集群的请求体。
type CreateClusterRequest struct {
	ID             string `json:"id" binding:"required"`
	Name           string `json:"name" binding:"required"`
	KubeconfigPath string `json:"kubeconfig_path"`
	PrometheusURL  string `json:"prometheus_url"`
	Description    string `json:"description"`
	Status         string `json:"status"`
}

// UpdateClusterRequest 是修改集群的请求体。
type UpdateClusterRequest struct {
	Name           string `json:"name" binding:"required"`
	KubeconfigPath string `json:"kubeconfig_path"`
	PrometheusURL  string `json:"prometheus_url"`
	Description    string `json:"description"`
	Status         string `json:"status"`
}
