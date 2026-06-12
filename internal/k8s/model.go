package k8s

import "time"

// NamespaceRecord 是 Namespace 的数据库缓存表。
type NamespaceRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	ClusterID string    `gorm:"column:cluster_id;type:varchar(64);uniqueIndex:uk_cluster_namespace"`
	Name      string    `gorm:"column:name;type:varchar(128);uniqueIndex:uk_cluster_namespace"`
	Status    string    `gorm:"column:status;type:varchar(32)"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (NamespaceRecord) TableName() string { return "k8s_namespaces" }

// NodeRecord 是 Node 的数据库缓存表。
type NodeRecord struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement"`
	ClusterID           string    `gorm:"column:cluster_id;type:varchar(64);uniqueIndex:uk_cluster_node"`
	Name                string    `gorm:"column:name;type:varchar(128);uniqueIndex:uk_cluster_node"`
	InternalIP          string    `gorm:"column:internal_ip;type:varchar(64)"`
	Status              string    `gorm:"column:status;type:varchar(32)"`
	RolesJSON           string    `gorm:"column:roles_json;type:text"`
	CPUCapacity         int       `gorm:"column:cpu_capacity"`
	MemoryCapacityBytes int64     `gorm:"column:memory_capacity_bytes"`
	GPUCount            int       `gorm:"column:gpu_count"`
	OSImage             string    `gorm:"column:os_image;type:varchar(255)"`
	KernelVersion       string    `gorm:"column:kernel_version;type:varchar(128)"`
	ContainerRuntime    string    `gorm:"column:container_runtime;type:varchar(128)"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (NodeRecord) TableName() string { return "k8s_nodes" }

func (r NodeRecord) ToDTO() Node {
	return Node{Name: r.Name, InternalIP: r.InternalIP, Status: r.Status, Roles: parseStringSlice(r.RolesJSON), CPUCapacity: r.CPUCapacity, MemoryCapacityBytes: r.MemoryCapacityBytes, GPUCount: r.GPUCount, OSImage: r.OSImage, KernelVersion: r.KernelVersion, ContainerRuntime: r.ContainerRuntime}
}

// PodRecord 是 Pod 的数据库缓存表。
type PodRecord struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	ClusterID    string    `gorm:"column:cluster_id;type:varchar(64);uniqueIndex:uk_cluster_pod"`
	Namespace    string    `gorm:"column:namespace;type:varchar(128);uniqueIndex:uk_cluster_pod"`
	Name         string    `gorm:"column:name;type:varchar(128);uniqueIndex:uk_cluster_pod"`
	NodeName     string    `gorm:"column:node_name;type:varchar(128)"`
	Phase        string    `gorm:"column:phase;type:varchar(32)"`
	PodIP        string    `gorm:"column:pod_ip;type:varchar(64)"`
	HostIP       string    `gorm:"column:host_ip;type:varchar(64)"`
	RestartCount int       `gorm:"column:restart_count"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PodRecord) TableName() string { return "k8s_pods" }

func (r PodRecord) ToDTO() Pod {
	return Pod{Name: r.Name, Namespace: r.Namespace, NodeName: r.NodeName, Phase: r.Phase, PodIP: r.PodIP, HostIP: r.HostIP, RestartCount: r.RestartCount}
}

// DeploymentRecord 是 Deployment 的数据库缓存表。
type DeploymentRecord struct {
	ID                int64     `gorm:"primaryKey;autoIncrement"`
	ClusterID         string    `gorm:"column:cluster_id;type:varchar(64);uniqueIndex:uk_cluster_deployment"`
	Namespace         string    `gorm:"column:namespace;type:varchar(128);uniqueIndex:uk_cluster_deployment"`
	Name              string    `gorm:"column:name;type:varchar(128);uniqueIndex:uk_cluster_deployment"`
	Replicas          int       `gorm:"column:replicas"`
	ReadyReplicas     int       `gorm:"column:ready_replicas"`
	AvailableReplicas int       `gorm:"column:available_replicas"`
	LabelsJSON        string    `gorm:"column:labels_json;type:text"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (DeploymentRecord) TableName() string { return "k8s_deployments" }

func (r DeploymentRecord) ToDTO() Deployment {
	return Deployment{Name: r.Name, Namespace: r.Namespace, Replicas: r.Replicas, ReadyReplicas: r.ReadyReplicas, AvailableReplicas: r.AvailableReplicas, Labels: parseStringMap(r.LabelsJSON)}
}

// ServiceRecord 是 Kubernetes Service 的数据库缓存表。
type ServiceRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	ClusterID string    `gorm:"column:cluster_id;type:varchar(64);uniqueIndex:uk_cluster_service"`
	Namespace string    `gorm:"column:namespace;type:varchar(128);uniqueIndex:uk_cluster_service"`
	Name      string    `gorm:"column:name;type:varchar(128);uniqueIndex:uk_cluster_service"`
	Type      string    `gorm:"column:type;type:varchar(64)"`
	ClusterIP string    `gorm:"column:cluster_ip;type:varchar(64)"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ServiceRecord) TableName() string { return "k8s_services" }
