package audit

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册审计接口。
func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.GET("/audit/logs", handler.List)
}
