package cluster

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/page"

	"gorm.io/gorm"
)

// Service 负责集群管理业务逻辑。
type Service struct {
	repository       Repository
	deleteRepository DeleteRepository
	k8sFactory       k8s.ClientFactory
}

// NewService 创建集群服务。
func NewService(repository Repository, deleteRepository DeleteRepository, k8sFactory k8s.ClientFactory) *Service {
	return &Service{
		repository:       repository,
		deleteRepository: deleteRepository,
		k8sFactory:       k8sFactory,
	}
}

// Create 创建集群配置。
func (s *Service) Create(ctx context.Context, req CreateClusterRequest) (Cluster, error) {
	cluster := normalizeCluster(Cluster{
		ID:             req.ID,
		Name:           req.Name,
		KubeconfigPath: req.KubeconfigPath,
		PrometheusURL:  req.PrometheusURL,
		Description:    req.Description,
		Status:         req.Status,
	})

	if err := validateCluster(cluster); err != nil {
		return Cluster{}, err
	}
	if cluster.Status == "" {
		cluster.Status = StatusRunning
	}
	if _, err := s.repository.Get(ctx, cluster.ID); err == nil {
		return Cluster{}, ErrClusterExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Cluster{}, err
	}
	return s.repository.Create(ctx, cluster)
}

// List 分页查询集群列表。
func (s *Service) List(ctx context.Context, query ListQuery) (page.Result[Cluster], error) {
	items, total, err := s.repository.List(ctx, query)
	if err != nil {
		return page.Result[Cluster]{}, err
	}
	return page.NewResult(items, total, query.Query), nil
}

// Get 查询集群详情。
func (s *Service) Get(ctx context.Context, clusterID string) (Cluster, error) {
	cluster, err := s.repository.Get(ctx, clusterID)
	if err != nil {
		return Cluster{}, normalizeRecordNotFound(err)
	}
	return cluster, nil
}

// Update 修改集群配置。
func (s *Service) Update(ctx context.Context, clusterID string, req UpdateClusterRequest) (Cluster, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return Cluster{}, ErrClusterIDRequired
	}

	cluster := normalizeCluster(Cluster{
		ID:             clusterID,
		Name:           req.Name,
		KubeconfigPath: req.KubeconfigPath,
		PrometheusURL:  req.PrometheusURL,
		Description:    req.Description,
		Status:         req.Status,
	})
	if err := validateCluster(cluster); err != nil {
		return Cluster{}, err
	}
	if cluster.Status == "" {
		cluster.Status = StatusRunning
	}

	updated, err := s.repository.Update(ctx, cluster)
	if err != nil {
		return Cluster{}, normalizeRecordNotFound(err)
	}
	return updated, nil
}

// TestConnection 根据集群配置创建 client-go 客户端，并真实请求 Kubernetes API。
func (s *Service) TestConnection(ctx context.Context, clusterID string) (TestConnectionResponse, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return TestConnectionResponse{}, ErrClusterIDRequired
	}
	if _, err := s.repository.Get(ctx, clusterID); err != nil {
		return TestConnectionResponse{}, normalizeRecordNotFound(err)
	}
	client, err := s.k8sFactory.ForCluster(ctx, clusterID)
	if err != nil {
		return TestConnectionResponse{}, err
	}
	namespaces, err := client.ListNamespaces(ctx)
	if err != nil {
		return TestConnectionResponse{}, err
	}
	return TestConnectionResponse{
		ClusterID:      clusterID,
		Connected:      true,
		NamespaceCount: len(namespaces),
	}, nil
}

// Delete 删除集群配置。
func (s *Service) Delete(ctx context.Context, clusterID string) error {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return ErrClusterIDRequired
	}
	if err := s.deleteRepository.Delete(ctx, clusterID); err != nil {
		return normalizeRecordNotFound(err)
	}
	return nil
}

func validateCluster(cluster Cluster) error {
	if cluster.ID == "" {
		return ErrClusterIDRequired
	}
	if cluster.Name == "" {
		return ErrClusterNameRequired
	}
	if err := validateClusterStatus(cluster.Status); err != nil {
		return err
	}
	if err := validatePrometheusURL(cluster.PrometheusURL); err != nil {
		return err
	}
	return nil
}

func normalizeCluster(cluster Cluster) Cluster {
	cluster.ID = strings.TrimSpace(cluster.ID)
	cluster.Name = strings.TrimSpace(cluster.Name)
	cluster.KubeconfigPath = strings.TrimSpace(cluster.KubeconfigPath)
	cluster.PrometheusURL = strings.TrimSpace(cluster.PrometheusURL)
	cluster.Description = strings.TrimSpace(cluster.Description)
	cluster.Status = strings.TrimSpace(cluster.Status)
	return cluster
}

func validateClusterStatus(status string) error {
	if status == "" {
		return nil
	}
	switch status {
	case StatusRunning, StatusNotReady, StatusDisabled:
		return nil
	default:
		return ErrInvalidClusterStatus
	}
}

func validatePrometheusURL(rawURL string) error {
	if rawURL == "" {
		return nil
	}
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidPrometheusURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidPrometheusURL
	}
	return nil
}

func newK8sClientForCluster(cluster Cluster) (k8s.Client, error) {
	opts := k8s.Options{
		Mode:           "in_cluster",
		KubeconfigPath: cluster.KubeconfigPath,
	}
	if cluster.KubeconfigPath != "" {
		opts.Mode = "kubeconfig"
	}
	return k8s.NewClient(opts)
}

func normalizeRecordNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrClusterNotFound
	}
	return err
}
