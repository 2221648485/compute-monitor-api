package app

import (
	"context"
	"fmt"
	"strings"

	"compute-monitor-api/internal/cluster"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/k8s"
)

type clusterK8sClientFactory struct {
	clusterRepository cluster.Repository
	defaultConfig     config.K8sConfig
}

func newClusterK8sClientFactory(clusterRepository cluster.Repository, defaultConfig config.K8sConfig) *clusterK8sClientFactory {
	return &clusterK8sClientFactory{
		clusterRepository: clusterRepository,
		defaultConfig:     defaultConfig,
	}
}

func (f *clusterK8sClientFactory) ForCluster(ctx context.Context, clusterID string) (k8s.Client, error) {
	current, err := f.clusterRepository.Get(ctx, strings.TrimSpace(clusterID))
	if err != nil {
		return nil, err
	}

	kubeconfigPath := strings.TrimSpace(current.KubeconfigPath)
	if kubeconfigPath == "" {
		return nil, fmt.Errorf("cluster %s has no kubeconfig", clusterID)
	}

	opts := k8s.Options{
		Mode:           "kubeconfig",
		APIServer:      f.defaultConfig.ApiServer,
		KubeconfigPath: kubeconfigPath,
	}
	return k8s.NewClient(opts)
}
