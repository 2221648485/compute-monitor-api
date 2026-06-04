package compat

import "github.com/gin-gonic/gin"

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
}
