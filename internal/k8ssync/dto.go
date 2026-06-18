package k8ssync

// Result 是全量同步接口响应。
type Result struct {
	ClusterID   string `json:"clusterId"`
	Namespaces  int    `json:"namespaces"`
	Nodes       int    `json:"nodes"`
	Pods        int    `json:"pods"`
	Deployments int    `json:"deployments"`
	Services    int    `json:"services"`
}

// NamespacesResponse 是 Namespace 同步接口响应。
type NamespacesResponse struct {
	ClusterID  string `json:"clusterId"`
	Namespaces int    `json:"namespaces"`
}

// NodesResponse 是 Node 同步接口响应。
type NodesResponse struct {
	ClusterID string `json:"clusterId"`
	Nodes     int    `json:"nodes"`
}

// PodsResponse 是 Pod 同步接口响应。
type PodsResponse struct {
	ClusterID string `json:"clusterId"`
	Pods      int    `json:"pods"`
}

// DeploymentsResponse 是 Deployment 同步接口响应。
type DeploymentsResponse struct {
	ClusterID   string `json:"clusterId"`
	Deployments int    `json:"deployments"`
}

// ServicesResponse 是 Service 同步接口响应。
type ServicesResponse struct {
	ClusterID string `json:"clusterId"`
	Services  int    `json:"services"`
}
