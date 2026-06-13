package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册认证路由。
func RegisterRoutes(publicAPI gin.IRouter, privateAPI gin.IRouter, handler *Handler) {
	publicAPI.POST("/auth/login", handler.Login)
	publicAPI.POST("/auth/refresh", handler.RefreshToken)

	privateAPI.GET("/auth/me", handler.GetCurrentUser)
	privateAPI.PUT("/auth/password", handler.ChangePassword)
}
