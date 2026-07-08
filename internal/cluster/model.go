package cluster

import "time"

const (
	// AccessModePath 表示使用后端服务器上已经存在的 kubeconfig 文件路径接入集群。
	AccessModePath = "path"
	// AccessModeUpload 表示由前端上传 kubeconfig 文件，后端保存后再接入集群。
	AccessModeUpload = "upload"
	// AccessModeManual 表示由用户填写 API Server 和凭证，后端生成 kubeconfig 后再接入集群。
	AccessModeManual = "manual"

	// StatusRunning 表示集群配置可用。
	StatusRunning = "Running"
	// StatusNotReady 表示集群暂时不可用或连接异常。
	StatusNotReady = "NotReady"
	// StatusDisabled 表示集群被后台禁用。
	StatusDisabled = "Disabled"
)

// Cluster 是平台纳管的 Kubernetes 集群配置。
type Cluster struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	AccessMode     string    `json:"access_mode"`
	APIServer      string    `json:"api_server"`
	KubeconfigPath string    `json:"kubeconfig_path"`
	PrometheusURL  string    `json:"prometheus_url"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Record 是集群配置数据库表模型。
type Record struct {
	ID             string    `gorm:"column:id;type:varchar(64);primaryKey"`
	Name           string    `gorm:"column:name;type:varchar(128);not null"`
	AccessMode     string    `gorm:"column:access_mode;type:varchar(32)"`
	APIServer      string    `gorm:"column:api_server;type:varchar(255)"`
	KubeconfigPath string    `gorm:"column:kubeconfig_path;type:varchar(255)"`
	PrometheusURL  string    `gorm:"column:prometheus_url;type:varchar(255)"`
	Description    string    `gorm:"column:description;type:varchar(512)"`
	Status         string    `gorm:"column:status;type:varchar(32)"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Record) TableName() string { return "clusters" }

func (r Record) ToDTO() Cluster {
	return Cluster{
		ID:             r.ID,
		Name:           r.Name,
		AccessMode:     r.AccessMode,
		APIServer:      r.APIServer,
		KubeconfigPath: r.KubeconfigPath,
		PrometheusURL:  r.PrometheusURL,
		Description:    r.Description,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

func newRecord(cluster Cluster) Record {
	return Record{
		ID:             cluster.ID,
		Name:           cluster.Name,
		AccessMode:     cluster.AccessMode,
		APIServer:      cluster.APIServer,
		KubeconfigPath: cluster.KubeconfigPath,
		PrometheusURL:  cluster.PrometheusURL,
		Description:    cluster.Description,
		Status:         cluster.Status,
	}
}
