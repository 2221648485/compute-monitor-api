package alert

import (
	"context"

	"compute-monitor-api/internal/page"
)

// Service 负责告警规则和告警事件业务逻辑。
type Service struct {
	repository Repository
}

// NewService 创建告警服务。
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateRule(ctx context.Context, rule RuleRecord) (RuleRecord, error) {
	return s.repository.CreateRule(ctx, rule)
}

func (s *Service) ListRules(ctx context.Context, query page.Query) (page.Result[RuleRecord], error) {
	items, total, err := s.repository.ListRules(ctx, query)
	if err != nil {
		return page.Result[RuleRecord]{}, err
	}
	return page.NewResult(items, total, query), nil
}

func (s *Service) UpdateRule(ctx context.Context, id string, rule RuleRecord) error {
	return s.repository.UpdateRule(ctx, id, rule)
}

func (s *Service) DeleteRule(ctx context.Context, id string) error {
	return s.repository.DeleteRule(ctx, id)
}

func (s *Service) ListEvents(ctx context.Context, clusterID string, status string, query page.Query) (page.Result[EventRecord], error) {
	items, total, err := s.repository.ListEvents(ctx, clusterID, status, query)
	if err != nil {
		return page.Result[EventRecord]{}, err
	}
	return page.NewResult(items, total, query), nil
}
