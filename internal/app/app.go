package app

import (
	"compute-monitor-api/internal/compat"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/middleware"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter builds the HTTP router and wires module dependencies.
func NewRouter(cfg config.Config, db *gorm.DB) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.RequestLog())

	// 健康检查用于确认服务是否启动成功，后续也可以给 Docker/K8s 探针使用。
	router.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{
			"service": "compute-monitor-api",
			"status":  "ok",
		})
	})

	// /api/v2 是调度器兼容层，第一阶段先返回 mock，后续逐步替换为真实 K8s/Prometheus 数据。
	compatAPI := router.Group("/api/v2")
	registerCompatModules(compatAPI)

	// /api/admin 是 ComputeMonitor 自己的管理接口，供前端和 Postman 调试使用。
	adminAPI := router.Group("/api/admin")
	registerAdminModules(adminAPI, cfg, db)

	return router
}

func registerCompatModules(api gin.IRouter) {
	// compat 模块目前不直接依赖数据库；真实数据会通过内部 service 注入进来。
	compatHandler := compat.NewHandler()
	compat.RegisterRoutes(api, compatHandler)
}

func registerAdminModules(api gin.IRouter, cfg config.Config, db *gorm.DB) {
	// 阶段 0-1 暂不注册 admin 业务模块，保留参数避免后续装配时改动 NewRouter 签名。
	_ = api
	_ = cfg
	_ = db
}
