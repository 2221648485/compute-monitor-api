package k8ssync

import (
	"context"
	"strings"

	"compute-monitor-api/internal/k8s"
)

type Result struct {
	ClusterID   string `json:"clusterId"`
	Namespaces  int    `json:"namespaces"`
	Nodes       int    `json:"nodes"`
	Pods        int    `json:"pods"`
	Deployments int    `json:"deployments"`
	Services    int    `json:"services"`
}

// Service 负责把 Kubernetes 当前资源同步到数据库缓存。
type Service struct {
	k8sFactory k8s.ClientFactory
	repository k8s.Repository
}

func NewService(k8sFactory k8s.ClientFactory, repository k8s.Repository) *Service {
	return &Service{k8sFactory: k8sFactory, repository: repository}
}

func (s *Service) SyncAll(ctx context.Context, clusterID string, namespace string) (Result, error) {
	result := Result{ClusterID: strings.TrimSpace(clusterID)}

	namespaces, err := s.SyncNamespaces(ctx, clusterID)
	if err != nil {
		return Result{}, err
	}
	result.Namespaces = namespaces

	nodes, err := s.SyncNodes(ctx, clusterID)
	if err != nil {
		return Result{}, err
	}
	result.Nodes = nodes

	pods, err := s.SyncPods(ctx, clusterID, namespace)
	if err != nil {
		return Result{}, err
	}
	result.Pods = pods

	deployments, err := s.SyncDeployments(ctx, clusterID, namespace)
	if err != nil {
		return Result{}, err
	}
	result.Deployments = deployments

	services, err := s.SyncServices(ctx, clusterID, namespace)
	if err != nil {
		return Result{}, err
	}
	result.Services = services

	return result, nil
}

func (s *Service) SyncNamespaces(ctx context.Context, clusterID string) (int, error) {
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return 0, err
	}
	items, err := client.ListNamespaces(ctx)
	if err != nil {
		return 0, err
	}
	if err := s.repository.SaveNamespaces(ctx, clusterID, items); err != nil {
		return 0, err
	}
	return len(items), nil
}

func (s *Service) SyncNodes(ctx context.Context, clusterID string) (int, error) {
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return 0, err
	}
	items, err := client.ListNodes(ctx)
	if err != nil {
		return 0, err
	}
	if err := s.repository.SaveNodes(ctx, clusterID, items); err != nil {
		return 0, err
	}
	return len(items), nil
}

func (s *Service) SyncPods(ctx context.Context, clusterID string, namespace string) (int, error) {
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return 0, err
	}
	items, err := client.ListPods(ctx, namespace)
	if err != nil {
		return 0, err
	}
	if err := s.repository.SavePods(ctx, clusterID, items); err != nil {
		return 0, err
	}
	return len(items), nil
}

func (s *Service) SyncDeployments(ctx context.Context, clusterID string, namespace string) (int, error) {
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return 0, err
	}
	items, err := client.ListDeployments(ctx, namespace)
	if err != nil {
		return 0, err
	}
	if err := s.repository.SaveDeployments(ctx, clusterID, items); err != nil {
		return 0, err
	}
	return len(items), nil
}

func (s *Service) SyncServices(ctx context.Context, clusterID string, namespace string) (int, error) {
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return 0, err
	}
	items, err := client.ListServices(ctx, namespace)
	if err != nil {
		return 0, err
	}
	if err := s.repository.SaveServices(ctx, clusterID, items); err != nil {
		return 0, err
	}
	return len(items), nil
}
