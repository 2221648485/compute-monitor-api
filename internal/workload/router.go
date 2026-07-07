package workload

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册工作负载查询和操作路由。
func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.GET("/clusters/:clusterId/pods", handler.ListPods)
	api.GET("/clusters/:clusterId/pods/:namespace/:name", handler.GetPod)
	api.GET("/clusters/:clusterId/pods/:namespace/:name/logs", handler.PodLogs)
	api.GET("/clusters/:clusterId/deployments", handler.ListDeployments)
	api.GET("/clusters/:clusterId/services", handler.ListServices)
	api.POST("/clusters/:clusterId/workloads/yaml", handler.ApplyYAML)
	api.DELETE("/clusters/:clusterId/workloads/deployments/:namespace/:name", handler.DeleteDeployment)
	api.PUT("/clusters/:clusterId/workloads/deployments/:namespace/:name/scale", handler.ScaleDeployment)
}
