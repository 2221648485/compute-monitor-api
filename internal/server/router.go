package server

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册平台级辅助路由。
func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.GET("/compat/endpoints/status", handler.CompatEndpointStatus)
}
