package student

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	group := router.Group("/students")
	{
		group.POST("", handler.Create)
		group.GET("/:id", handler.GetByID)
	}
}
