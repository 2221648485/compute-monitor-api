package alert

import "time"

// RuleRecord 是告警规则表。
type RuleRecord struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClusterID       string    `gorm:"column:cluster_id;type:varchar(64)" json:"clusterId"`
	NodeName        string    `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	MetricType      string    `gorm:"column:metric_type;type:varchar(128)" json:"metricType"`
	Operator        string    `gorm:"column:operator;type:varchar(16)" json:"operator"`
	Threshold       float64   `gorm:"column:threshold" json:"threshold"`
	DurationSeconds int       `gorm:"column:duration_seconds" json:"durationSeconds"`
	Enabled         bool      `gorm:"column:enabled" json:"enabled"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// TableName 指定告警规则表名。
func (RuleRecord) TableName() string {
	return "alert_rules"
}

// EventRecord 是告警事件表。
type EventRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClusterID string    `gorm:"column:cluster_id;type:varchar(64)" json:"clusterId"`
	NodeName  string    `gorm:"column:node_name;type:varchar(128)" json:"nodeName"`
	Status    string    `gorm:"column:status;type:varchar(32)" json:"status"`
	Message   string    `gorm:"column:message;type:varchar(512)" json:"message"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName 指定告警事件表名。
func (EventRecord) TableName() string {
	return "alert_events"
}
