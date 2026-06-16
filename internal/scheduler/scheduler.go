package scheduler

import (
	"context"
	"log"
	"time"

	"compute-monitor-api/internal/cluster"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/k8ssync"
	"compute-monitor-api/internal/page"
)

const defaultClusterBatchSize = 100

// Scheduler 负责启动和管理后台定时任务。
// 当前只包含 Kubernetes 资源同步任务，后续可以继续加入告警计算、指标聚合等任务。
type Scheduler struct {
	clusterRepository cluster.Repository
	k8sSyncService    *k8ssync.Service
	k8sSyncConfig     config.K8sSyncSchedulerConfig
}

// NewScheduler 创建后台任务调度器。
func NewScheduler(clusterRepository cluster.Repository, k8sSyncService *k8ssync.Service, cfg config.SchedulerConfig) *Scheduler {
	return &Scheduler{
		clusterRepository: clusterRepository,
		k8sSyncService:    k8sSyncService,
		k8sSyncConfig:     normalizeK8sSyncConfig(cfg.K8sSync),
	}
}

// Start 启动后台任务。调用方需要传入可取消的 ctx，用于应用关闭时停止任务。
func (s *Scheduler) Start(ctx context.Context) {
	if !s.k8sSyncConfig.Enabled {
		log.Println("scheduler k8s sync disabled")
		return
	}

	go s.runK8sSyncLoop(ctx)
}

func (s *Scheduler) runK8sSyncLoop(ctx context.Context) {
	interval := time.Duration(s.k8sSyncConfig.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("scheduler k8s sync started: interval=%s", interval)
	s.syncRunningClusters(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler k8s sync stopped")
			return
		case <-ticker.C:
			s.syncRunningClusters(ctx)
		}
	}
}

func (s *Scheduler) syncRunningClusters(ctx context.Context) {
	timeout := time.Duration(s.k8sSyncConfig.TimeoutSeconds) * time.Second
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	clusters, err := s.listRunningClusters(runCtx)
	if err != nil {
		log.Printf("scheduler k8s sync list clusters failed: error=%v", err)
		return
	}
	if len(clusters) == 0 {
		return
	}

	for _, item := range clusters {
		select {
		case <-runCtx.Done():
			log.Printf("scheduler k8s sync canceled: clusterID=%s error=%v", item.ID, runCtx.Err())
			return
		default:
		}

		result, err := s.k8sSyncService.SyncAll(runCtx, item.ID, s.k8sSyncConfig.Namespace)
		if err != nil {
			log.Printf("scheduler k8s sync cluster failed: clusterID=%s error=%v", item.ID, err)
			continue
		}
		log.Printf(
			"scheduler k8s sync cluster done: clusterID=%s namespaces=%d nodes=%d pods=%d deployments=%d services=%d",
			result.ClusterID,
			result.Namespaces,
			result.Nodes,
			result.Pods,
			result.Deployments,
			result.Services,
		)
	}
}

func (s *Scheduler) listRunningClusters(ctx context.Context) ([]cluster.Cluster, error) {
	var result []cluster.Cluster

	for currentPage := 1; ; currentPage++ {
		items, total, err := s.clusterRepository.List(ctx, cluster.ListQuery{
			Status: cluster.StatusRunning,
			Query: page.Query{
				Page: currentPage,
				Size: defaultClusterBatchSize,
			},
		})
		if err != nil {
			return nil, err
		}

		result = append(result, items...)
		if int64(len(result)) >= total || len(items) == 0 {
			return result, nil
		}
	}
}

func normalizeK8sSyncConfig(cfg config.K8sSyncSchedulerConfig) config.K8sSyncSchedulerConfig {
	if cfg.IntervalSeconds <= 0 {
		cfg.IntervalSeconds = 60
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 30
	}
	return cfg
}
