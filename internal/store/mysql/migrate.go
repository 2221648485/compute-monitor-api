package mysql

import (
	"compute-monitor-api/internal/alert"
	"compute-monitor-api/internal/audit"
	"compute-monitor-api/internal/cluster"
	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/metrics"
	"compute-monitor-api/internal/migration"
	userpkg "compute-monitor-api/internal/user"

	"gorm.io/gorm"
)

// Migrate 创建或补齐当前应用需要的 MySQL 表结构。
// 生产项目后期可以替换为 goose、Atlas、Liquibase 这类版本化迁移工具。
func Migrate(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	return db.AutoMigrate(
		&userpkg.User{},
		&userpkg.LoginLog{},
		&cluster.Record{},
		&k8s.NamespaceRecord{},
		&k8s.NodeRecord{},
		&k8s.PodRecord{},
		&k8s.DeploymentRecord{},
		&k8s.ServiceRecord{},
		&metrics.PointRecord{},
		&alert.RuleRecord{},
		&alert.EventRecord{},
		&audit.LogRecord{},
		&migration.PlanRecord{},
		&migration.TaskRecord{},
		&migration.StepRecord{},
	)
}
