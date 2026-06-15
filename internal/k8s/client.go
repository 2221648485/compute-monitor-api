package k8s

import (
	"compute-monitor-api/internal/config"
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type Options struct {
	Mode           string
	APIServer      string
	KubeconfigPath string
}

type LogOptions struct {
	Container  string
	TailLines  int64
	LimitBytes int64
}

// Client 是访问单个 Kubernetes 集群的统一入口。
type Client interface {
	ListNamespaces(ctx context.Context) ([]Namespace, error)
	ListNodes(ctx context.Context) ([]Node, error)
	ListPods(ctx context.Context, namespace string) ([]Pod, error)
	ListDeployments(ctx context.Context, namespace string) ([]Deployment, error)
	ListServices(ctx context.Context, namespace string) ([]Service, error)
	PodLogs(ctx context.Context, namespace string, name string, opts LogOptions) (string, error)
	ApplyYAML(ctx context.Context, namespace string, yamlContent string) (ApplyResult, error)
	DeleteDeployment(ctx context.Context, namespace string, name string) error
	ScaleDeployment(ctx context.Context, namespace string, name string, replicas int) error
}

type ClientFactory interface {
	ForCluster(ctx context.Context, clusterID string) (Client, error)
}

type ClientGoClient struct {
	clientset kubernetes.Interface
	dynamic   dynamic.Interface
	mapper    meta.RESTMapper
}

func NewClient(opts Options) (Client, error) {
	restConfig, err := buildRESTConfig(opts)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes clientset: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes dynamic client: %w", err)
	}
	discoveryClient := memory.NewMemCacheClient(clientset.Discovery())
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	return &ClientGoClient{
		clientset: clientset,
		dynamic:   dynamicClient,
		mapper:    mapper,
	}, nil
}

func NewClientFromConfig(cfg config.K8sConfig) (Client, error) {
	return NewClient(Options{
		Mode:           cfg.Mode,
		APIServer:      cfg.ApiServer,
		KubeconfigPath: cfg.KubeconfigPath,
	})
}

// ListNamespaces 从 Kubernetes API 读取 Namespace。
func (c *ClientGoClient) ListNamespaces(ctx context.Context) ([]Namespace, error) {
	items, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing namespaces: %w", err)
	}
	result := make([]Namespace, 0, len(items.Items))
	for _, item := range items.Items {
		result = append(result, toNamespace(item))
	}
	return result, nil
}

func (c *ClientGoClient) ListNodes(ctx context.Context) ([]Node, error) {
	items, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing nodes: %w", err)
	}
	result := make([]Node, 0, len(items.Items))
	for _, item := range items.Items {
		result = append(result, toNode(item))
	}
	return result, nil
}

func (c *ClientGoClient) ListPods(ctx context.Context, namespace string) ([]Pod, error) {
	items, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing pods: %w", err)
	}
	result := make([]Pod, 0, len(items.Items))
	for _, item := range items.Items {
		result = append(result, toPod(item))
	}
	return result, nil
}

func (c *ClientGoClient) ListDeployments(ctx context.Context, namespace string) ([]Deployment, error) {
	items, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing deployments: %w", err)
	}
	result := make([]Deployment, 0, len(items.Items))
	for _, item := range items.Items {
		result = append(result, toDeployment(item))
	}
	return result, nil
}

func (c *ClientGoClient) ListServices(ctx context.Context, namespace string) ([]Service, error) {
	items, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing services: %w", err)
	}
	result := make([]Service, 0, len(items.Items))
	for _, item := range items.Items {
		result = append(result, toService(item))
	}
	return result, nil
}

func (c *ClientGoClient) PodLogs(ctx context.Context, namespace string, name string, opts LogOptions) (string, error) {
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	opts = normalizeLogOptions(opts)
	stream, err := c.clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
		Container:  opts.Container,
		Follow:     false,
		TailLines:  &opts.TailLines,
		LimitBytes: &opts.LimitBytes,
	}).Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting pod logs: %w", err)
	}
	defer stream.Close()
	reader := io.LimitReader(stream, opts.LimitBytes+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("error reading pod logs: %w", err)
	}
	return string(data), nil
}

func normalizeLogOptions(opts LogOptions) LogOptions {
	opts.Container = strings.TrimSpace(opts.Container)
	if opts.TailLines <= 0 {
		opts.TailLines = 500
	}
	if opts.TailLines > 5000 {
		opts.TailLines = 5000
	}
	if opts.LimitBytes <= 0 {
		opts.LimitBytes = 1024 * 1024
	}
	if opts.LimitBytes > 5*1024*1024 {
		opts.LimitBytes = 5 * 1024 * 1024
	}
	return opts
}

// ApplyYAML 使用 dynamic client 根据 YAML 创建资源。
func (c *ClientGoClient) ApplyYAML(ctx context.Context, namespace string, yamlContent string) (ApplyResult, error) {
	if strings.TrimSpace(yamlContent) == "" {
		return ApplyResult{}, fmt.Errorf("empty yaml content")
	}

	obj := &unstructured.Unstructured{}
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := decoder.Decode([]byte(yamlContent), nil, obj)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("error decoding yaml: %w", err)
	}
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("error getting mapping: %w", err)
	}
	if namespace == "" {
		namespace = obj.GetNamespace()
	}
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	obj.SetNamespace(namespace)

	resource := c.dynamic.Resource(mapping.Resource).Namespace(namespace)
	created, err := resource.Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return ApplyResult{}, fmt.Errorf("error creating object: %w", err)
	}
	return ApplyResult{Kind: created.GetKind(), Namespace: created.GetNamespace(), Name: created.GetName()}, nil
}

func (c *ClientGoClient) DeleteDeployment(ctx context.Context, namespace string, name string) error {
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	return c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *ClientGoClient) ScaleDeployment(ctx context.Context, namespace string, name string, replicas int) error {
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting deployment: %w", err)
	}
	deployment.Spec.Replicas = new(int32(replicas))
	_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating deployment: %w", err)
	}
	return nil
}

func buildRESTConfig(opts Options) (*rest.Config, error) {
	if opts.Mode == "in_cluster" {
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("error building in-cluster config: %w", err)
		}
		return restConfig, nil
	}
	restConfig, err := clientcmd.BuildConfigFromFlags(opts.APIServer, opts.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("error building config from flags: %w", err)
	}
	return restConfig, nil
}
