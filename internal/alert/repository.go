package alert

import (
	"context"

	"compute-monitor-api/internal/page"

	"gorm.io/gorm"
)

// Repository 定义告警规则和告警事件的数据访问能力。
type Repository interface {
	CreateRule(ctx context.Context, rule RuleRecord) (RuleRecord, error)
	ListRules(ctx context.Context, query page.Query) ([]RuleRecord, int64, error)
	UpdateRule(ctx context.Context, id string, rule RuleRecord) error
	DeleteRule(ctx context.Context, id string) error
	ListEvents(ctx context.Context, clusterID string, status string, query page.Query) ([]EventRecord, int64, error)
}

// MySQLRepository 是基于 GORM 的告警仓储实现。
type MySQLRepository struct {
	db *gorm.DB
}

// NewRepository 创建告警仓储。
func NewRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) CreateRule(ctx context.Context, rule RuleRecord) (RuleRecord, error) {
	if err := r.db.WithContext(ctx).Create(&rule).Error; err != nil {
		return RuleRecord{}, err
	}
	return rule, nil
}

func (r *MySQLRepository) ListRules(ctx context.Context, query page.Query) ([]RuleRecord, int64, error) {
	var rules []RuleRecord
	db := r.db.WithContext(ctx).Model(&RuleRecord{})
	total, err := count(db)
	if err != nil {
		return nil, 0, err
	}
	db = page.Apply(db, query)
	if err := db.Order("id DESC").Find(&rules).Error; err != nil {
		return nil, 0, err
	}
	return rules, total, nil
}

func (r *MySQLRepository) UpdateRule(ctx context.Context, id string, rule RuleRecord) error {
	return r.db.WithContext(ctx).Model(&RuleRecord{}).Where("id = ?", id).Updates(rule).Error
}

func (r *MySQLRepository) DeleteRule(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&RuleRecord{}, "id = ?", id).Error
}

func (r *MySQLRepository) ListEvents(ctx context.Context, clusterID string, status string, query page.Query) ([]EventRecord, int64, error) {
	var events []EventRecord
	db := r.db.WithContext(ctx).Model(&EventRecord{})
	if clusterID != "" {
		db = db.Where("cluster_id = ?", clusterID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	total, err := count(db)
	if err != nil {
		return nil, 0, err
	}
	db = page.Apply(db, query)
	if err := db.Order("id DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func count(db *gorm.DB) (int64, error) {
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
