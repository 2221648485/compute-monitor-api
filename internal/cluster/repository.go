package cluster

import (
	"context"
	"strings"

	"compute-monitor-api/internal/page"

	"gorm.io/gorm"
)

// Repository 定义集群配置的基础数据访问能力。
type Repository interface {
	Create(ctx context.Context, cluster Cluster) (Cluster, error)
	Update(ctx context.Context, cluster Cluster) (Cluster, error)
	List(ctx context.Context, query ListQuery) ([]Cluster, int64, error)
	Get(ctx context.Context, clusterID string) (Cluster, error)
}

// MySQLRepository 是基于 GORM 的 MySQL 集群配置仓储实现。
type MySQLRepository struct {
	db *gorm.DB
}

// NewRepository 创建集群配置仓储。
func NewRepository(db *gorm.DB) *MySQLRepository {
	return NewMySQLRepository(db)
}

// NewMySQLRepository 创建 MySQL 集群配置仓储。
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
	db := applyListQuery(r.db.WithContext(ctx).Model(&Record{}), query)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []Record
	db = page.Apply(db, query.Query)
	if err := db.Order("id ASC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

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
