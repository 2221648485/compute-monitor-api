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
	AccessMode     string `json:"access_mode"`
	APIServer      string `json:"api_server"`
	KubeconfigPath string `json:"kubeconfig_path"`
	CACert         string `json:"ca_cert"`
	BearerToken    string `json:"bearer_token"`
	ClientCert     string `json:"client_cert"`
	ClientKey      string `json:"client_key"`
	Insecure       bool   `json:"insecure"`
	PrometheusURL  string `json:"prometheus_url"`
	Description    string `json:"description"`
	Status         string `json:"status"`
}

// UpdateClusterRequest 是修改集群的请求体。
type UpdateClusterRequest struct {
	Name           string `json:"name" binding:"required"`
	AccessMode     string `json:"access_mode"`
	APIServer      string `json:"api_server"`
	KubeconfigPath string `json:"kubeconfig_path"`
	CACert         string `json:"ca_cert"`
	BearerToken    string `json:"bearer_token"`
	ClientCert     string `json:"client_cert"`
	ClientKey      string `json:"client_key"`
	Insecure       bool   `json:"insecure"`
	PrometheusURL  string `json:"prometheus_url"`
	Description    string `json:"description"`
	Status         string `json:"status"`
}

// TestConnectionResponse 是集群连接测试接口响应。
type TestConnectionResponse struct {
	ClusterID      string `json:"clusterId"`
	Connected      bool   `json:"connected"`
	NamespaceCount int    `json:"namespaceCount"`
}

// DeleteClusterResponse 是删除集群配置后的响应。
type DeleteClusterResponse struct {
	ClusterID string `json:"clusterId"`
	Deleted   bool   `json:"deleted"`
}
