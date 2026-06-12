package k8s

import "encoding/json"

// Node 是 Node 的接口返回模型。
type Node struct {
	Name                string   `json:"name"`
	InternalIP          string   `json:"internalIP"`
	Status              string   `json:"status"`
	Roles               []string `json:"roles"`
	CPUCapacity         int      `json:"cpuCapacity"`
	MemoryCapacityBytes int64    `json:"memoryCapacityBytes"`
	GPUCount            int      `json:"gpuCount"`
	OSImage             string   `json:"osImage"`
	KernelVersion       string   `json:"kernelVersion"`
	ContainerRuntime    string   `json:"containerRuntime"`
}

// Pod 是 Pod 的接口返回模型。
type Pod struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	NodeName     string `json:"nodeName"`
	Phase        string `json:"phase"`
	PodIP        string `json:"podIP"`
	HostIP       string `json:"hostIP"`
	RestartCount int    `json:"restartCount"`
}

// Deployment 是 Deployment 的接口返回模型。
type Deployment struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int               `json:"replicas"`
	ReadyReplicas     int               `json:"readyReplicas"`
	AvailableReplicas int               `json:"availableReplicas"`
	Labels            map[string]string `json:"labels"`
}

// Service 是 Kubernetes Service 的接口返回模型。
type Service struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	ClusterIP string `json:"clusterIP"`
}

func parseStringSlice(raw string) []string {
	var result []string
	_ = json.Unmarshal([]byte(raw), &result)
	return result
}

func parseStringMap(raw string) map[string]string {
	result := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &result)
	return result
}
