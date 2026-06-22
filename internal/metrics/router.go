package metrics

import "github.com/gin-gonic/gin"

func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.GET("/clusters/:clusterId/metrics/node/:nodeName/cpu", handler.CPU)
	api.GET("/clusters/:clusterId/metrics/node/:nodeName/memory", handler.Memory)
	api.GET("/clusters/:clusterId/metrics/node/:nodeName/disk", handler.Disk)
	api.GET("/clusters/:clusterId/metrics/node/:nodeName/network", handler.Network)
}
