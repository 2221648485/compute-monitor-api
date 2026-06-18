package metrics

import "compute-monitor-api/internal/prometheus"

// Series 是指标接口统一返回结构。
type Series struct {
	NodeName string             `json:"nodeName"`
	Metric   string             `json:"metric"`
	Unit     string             `json:"unit"`
	Source   string             `json:"source"`
	Values   []prometheus.Point `json:"values"`
}
