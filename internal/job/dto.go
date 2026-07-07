package job

// HealthResponse 是 Job 模块健康状态响应。
type HealthResponse struct {
	Module string `json:"module"`
	Status string `json:"status"`
}

type ExecuteMigrationTaskRequest struct {
	Namespace         string   `json:"namespace" binding:"required"`
	TargetYAML        string   `json:"targetYaml"`
	SourceDeployments []string `json:"sourceDeployments"`
	CleanupSource     bool     `json:"cleanupSource"`
	DryRun            bool     `json:"dryRun"`
}
