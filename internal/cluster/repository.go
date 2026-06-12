package cluster

import (
	"context"
	"strings"

	"compute-monitor-api/internal/alert"
	"compute-monitor-api/internal/audit"
	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/metrics"
	"compute-monitor-api/internal/page"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, cluster Cluster) (Cluster, error)
	Update(ctx context.Context, cluster Cluster) (Cluster, error)
	List(ctx context.Context, query ListQuery) ([]Cluster, int64, error)
	Get(ctx context.Context, clusterID string) (Cluster, error)
	Delete(ctx context.Context, clusterID string) error
}

type MySQLRepository struct {
	db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// Create 新增集群配置。
func (r *MySQLRepository) Create(ctx context.Context, cluster Cluster) (Cluster, error) {
	record := newRecord(cluster)
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return Cluster{}, err
	}
	return record.ToDTO(), nil
}

// Update 修改集群配置。
func (r *MySQLRepository) Update(ctx context.Context, cluster Cluster) (Cluster, error) {
	updates := map[string]interface{}{
		"name":            cluster.Name,
		"kubeconfig_path": cluster.KubeconfigPath,
		"prometheus_url":  cluster.PrometheusURL,
		"description":     cluster.Description,
		"status":          cluster.Status,
	}
	result := r.db.WithContext(ctx).Model(&Record{}).Where("id = ?", cluster.ID).Updates(updates)
	if result.Error != nil {
		return Cluster{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Cluster{}, gorm.ErrRecordNotFound
	}
	return r.Get(ctx, cluster.ID)
}

// List 分页查询集群列表。
func (r *MySQLRepository) List(ctx context.Context, query ListQuery) ([]Cluster, int64, error) {
	query.Query = page.Normalize(query.Query)

	db := applyListQuery(r.db.WithContext(ctx).Model(&Record{}), query)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []Record
	if err := db.
		Order("id ASC").
		Offset(page.Offset(query.Query)).
		Limit(query.Size).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	// 转换为DTO
	result := make([]Cluster, 0, len(records))
	for _, item := range records {
		result = append(result, item.ToDTO())
	}
	return result, total, nil
}

// Get 查询集群详情。
func (r *MySQLRepository) Get(ctx context.Context, clusterID string) (Cluster, error) {
	var record Record
	if err := r.db.WithContext(ctx).First(&record, "id = ?", clusterID).Error; err != nil {
		return Cluster{}, err
	}
	return record.ToDTO(), nil
}

// Delete 删除集群配置及关联缓存数据。
func (r *MySQLRepository) Delete(ctx context.Context, clusterID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.deleteClusterInTransaction(tx, clusterID)
	})
}

func (r *MySQLRepository) deleteClusterInTransaction(tx *gorm.DB, clusterID string) error {
	var count int64
	if err := tx.Model(&Record{}).Where("id = ?", clusterID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}

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

	return tx.Delete(&Record{}, "id = ?", clusterID).Error
}

func applyListQuery(db *gorm.DB, query ListQuery) *gorm.DB {
	keyword := strings.TrimSpace(query.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("id LIKE ? OR name LIKE ? OR description LIKE ?", like, like, like)
	}
	if strings.TrimSpace(query.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(query.Status))
	}
	return db
}
