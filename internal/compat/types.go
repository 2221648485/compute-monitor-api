package compat

type Cluster struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Status      string `json:"status"`
	Type        string `json:"type"`
}

type StaticSummary struct {
	ClusterID        string `json:"clusterId"`
	NodeCount        int    `json:"nodeCount"`
	CPUTotal         int    `json:"cpuTotal"`
	MemoryTotalBytes int64  `json:"memoryTotalBytes"`
	GPUTotal         int    `json:"gpuTotal"`
	PodCapacity      int    `json:"podCapacity"`
}

type DynamicSummary struct {
	ClusterID       string  `json:"clusterId"`
	CPUUsageRate    float64 `json:"cpuUsageRate"`
	MemoryUsageRate float64 `json:"memoryUsageRate"`
	GPUUsageRate    float64 `json:"gpuUsageRate"`
	PodRunningCount int     `json:"podRunningCount"`
	PodPendingCount int     `json:"podPendingCount"`
}

type Node struct {
	Name                string `json:"name"`
	InternalIP          string `json:"internalIP"`
	Status              string `json:"status"`
	Role                string `json:"role"`
	CPUCapacity         int    `json:"cpuCapacity"`
	MemoryCapacityBytes int64  `json:"memoryCapacityBytes"`
	GPUCount            int    `json:"gpuCount"`
	OSImage             string `json:"osImage"`
	ContainerRuntime    string `json:"containerRuntime"`
}

type NodeResourceConsumption struct {
	NodeName        string  `json:"nodeName"`
	CPUUsageRate    float64 `json:"cpuUsageRate"`
	MemoryUsageRate float64 `json:"memoryUsageRate"`
	DiskUsageRate   float64 `json:"diskUsageRate"`
	GPUUsageRate    float64 `json:"gpuUsageRate"`
}

type App struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Kind          string `json:"kind"`
	Status        string `json:"status"`
	Replicas      int    `json:"replicas"`
	ReadyReplicas int    `json:"readyReplicas"`
	CreatedAt     string `json:"createdAt"`
}

type Instance struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	AppName      string `json:"appName"`
	NodeName     string `json:"nodeName"`
	Phase        string `json:"phase"`
	PodIP        string `json:"podIP"`
	HostIP       string `json:"hostIP"`
	RestartCount int    `json:"restartCount"`
}

type Deployment struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int               `json:"replicas"`
	ReadyReplicas     int               `json:"readyReplicas"`
	AvailableReplicas int               `json:"availableReplicas"`
	Labels            map[string]string `json:"labels"`
}

type MetricPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type NodeMetricSeries struct {
	NodeName string        `json:"nodeName"`
	Type     string        `json:"type"`
	Unit     string        `json:"unit"`
	Values   []MetricPoint `json:"values"`
}
