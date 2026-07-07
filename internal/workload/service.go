package workload

import (
	"context"

	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/page"
)

// Service 负责工作负载查询和操作。
type Service struct {
	k8sFactory k8s.ClientFactory
	repository k8s.Repository
}

func NewService(k8sFactory k8s.ClientFactory, repository k8s.Repository) *Service {
	return &Service{k8sFactory: k8sFactory, repository: repository}
}

func (s *Service) ListPods(ctx context.Context, clusterID string, namespace string, query page.Query) (page.Result[k8s.Pod], error) {
	saved, total, err := s.repository.ListPods(ctx, clusterID, namespace, query)
	if err != nil {
		return page.Result[k8s.Pod]{}, err
	}
	return page.NewResult(saved, total, query), nil
}

func (s *Service) GetPod(ctx context.Context, clusterID string, namespace string, name string) (k8s.Pod, error) {
	return s.repository.GetPod(ctx, clusterID, namespace, name)
}

func (s *Service) PodLogs(ctx context.Context, clusterID string, namespace string, name string, opts k8s.LogOptions) (PodLogsResponse, error) {
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return PodLogsResponse{}, err
	}
	logs, err := client.PodLogs(ctx, namespace, name, opts)
	if err != nil {
		return PodLogsResponse{}, err
	}
	return PodLogsResponse{
		Namespace:  namespace,
		Name:       name,
		Container:  opts.Container,
		TailLines:  opts.TailLines,
		LimitBytes: opts.LimitBytes,
		Logs:       logs,
	}, nil
}

func (s *Service) ListDeployments(ctx context.Context, clusterID string, namespace string, query page.Query) (page.Result[k8s.Deployment], error) {
	saved, total, err := s.repository.ListDeployments(ctx, clusterID, namespace, query)
	if err != nil {
		return page.Result[k8s.Deployment]{}, err
	}
	return page.NewResult(saved, total, query), nil
}

func (s *Service) ListServices(ctx context.Context, clusterID string, namespace string, query page.Query) (page.Result[k8s.Service], error) {
	saved, total, err := s.repository.ListServices(ctx, clusterID, namespace, query)
	if err != nil {
		return page.Result[k8s.Service]{}, err
	}
	return page.NewResult(saved, total, query), nil
}

func (s *Service) ApplyYAML(ctx context.Context, clusterID string, namespace string, yamlContent string) (k8s.ApplyResult, error) {
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return k8s.ApplyResult{}, err
	}
	return client.ApplyYAML(ctx, namespace, yamlContent)
}

func (s *Service) DeleteDeployment(ctx context.Context, clusterID string, namespace string, name string) error {
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return err
	}
	return client.DeleteDeployment(ctx, namespace, name)
}

func (s *Service) ScaleDeployment(ctx context.Context, clusterID string, namespace string, name string, replicas int) error {
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return err
	}
	return client.ScaleDeployment(ctx, namespace, name, replicas)
}
