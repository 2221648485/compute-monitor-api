package prometheus

import (
	"context"
	"fmt"
	"time"

	"compute-monitor-api/internal/config"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Client 是业务模块访问 Prometheus 查询 API 的统一入口。
type Client interface {
	QueryRange(ctx context.Context, query string, start int64, end int64, step int64) ([]Point, error)
	BaseURL() string
}

// ClientFactory 根据 clusterId 创建对应集群的 Prometheus 客户端。
type ClientFactory interface {
	ForCluster(ctx context.Context, clusterID string) (Client, error)
}

// Point 是 Prometheus 区间查询返回的单个时间点。
type Point struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// HTTPClient 基于 Prometheus 官方 Go client 访问 Prometheus HTTP API。
type HTTPClient struct {
	baseURL string
	api     promv1.API
	initErr error
}

// NewClient 创建 Prometheus 客户端。
func NewClient(baseURL string) Client {
	apiClient, err := promapi.NewClient(promapi.Config{
		Address: baseURL,
	})
	if err != nil {
		return &HTTPClient{baseURL: baseURL, initErr: err}
	}

	return &HTTPClient{
		baseURL: baseURL,
		api:     promv1.NewAPI(apiClient),
	}
}

// NewClientFromConfig 根据配置创建 Prometheus 客户端。
func NewClientFromConfig(cfg config.PrometheusConfig) Client {
	return NewClient(cfg.BaseURL)
}

// BaseURL 返回当前配置的 Prometheus 地址。
func (c *HTTPClient) BaseURL() string {
	return c.baseURL
}

// QueryRange 调用 Prometheus 官方 QueryRange API 查询一段时间内的指标序列。
func (c *HTTPClient) QueryRange(ctx context.Context, query string, start int64, end int64, step int64) ([]Point, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("prometheus base url is empty")
	}
	if c.initErr != nil {
		return nil, fmt.Errorf("create prometheus client: %w", c.initErr)
	}
	if step <= 0 {
		step = 60
	}
	if end <= 0 {
		end = time.Now().Unix()
	}
	if start <= 0 {
		start = end - 3600
	}

	result, warnings, err := c.api.QueryRange(ctx, query, promv1.Range{
		Start: time.Unix(start, 0),
		End:   time.Unix(end, 0),
		Step:  time.Duration(step) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("query prometheus range: %w", err)
	}
	// Prometheus warning 不是失败；当前业务返回结构没有 warning 字段，先忽略。
	_ = warnings

	return valueToPoints(result)
}

func valueToPoints(value model.Value) ([]Point, error) {
	if value == nil {
		return []Point{}, nil
	}

	switch current := value.(type) {
	case model.Matrix:
		return matrixToPoints(current), nil
	case model.Vector:
		return vectorToPoints(current), nil
	case *model.Scalar:
		return []Point{{
			Timestamp: current.Timestamp.Time().Unix(),
			Value:     float64(current.Value),
		}}, nil
	default:
		return nil, fmt.Errorf("unsupported prometheus result type: %s", value.Type())
	}
}

func matrixToPoints(matrix model.Matrix) []Point {
	if len(matrix) == 0 {
		return []Point{}
	}

	// 当前业务接口返回单条曲线；如果一个 PromQL 返回多条曲线，先取第一条。
	stream := matrix[0]
	points := make([]Point, 0, len(stream.Values))
	for _, sample := range stream.Values {
		points = append(points, Point{
			Timestamp: sample.Timestamp.Time().Unix(),
			Value:     float64(sample.Value),
		})
	}
	return points
}

func vectorToPoints(vector model.Vector) []Point {
	if len(vector) == 0 {
		return []Point{}
	}

	points := make([]Point, 0, len(vector))
	for _, sample := range vector {
		points = append(points, Point{
			Timestamp: sample.Timestamp.Time().Unix(),
			Value:     float64(sample.Value),
		})
	}
	return points
}
