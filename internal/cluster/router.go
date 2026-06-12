package cluster

import "github.com/gin-gonic/gin"

func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.POST("/clusters", handler.Create)
	api.GET("/clusters", handler.List)
	api.GET("/clusters/:clusterId", handler.Get)
	api.PUT("/clusters/:clusterId", handler.Update)
	api.POST("/clusters/:clusterId/test", handler.Test)
	api.DELETE("/clusters/:clusterId", handler.Delete)
}
