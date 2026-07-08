package node

import (
	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/page"
	"context"
)

type Service struct {
	repository k8s.Repository
}

func NewService(repository k8s.Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListNamespaces(ctx context.Context, clusterID string, query page.Query) (page.Result[k8s.Namespace], error) {
	saved, total, err := s.repository.ListNamespaces(ctx, clusterID, query)
	if err != nil {
		return page.Result[k8s.Namespace]{}, err
	}
	return page.NewResult(saved, total, query), nil
}

// ListNodes 从数据库分页查询 Node。
func (s *Service) ListNodes(ctx context.Context, clusterID string, query page.Query) (page.Result[k8s.Node], error) {
	saved, total, err := s.repository.ListNodes(ctx, clusterID, query)
	if err != nil {
		return page.Result[k8s.Node]{}, err
	}
	return page.NewResult(saved, total, query), nil
}

func (s *Service) GetNode(ctx context.Context, clusterID string, nodeName string) (k8s.Node, error) {
	return s.repository.GetNode(ctx, clusterID, nodeName)
}

func (s *Service) ListPods(ctx context.Context, clusterID string, namespace string, query page.Query) (page.Result[k8s.Pod], error) {
	saved, total, err := s.repository.ListPods(ctx, clusterID, namespace, "", query)
	if err != nil {
		return page.Result[k8s.Pod]{}, err
	}
	return page.NewResult(saved, total, query), nil
}
