package node

import "github.com/gin-gonic/gin"

func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.GET("/clusters/:clusterId/namespaces", handler.ListNamespaces)
	api.GET("/clusters/:clusterId/nodes", handler.ListNodes)
	api.GET("/clusters/:clusterId/nodes/:nodeName", handler.GetNode)
	api.GET("/clusters/:clusterId/nodes/:nodeName/pods", handler.ListNodePods)
}
