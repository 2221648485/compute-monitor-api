package audit

import (
	"context"

	"compute-monitor-api/internal/page"
)

// Service 负责审计日志业务逻辑。
type Service struct {
	repository Repository
}

// NewService 创建审计日志服务。
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// List 分页查询审计日志。
func (s *Service) List(ctx context.Context, clusterID string, action string, query page.Query) (page.Result[LogRecord], error) {
	items, total, err := s.repository.List(ctx, clusterID, action, query)
	if err != nil {
		return page.Result[LogRecord]{}, err
	}
	return page.NewResult(items, total, query), nil
}
