package cluster

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/page"

	"gorm.io/gorm"
)

// Service 负责集群管理业务逻辑。
type Service struct {
	repository       Repository
	deleteRepository DeleteRepository
	k8sFactory       k8s.ClientFactory
	k8sConfig        config.K8sConfig
}

// NewService 创建集群服务。
func NewService(repository Repository, deleteRepository DeleteRepository, k8sFactory k8s.ClientFactory) *Service {
	return NewServiceWithConfig(repository, deleteRepository, k8sFactory, config.K8sConfig{})
}

// NewServiceWithConfig 创建集群服务，并注入 kubeconfig 文件保存目录等配置。
func NewServiceWithConfig(repository Repository, deleteRepository DeleteRepository, k8sFactory k8s.ClientFactory, k8sConfig config.K8sConfig) *Service {
	return &Service{
		repository:       repository,
		deleteRepository: deleteRepository,
		k8sFactory:       k8sFactory,
		k8sConfig:        k8sConfig,
	}
}

// Create 创建集群配置。
func (s *Service) Create(ctx context.Context, req CreateClusterRequest) (Cluster, error) {
	cluster := normalizeCluster(Cluster{
		ID:             req.ID,
		Name:           req.Name,
		AccessMode:     req.AccessMode,
		APIServer:      req.APIServer,
		KubeconfigPath: req.KubeconfigPath,
		PrometheusURL:  req.PrometheusURL,
		Description:    req.Description,
		Status:         req.Status,
	})

	if cluster.Status == "" {
		cluster.Status = StatusRunning
	}
	if err := s.prepareClusterAccess(cluster.ID, &cluster, req); err != nil {
		return Cluster{}, err
	}
	if err := validateCluster(cluster); err != nil {
		return Cluster{}, err
	}
	if _, err := s.repository.Get(ctx, cluster.ID); err == nil {
		return Cluster{}, ErrClusterExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Cluster{}, err
	}
	return s.repository.Create(ctx, cluster)
}

