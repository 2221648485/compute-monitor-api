package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client 封装 Kubernetes client-go，业务模块不要直接散落创建 clientset。
type Client struct {
	clientset kubernetes.Interface
}

type NodeInfo struct {
	Name       string
	InternalIP string
	Status     string
}

type NamespaceInfo struct {
	Name string
}

func NewClient(clientset kubernetes.Interface) *Client {
	return &Client{clientset: clientset}
}

// NewClientFromKubeconfig 用于本地开发或平台服务连接外部集群。
func NewClientFromKubeconfig(kubeconfigPath string) (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build kubeconfig: %w", err)
	}

	return NewClientFromRESTConfig(config)
}

// NewInClusterClient 用于 compute-monitor-api 部署在 K8s 集群内部时，通过 ServiceAccount 访问 API Server。
func NewInClusterClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("build in-cluster config: %w", err)
	}

	return NewClientFromRESTConfig(config)
}

func NewClientFromRESTConfig(config *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes clientset: %w", err)
	}

	return NewClient(clientset), nil
}

func (c *Client) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}

	result := make([]NodeInfo, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		result = append(result, NodeInfo{
			Name:       node.Name,
			InternalIP: findNodeInternalIP(node),
			Status:     nodeReadyStatus(node),
		})
	}

	return result, nil
}

func (c *Client) ListNamespaces(ctx context.Context) ([]NamespaceInfo, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	result := make([]NamespaceInfo, 0, len(namespaces.Items))
	for _, namespace := range namespaces.Items {
		result = append(result, NamespaceInfo{Name: namespace.Name})
	}

	return result, nil
}

func findNodeInternalIP(node corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}

	return ""
}

func nodeReadyStatus(node corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type != corev1.NodeReady {
			continue
		}
		if condition.Status == corev1.ConditionTrue {
			return "Ready"
		}

		return "NotReady"
	}

	return "Unknown"
}
