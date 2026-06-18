package metrics

import (
	"compute-monitor-api/internal/prometheus"
	"context"
	"fmt"
)

type Service struct {
	promFactory prometheus.ClientFactory
	repository  Repository
}

// NewService 创建指标服务。
func NewService(promFactory prometheus.ClientFactory, repository Repository) *Service {
	return &Service{promFactory: promFactory, repository: repository}
}

func (s Service) NodeMetric(ctx context.Context, clusterID string, nodeName string, metric string, unit string, start int64, end int64, step int64) (Series, error) {
	promClient, err := s.promFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return Series{}, nil
	}
	query := fmt.Sprintf(`%s{node="%s"}`, metric, nodeName)
	points, err := promClient.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return Series{}, nil
	}
	if err := s.repository.SavePoints(ctx, clusterID, nodeName, metric, points); err != nil {
		return Series{}, err
	}
	values, err := s.repository.ListPoints(ctx, clusterID, nodeName, metric)
	if err != nil {
		return Series{}, err
	}
	return Series{
		NodeName: nodeName,
		Metric:   metric,
		Unit:     unit,
		Source:   promClient.BaseURL(),
		Values:   values,
	}, nil
}
