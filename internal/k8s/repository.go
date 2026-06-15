package k8s

import (
	"context"

	"compute-monitor-api/internal/page"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	SaveNamespaces(ctx context.Context, clusterID string, items []Namespace) error
	ListNamespaces(ctx context.Context, clusterID string, query page.Query) ([]Namespace, int64, error)
	SaveNodes(ctx context.Context, clusterID string, items []Node) error
	ListNodes(ctx context.Context, clusterID string, query page.Query) ([]Node, int64, error)
	GetNode(ctx context.Context, clusterID string, name string) (Node, error)
	SavePods(ctx context.Context, clusterID string, items []Pod) error
	ListPods(ctx context.Context, clusterID string, namespace string, query page.Query) ([]Pod, int64, error)
	GetPod(ctx context.Context, clusterID string, namespace string, name string) (Pod, error)
	SaveDeployments(ctx context.Context, clusterID string, items []Deployment) error
	ListDeployments(ctx context.Context, clusterID string, namespace string, query page.Query) ([]Deployment, int64, error)
	SaveServices(ctx context.Context, clusterID string, items []Service) error
	ListServices(ctx context.Context, clusterID string, namespace string, query page.Query) ([]Service, int64, error)
}

type MySQLRepository struct {
	db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func NewRepository(db *gorm.DB) *MySQLRepository {
	return NewMySQLRepository(db)
}

func (r *MySQLRepository) SaveNamespaces(ctx context.Context, clusterID string, items []Namespace) error {
	records := NewNamespaceRecords(clusterID, items)
	return r.upsert(ctx, &records)
}

func (r *MySQLRepository) ListNamespaces(ctx context.Context, clusterID string, query page.Query) ([]Namespace, int64, error) {
	var records []NamespaceRecord
	db := r.db.WithContext(ctx).Model(&NamespaceRecord{}).Where("cluster_id = ?", clusterID)
	total, err := count(db)
	if err != nil {
		return nil, 0, err
	}
	db = page.Apply(db, query)
	if err := db.Order("name ASC").Find(&records).Error; err != nil {
		return nil, 0, err
	}
	result := make([]Namespace, 0, len(records))
	for _, item := range records {
		result = append(result, item.ToDTO())
	}
	return result, total, nil
}

func (r *MySQLRepository) SaveNodes(ctx context.Context, clusterID string, items []Node) error {
	records := NewNodeRecords(clusterID, items)
	return r.upsert(ctx, &records)
}

func (r *MySQLRepository) ListNodes(ctx context.Context, clusterID string, query page.Query) ([]Node, int64, error) {
	var records []NodeRecord
	db := r.db.WithContext(ctx).Model(&NodeRecord{}).Where("cluster_id = ?", clusterID)
	total, err := count(db)
	if err != nil {
		return nil, 0, err
	}
	db = page.Apply(db, query)
	if err := db.Order("name ASC").Find(&records).Error; err != nil {
		return nil, 0, err
	}
	result := make([]Node, 0, len(records))
	for _, item := range records {
		result = append(result, item.ToDTO())
	}
	return result, total, nil
}

func (r *MySQLRepository) GetNode(ctx context.Context, clusterID string, name string) (Node, error) {
	var record NodeRecord
	if err := r.db.WithContext(ctx).First(&record, "cluster_id = ? AND name = ?", clusterID, name).Error; err != nil {
		return Node{}, err
	}
	return record.ToDTO(), nil
}

func (r *MySQLRepository) SavePods(ctx context.Context, clusterID string, items []Pod) error {
	records := NewPodRecords(clusterID, items)
	return r.upsert(ctx, &records)
}

func (r *MySQLRepository) ListPods(ctx context.Context, clusterID string, namespace string, query page.Query) ([]Pod, int64, error) {
	var records []PodRecord
	db := r.db.WithContext(ctx).Model(&PodRecord{}).Where("cluster_id = ?", clusterID)
	if namespace != "" {
		db = db.Where("namespace = ?", namespace)
	}
	total, err := count(db)
	if err != nil {
		return nil, 0, err
	}
	db = page.Apply(db, query)
	if err := db.Order("namespace ASC, name ASC").Find(&records).Error; err != nil {
		return nil, 0, err
	}
	result := make([]Pod, 0, len(records))
	for _, item := range records {
		result = append(result, item.ToDTO())
	}
	return result, total, nil
}

func (r *MySQLRepository) GetPod(ctx context.Context, clusterID string, namespace string, name string) (Pod, error) {
	var record PodRecord
	if err := r.db.WithContext(ctx).First(&record, "cluster_id = ? AND namespace = ? AND name = ?", clusterID, namespace, name).Error; err != nil {
		return Pod{}, err
	}
	return record.ToDTO(), nil
}

func (r *MySQLRepository) SaveDeployments(ctx context.Context, clusterID string, items []Deployment) error {
	records := NewDeploymentRecords(clusterID, items)
	return r.upsert(ctx, &records)
}

func (r *MySQLRepository) ListDeployments(ctx context.Context, clusterID string, namespace string, query page.Query) ([]Deployment, int64, error) {
	var records []DeploymentRecord
	db := r.db.WithContext(ctx).Model(&DeploymentRecord{}).Where("cluster_id = ?", clusterID)
	if namespace != "" {
		db = db.Where("namespace = ?", namespace)
	}
	total, err := count(db)
	if err != nil {
		return nil, 0, err
	}
	db = page.Apply(db, query)
	if err := db.Order("namespace ASC, name ASC").Find(&records).Error; err != nil {
		return nil, 0, err
	}
	result := make([]Deployment, 0, len(records))
	for _, item := range records {
		result = append(result, item.ToDTO())
	}
	return result, total, nil
}

func (r *MySQLRepository) SaveServices(ctx context.Context, clusterID string, items []Service) error {
	records := NewServiceRecords(clusterID, items)
	return r.upsert(ctx, &records)
}

func (r *MySQLRepository) ListServices(ctx context.Context, clusterID string, namespace string, query page.Query) ([]Service, int64, error) {
	var records []ServiceRecord
	db := r.db.WithContext(ctx).Model(&ServiceRecord{}).Where("cluster_id = ?", clusterID)
	if namespace != "" {
		db = db.Where("namespace = ?", namespace)
	}
	total, err := count(db)
	if err != nil {
		return nil, 0, err
	}
	db = page.Apply(db, query)
	if err := db.Order("namespace ASC, name ASC").Find(&records).Error; err != nil {
		return nil, 0, err
	}
	result := make([]Service, 0, len(records))
	for _, item := range records {
		result = append(result, item.ToDTO())
	}
	return result, total, nil
}

func count(db *gorm.DB) (int64, error) {
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *MySQLRepository) upsert(ctx context.Context, value interface{}) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(value).Error
}
