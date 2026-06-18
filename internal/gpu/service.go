package gpu

import (
	"context"
	"fmt"

	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/page"
	"compute-monitor-api/internal/prometheus"
)

type Service struct {
	repository  k8s.Repository
	promFactory prometheus.ClientFactory
}

func NewService(repository k8s.Repository, promFactory prometheus.ClientFactory) *Service {
	return &Service{repository: repository, promFactory: promFactory}
}

// List 从数据库中的 Node 缓存读取 GPU 容量并组装 GPU 分页列表。
func (s *Service) List(ctx context.Context, clusterID string, nodeName string, query page.Query) (page.Result[GPUResponse], error) {
	savedNodes, _, err := s.repository.ListNodes(ctx, clusterID, page.All())
	if err != nil {
		return page.Result[GPUResponse]{}, err
	}
	result := make([]GPUResponse, 0)
	for _, node := range savedNodes {
		if nodeName != "" && node.Name != nodeName {
			continue
		}
		for index := 0; index < node.GPUCount; index++ {
			result = append(result, GPUResponse{NodeName: node.Name, GPUIndex: index, GPUUUID: fmt.Sprintf("%s-%d", node.Name, index)})
		}
	}
	return page.Slice(result, query), nil
}

// Summary 返回 GPU 汇总信息。
func (s *Service) Summary(ctx context.Context, clusterID string) (SummaryResponse, error) {
	promClient, err := s.promFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return SummaryResponse{}, err
	}
	gpus, err := s.List(ctx, clusterID, "", page.All())
	if err != nil {
		return SummaryResponse{}, err
	}
	return SummaryResponse{
		ClusterID: clusterID,
		Total:     gpus.Total,
		Source:    promClient.BaseURL(),
	}, nil
}

// Top 返回 GPU 分页列表；排序后续可以基于 Prometheus 利用率实现。
func (s *Service) Top(ctx context.Context, clusterID string, query page.Query) (page.Result[GPUResponse], error) {
	return s.List(ctx, clusterID, "", query)
}

// Metric 查询节点 GPU 利用率指标。
func (s *Service) Metric(ctx context.Context, clusterID string, nodeName string, start int64, end int64, step int64) ([]prometheus.Point, error) {
	promClient, err := s.promFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`DCGM_FI_DEV_GPU_UTIL{node="%s"`, nodeName)
	return promClient.QueryRange(ctx, query, start, end, step)
}
