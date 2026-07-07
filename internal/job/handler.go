package job

import (
	"compute-monitor-api/internal/migration"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	migrationService *migration.Service
}

func NewHandler(migrationService *migration.Service) *Handler {
	return &Handler{migrationService: migrationService}
}

// Health 返回 Job 模块状态。
func (h *Handler) Health(c *gin.Context) {
	response.OK(c, HealthResponse{Module: "job", Status: "ready"})
}

// ExecuteMigrationTask 作为 Job 模块的任务入口触发迁移执行。
func (h *Handler) ExecuteMigrationTask(c *gin.Context) {
	var req ExecuteMigrationTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.migrationService.ExecuteTask(c.Request.Context(), c.Param("taskId"), migration.ExecuteTaskRequest{
		Namespace:         req.Namespace,
		TargetYAML:        req.TargetYAML,
		SourceDeployments: req.SourceDeployments,
		CleanupSource:     req.CleanupSource,
		DryRun:            req.DryRun,
	})
	if err != nil {
		response.BadRequest(c, migration.ErrorMessage(err))
		return
	}
	response.OK(c, result)
}
