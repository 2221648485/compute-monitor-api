package gpu

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 GPU 查询路由。
func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.GET("/clusters/:clusterId/gpus", handler.List)
	api.GET("/clusters/:clusterId/nodes/:nodeName/gpus", handler.ListByNode)
	api.GET("/clusters/:clusterId/gpus/summary", handler.Summary)
	api.GET("/clusters/:clusterId/gpus/top", handler.Top)
	api.GET("/clusters/:clusterId/metrics/node/:nodeName/gpu", handler.Metrics)
}
