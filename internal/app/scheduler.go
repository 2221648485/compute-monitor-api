package app

import (
	"compute-monitor-api/internal/cluster"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/k8ssync"
	"compute-monitor-api/internal/scheduler"
	"context"
	"log"
)

// StartSchedulers 组装并启动后台定时任务。
// 返回 cancel 函数，应用退出时调用它停止后台 goroutine。
func StartSchedulers(parent context.Context, cfg config.Config, resources Resources) context.CancelFunc {
	ctx, cancel := context.WithCancel(parent)
	if resources.DB == nil {
		log.Println("app scheduler skipped: database is not available")
		return cancel
	}

	clusterRepository := cluster.NewRepository(resources.DB)
	k8sRepository := k8s.NewRepository(resources.DB)
	k8sFactory := newClusterK8sClientFactory(clusterRepository, cfg.K8s)
	k8sSyncService := k8ssync.NewService(k8sFactory, k8sRepository)

	s := scheduler.NewScheduler(clusterRepository, k8sSyncService, cfg.Scheduler)
	s.Start(ctx)
	return cancel

}
