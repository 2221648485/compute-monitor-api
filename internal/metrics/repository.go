package metrics

import (
	"compute-monitor-api/internal/prometheus"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	SavePoints(ctx context.Context, clusterID string, nodeName string, metric string, points []prometheus.Point) error
	ListPoints(ctx context.Context, clusterID string, nodeName string, metric string) ([]prometheus.Point, error)
}

// MySQLRepository 是基于 GORM 的指标仓储实现。
type MySQLRepository struct {
	db *gorm.DB
}

// NewRepository 创建指标仓储。
func NewRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) SavePoints(ctx context.Context, clusterID string, nodeName string, metric string, points []prometheus.Point) error {
	records := NewPointRecords(clusterID, nodeName, metric, points)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&records).Error
}

func (r *MySQLRepository) ListPoints(ctx context.Context, clusterID string, nodeName string, metric string) ([]prometheus.Point, error) {
	var records []PointRecord
	if err := r.db.WithContext(ctx).
		Where("cluster_id = ? AND node_name = ? AND metric = ?", clusterID, nodeName, metric).
		Order("timestamp ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	result := make([]prometheus.Point, len(records))
	for _, item := range records {
		result = append(result, item.ToDTO())
	}
	return result, nil
}
