package k8s

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Namespace 是 Namespace 的接口返回模型。
type Namespace struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

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

// ApplyResult 是 YAML 创建资源后的结果。
type ApplyResult struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func toNamespace(item corev1.Namespace) Namespace {
	return Namespace{Name: item.Name, Status: string(item.Status.Phase)}
}

func toNode(item corev1.Node) Node {
	return Node{
		Name:                item.Name,
		InternalIP:          nodeInternalIP(item),
		Status:              nodeReadyStatus(item),
		Roles:               nodeRoles(item),
		CPUCapacity:         int(item.Status.Capacity.Cpu().Value()),
		MemoryCapacityBytes: item.Status.Capacity.Memory().Value(),
		GPUCount:            gpuCount(item),
		OSImage:             item.Status.NodeInfo.OSImage,
		KernelVersion:       item.Status.NodeInfo.KernelVersion,
		ContainerRuntime:    item.Status.NodeInfo.ContainerRuntimeVersion,
	}
}

func toPod(item corev1.Pod) Pod {
	restarts := 0
	for _, status := range item.Status.ContainerStatuses {
		restarts += int(status.RestartCount)
	}
	return Pod{
		Name:         item.Name,
		Namespace:    item.Namespace,
		NodeName:     item.Spec.NodeName,
		Phase:        string(item.Status.Phase),
		PodIP:        item.Status.PodIP,
		HostIP:       item.Status.HostIP,
		RestartCount: restarts,
	}
}

func toDeployment(item appsv1.Deployment) Deployment {
	replicas := 0
	if item.Spec.Replicas != nil {
		replicas = int(*item.Spec.Replicas)
	}
	return Deployment{
		Name:              item.Name,
		Namespace:         item.Namespace,
		Replicas:          replicas,
		ReadyReplicas:     int(item.Status.ReadyReplicas),
		AvailableReplicas: int(item.Status.AvailableReplicas),
		Labels:            item.Labels,
	}
}

func toService(item corev1.Service) Service {
	return Service{
		Name:      item.Name,
		Namespace: item.Namespace,
		Type:      string(item.Spec.Type),
		ClusterIP: item.Spec.ClusterIP,
	}
}

func gpuCount(node corev1.Node) int {
	quantity, ok := node.Status.Allocatable[corev1.ResourceName("nvidia.com/gpu")]
	if !ok {
		return 0
	}
	return int(quantity.Value())
}

func nodeInternalIP(node corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}
	return ""
}

func nodeReadyStatus(node corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return "Ready"
		}
	}
	return "NotReady"
}

func nodeRoles(node corev1.Node) []string {
	roles := make([]string, 0)
	for key := range node.Labels {
		const prefix = "node-role.kubernetes.io/"
		if strings.HasPrefix(key, prefix) {
			roles = append(roles, strings.TrimPrefix(key, prefix))
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	return roles
}
