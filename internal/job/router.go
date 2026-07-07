package job

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 Job 模块路由。
func RegisterRoutes(api gin.IRouter, handler *Handler) {
	group := api.Group("/jobs")
	group.GET("/health", handler.Health)
	group.POST("/migration/tasks/:taskId/execute", handler.ExecuteMigrationTask)
}
