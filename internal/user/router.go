package user

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册后台用户管理路由。
func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.GET("/users", handler.List)
	api.POST("/users", handler.Create)
	api.GET("/users/:userId", handler.GetByID)
	api.PUT("/users/:userId", handler.Update)
	api.PUT("/users/:userId/status", handler.UpdateStatus)
	api.PUT("/users/:userId/password", handler.ResetPassword)
}
