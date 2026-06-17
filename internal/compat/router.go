package compat

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册调度器兼容接口。
func RegisterRoutes(router gin.IRouter, handler *Handler) {
	// P0 兼容接口：先保证 ser-cloud-adapter 能请求本地服务。
	router.GET("/clusters/exclusive/list", handler.ListClusters)
	router.GET("/clusters/:clusterId/summary/static", handler.StaticSummary)
	router.GET("/clusters/:clusterId/summary/dynamic", handler.DynamicSummary)
	router.GET("/clusters/:clusterId/nodes", handler.ListNodes)
	router.GET("/clusters/:clusterId/nodes/resource-consumption", handler.NodeResourceConsumption)
	router.GET("/clusters/:clusterId/apps", handler.ListApps)
	router.GET("/clusters/:clusterId/instances", handler.ListInstances)
	router.GET("/clusters/:clusterId/native/Deployment", handler.ListDeployments)
	router.GET("/clusters/:clusterId/metric/nodes/:nodeName/metrics", handler.NodeMetrics)

	// P1 兼容接口：通过 client-go 执行真实 K8s 操作。
	router.POST("/clusters/:clusterId/native", handler.CreateNativeResource)
	router.DELETE("/clusters/:clusterId/native/Deployment/:name", handler.DeleteDeployment)
	router.PUT("/clusters/:clusterId/apps/batch-start", handler.BatchStartApps)
	router.PUT("/clusters/:clusterId/apps/batch-stop", handler.BatchStopApps)
	router.POST("/clusters/:clusterId/apps/batch-delete", handler.BatchDeleteApps)
}