// CreateFromUpload 保存上传的 kubeconfig 文件后创建集群。
func (s *Service) CreateFromUpload(ctx context.Context, req CreateClusterRequest, file *multipart.FileHeader) (Cluster, error) {
	if file == nil {
		return Cluster{}, ErrKubeconfigRequired
	}
	req.AccessMode = AccessModeUpload
	cluster := normalizeCluster(Cluster{
		ID:             req.ID,
		Name:           req.Name,
		AccessMode:     req.AccessMode,
		APIServer:      req.APIServer,
		PrometheusURL:  req.PrometheusURL,
		Description:    req.Description,
		Status:         req.Status,
		KubeconfigPath: s.kubeconfigPath(req.ID),
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
	if err := s.saveUploadedKubeconfig(file, cluster.KubeconfigPath); err != nil {
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
		AccessMode:     req.AccessMode,
		APIServer:      req.APIServer,
		KubeconfigPath: req.KubeconfigPath,
		PrometheusURL:  req.PrometheusURL,
		Description:    req.Description,
		Status:         req.Status,
	})
	if cluster.Status == "" {
		cluster.Status = StatusRunning
	}
	if err := s.prepareClusterAccess(cluster.ID, &cluster, CreateClusterRequest{
		ID:             cluster.ID,
		Name:           cluster.Name,
		AccessMode:     cluster.AccessMode,
		APIServer:      cluster.APIServer,
		KubeconfigPath: cluster.KubeconfigPath,
		CACert:         req.CACert,
		BearerToken:    req.BearerToken,
		ClientCert:     req.ClientCert,
		ClientKey:      req.ClientKey,
		Insecure:       req.Insecure,
		PrometheusURL:  cluster.PrometheusURL,
		Description:    cluster.Description,
		Status:         cluster.Status,
	}); err != nil {
		return Cluster{}, err
	}
	if err := validateCluster(cluster); err != nil {
		return Cluster{}, err
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
	if err := validateAccessMode(cluster.AccessMode); err != nil {
		return err
	}
	if cluster.KubeconfigPath == "" {
		return ErrKubeconfigRequired
	}
	if err := validatePrometheusURL(cluster.PrometheusURL); err != nil {
		return err
	}
	return nil
}

func normalizeCluster(cluster Cluster) Cluster {
	cluster.ID = strings.TrimSpace(cluster.ID)
	cluster.Name = strings.TrimSpace(cluster.Name)
	cluster.AccessMode = strings.TrimSpace(cluster.AccessMode)
	cluster.APIServer = strings.TrimSpace(cluster.APIServer)
	cluster.KubeconfigPath = strings.TrimSpace(cluster.KubeconfigPath)
	cluster.PrometheusURL = strings.TrimSpace(cluster.PrometheusURL)
	cluster.Description = strings.TrimSpace(cluster.Description)
	cluster.Status = strings.TrimSpace(cluster.Status)
	return cluster
}

func validateAccessMode(accessMode string) error {
	switch accessMode {
	case AccessModePath, AccessModeUpload, AccessModeManual:
		return nil
	default:
		return ErrInvalidAccessMode
	}
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

// prepareClusterAccess 根据 access_mode 准备集群连接配置。
// path 模式只校验后端可读路径；upload 模式使用上传后保存的路径；
// manual 模式会根据 API Server 和凭证生成 kubeconfig 文件，并把路径写回集群配置。
func (s *Service) prepareClusterAccess(clusterID string, cluster *Cluster, req CreateClusterRequest) error {
	if cluster.AccessMode == "" {
		cluster.AccessMode = AccessModePath
	}
	switch cluster.AccessMode {
	case AccessModePath:
		if cluster.KubeconfigPath == "" {
			return ErrKubeconfigRequired
		}
	case AccessModeManual:
		if cluster.APIServer == "" {
			return ErrAPIServerRequired
		}
		if strings.TrimSpace(req.BearerToken) == "" && (strings.TrimSpace(req.ClientCert) == "" || strings.TrimSpace(req.ClientKey) == "") {
			return ErrCredentialRequired
		}
		content := buildManualKubeconfig(clusterID, cluster.APIServer, req)
		path := s.kubeconfigPath(clusterID)
		if err := writeKubeconfigFile(path, content); err != nil {
			return err
		}
		cluster.KubeconfigPath = path
	case AccessModeUpload:
		if cluster.KubeconfigPath == "" {
			return ErrKubeconfigRequired
		}
	default:
		return ErrInvalidAccessMode
	}
	return nil
}

// saveUploadedKubeconfig 把前端上传的 kubeconfig 保存到后端受控目录。
// 这里限制文件大小，避免用户上传过大的无关文件占用磁盘。
func (s *Service) saveUploadedKubeconfig(fileHeader *multipart.FileHeader, targetPath string) error {
	source, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer source.Close()
	data, err := io.ReadAll(io.LimitReader(source, 4*1024*1024+1))
	if err != nil {
		return err
	}
	if len(data) == 0 || len(data) > 4*1024*1024 {
		return ErrKubeconfigRequired
	}
	return writeKubeconfigFile(targetPath, string(data))
}

// kubeconfigPath 生成当前集群在后端本地的 kubeconfig 保存路径。
func (s *Service) kubeconfigPath(clusterID string) string {
	dir := strings.TrimSpace(s.k8sConfig.KubeconfigStoreDir)
	if dir == "" {
		dir = "data/kubeconfigs"
	}
	return filepath.Join(dir, safeClusterFileName(clusterID)+".yaml")
}

// writeKubeconfigFile 以仅当前用户可读写的权限写入 kubeconfig 文件。
func writeKubeconfigFile(path string, content string) error {
	if strings.TrimSpace(content) == "" {
		return ErrKubeconfigRequired
	}
	cleanPath := filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0700); err != nil {
		return err
	}
	return os.WriteFile(cleanPath, []byte(content), 0600)
}

// buildManualKubeconfig 根据用户填写的 API Server、CA 和认证凭证生成标准 kubeconfig。
// 支持 bearer token 和 client certificate 两种 Kubernetes 认证方式。
func buildManualKubeconfig(clusterID string, apiServer string, req CreateClusterRequest) string {
	clusterName := safeClusterFileName(clusterID)
	userName := clusterName + "-user"
	contextName := clusterName + "-context"
	lines := []string{
		"apiVersion: v1",
		"kind: Config",
		"clusters:",
		"- name: " + clusterName,
		"  cluster:",
		"    server: " + apiServer,
	}
	if strings.TrimSpace(req.CACert) != "" {
		lines = append(lines, "    certificate-authority-data: "+base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(req.CACert))))
	} else if req.Insecure {
		lines = append(lines, "    insecure-skip-tls-verify: true")
	}
	lines = append(lines,
		"users:",
		"- name: "+userName,
		"  user:",
	)
	if strings.TrimSpace(req.BearerToken) != "" {
		lines = append(lines, "    token: "+strings.TrimSpace(req.BearerToken))
	} else {
		lines = append(lines,
			"    client-certificate-data: "+base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(req.ClientCert))),
			"    client-key-data: "+base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(req.ClientKey))),
		)
	}
	lines = append(lines,
		"contexts:",
		"- name: "+contextName,
		"  context:",
		"    cluster: "+clusterName,
		"    user: "+userName,
		"current-context: "+contextName,
	)
	return strings.Join(lines, "\n") + "\n"
}

func safeClusterFileName(clusterID string) string {
	value := strings.TrimSpace(clusterID)
	if value == "" {
		return "cluster"
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	value = re.ReplaceAllString(value, "-")
	value = strings.Trim(value, ".-")
	if value == "" {
		return "cluster"
	}
	return value
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
