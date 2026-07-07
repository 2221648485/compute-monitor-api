package audit

import (
	"context"

	"compute-monitor-api/internal/page"

	"gorm.io/gorm"
)

// Repository 定义审计日志数据访问能力。
type Repository interface {
	List(ctx context.Context, clusterID string, action string, query page.Query) ([]LogRecord, int64, error)
}

// MySQLRepository 是基于 GORM 的审计日志仓储。
type MySQLRepository struct {
	db *gorm.DB
}

// NewRepository 创建审计日志仓储。
func NewRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// List 分页查询审计日志列表。
func (r *MySQLRepository) List(ctx context.Context, clusterID string, action string, query page.Query) ([]LogRecord, int64, error) {
	var logs []LogRecord
	db := r.db.WithContext(ctx).Model(&LogRecord{})
	if clusterID != "" {
		db = db.Where("cluster_id = ?", clusterID)
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	db = page.Apply(db, query)
	if err := db.Order("id DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
