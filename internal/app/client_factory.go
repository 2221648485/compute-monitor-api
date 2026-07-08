package app

import (
	"context"
	"fmt"
	"strings"

	"compute-monitor-api/internal/cluster"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/prometheus"
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
		KubeconfigPath: kubeconfigPath,
	}
	return k8s.NewClient(opts)
}

type clusterPrometheusClientFactory struct {
	clusterRepository cluster.Repository
}

func newClusterPrometheusClientFactory(clusterRepository cluster.Repository) *clusterPrometheusClientFactory {
	return &clusterPrometheusClientFactory{
		clusterRepository: clusterRepository,
	}
}

func (f *clusterPrometheusClientFactory) ForCluster(ctx context.Context, clusterID string) (prometheus.Client, error) {
	current, err := f.clusterRepository.Get(ctx, strings.TrimSpace(clusterID))
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimSpace(current.PrometheusURL)
	if baseURL == "" {
		return nil, fmt.Errorf("cluster %s has no prometheus url", clusterID)
	}
	return prometheus.NewClient(baseURL), nil
}
