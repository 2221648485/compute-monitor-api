package cluster

import (
	"context"

	"compute-monitor-api/internal/alert"
	"compute-monitor-api/internal/audit"
	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/metrics"

	"gorm.io/gorm"
)

// DeleteRepository 定义删除集群配置时需要的事务性清理能力。
type DeleteRepository interface {
	Delete(ctx context.Context, clusterID string) error
}

// MySQLDeleteRepository 负责在一个数据库事务里删除集群配置和关联数据。
type MySQLDeleteRepository struct {
	db *gorm.DB
}

// NewDeleteRepository 创建集群删除仓储。
func NewDeleteRepository(db *gorm.DB) *MySQLDeleteRepository {
	return NewMySQLDeleteRepository(db)
}

// NewMySQLDeleteRepository 创建 MySQL 集群删除仓储。
func NewMySQLDeleteRepository(db *gorm.DB) *MySQLDeleteRepository {
	return &MySQLDeleteRepository{db: db}
}

// Delete 删除集群配置以及该集群下已经入库的资源、指标、告警和审计数据。
func (r *MySQLDeleteRepository) Delete(ctx context.Context, clusterID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.deleteClusterInTransaction(tx, clusterID)
	})
}

func (r *MySQLDeleteRepository) deleteClusterInTransaction(tx *gorm.DB, clusterID string) error {
	if err := r.ensureClusterExists(tx, clusterID); err != nil {
		return err
	}
	if err := r.deleteRelatedRecords(tx, clusterID); err != nil {
		return err
	}
	return r.deleteClusterRecord(tx, clusterID)
}

func (r *MySQLDeleteRepository) ensureClusterExists(tx *gorm.DB, clusterID string) error {
	var count int64
	if err := tx.Model(&Record{}).Where("id = ?", clusterID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *MySQLDeleteRepository) deleteRelatedRecords(tx *gorm.DB, clusterID string) error {
	relatedRecords := []any{
		&k8s.NamespaceRecord{},
		&k8s.NodeRecord{},
		&k8s.PodRecord{},
		&k8s.DeploymentRecord{},
		&k8s.ServiceRecord{},
		&metrics.PointRecord{},
		&alert.RuleRecord{},
		&alert.EventRecord{},
		&audit.LogRecord{},
	}
	for _, record := range relatedRecords {
		if err := tx.Where("cluster_id = ?", clusterID).Delete(record).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLDeleteRepository) deleteClusterRecord(tx *gorm.DB, clusterID string) error {
	return tx.Delete(&Record{}, "id = ?", clusterID).Error
}
