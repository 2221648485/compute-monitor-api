package audit

import "time"

// LogRecord 是审计日志表。
type LogRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClusterID string    `gorm:"column:cluster_id;type:varchar(64)" json:"clusterId"`
	Action    string    `gorm:"column:action;type:varchar(128)" json:"action"`
	Operator  string    `gorm:"column:operator;type:varchar(64)" json:"operator"`
	Resource  string    `gorm:"column:resource;type:varchar(255)" json:"resource"`
	Detail    string    `gorm:"column:detail;type:text" json:"detail"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName 指定审计日志表名。
func (LogRecord) TableName() string {
	return "audit_logs"
}
