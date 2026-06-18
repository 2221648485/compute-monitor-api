package gpu

import "github.com/gin-gonic/gin"

func RegisterRoutes(api gin.IRoutes, handler Handler) {
	api.GET("/cluster/:clusterId/gpus", handler.List)
}
