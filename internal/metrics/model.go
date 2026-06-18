package metrics

import (
	"time"

	"compute-monitor-api/internal/prometheus"
)

// PointRecord 是 Prometheus 指标点缓存表。
type PointRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	ClusterID string    `gorm:"column:cluster_id;type:varchar(64);uniqueIndex:uk_metric_point"`
	NodeName  string    `gorm:"column:node_name;type:varchar(128);uniqueIndex:uk_metric_point"`
	Metric    string    `gorm:"column:metric;type:varchar(128);uniqueIndex:uk_metric_point"`
	Timestamp int64     `gorm:"column:timestamp;uniqueIndex:uk_metric_point"`
	Value     float64   `gorm:"column:value"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定指标点表名。
func (PointRecord) TableName() string {
	return "metric_points"
}

func NewPointRecord(clusterID string, nodeName string, metric string, item prometheus.Point) PointRecord {
	return PointRecord{
		ClusterID: clusterID,
		NodeName:  nodeName,
		Metric:    metric,
		Timestamp: item.Timestamp,
		Value:     item.Value,
	}
}

func NewPointRecords(clusterID string, nodeName string, metric string, points []prometheus.Point) []PointRecord {
	records := make([]PointRecord, 0, len(points))
	for _, item := range points {
		records = append(records, NewPointRecord(clusterID, nodeName, metric, item))
	}
	return records
}

func (r PointRecord) ToDTO() prometheus.Point {
	return prometheus.Point{Timestamp: r.Timestamp, Value: r.Value}
}
