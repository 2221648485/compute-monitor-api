package cluster

import "time"

const (
	// StatusRunning 表示集群当前可正常连接和使用。
	StatusRunning = "Running"
	// StatusNotReady 表示集群暂时不可用或连接异常。
	StatusNotReady = "NotReady"
	// StatusDisabled 表示集群被后台禁用。
	StatusDisabled = "Disabled"
)

// Cluster 是平台管理的集群配置，对外作为接口返回模型使用。
type Cluster struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
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
	KubeconfigPath string    `gorm:"column:kubeconfig_path;type:varchar(255)"`
	PrometheusURL  string    `gorm:"column:prometheus_url;type:varchar(255)"`
	Description    string    `gorm:"column:description;type:varchar(512)"`
	Status         string    `gorm:"column:status;type:varchar(32)"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定集群配置表名。
func (Record) TableName() string { return "clusters" }

// ToDTO 把数据库模型转换成接口模型。
func (r Record) ToDTO() Cluster {
	return Cluster{
		ID:             r.ID,
		Name:           r.Name,
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
		KubeconfigPath: cluster.KubeconfigPath,
		PrometheusURL:  cluster.PrometheusURL,
		Description:    cluster.Description,
		Status:         cluster.Status,
	}
}
