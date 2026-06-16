package k8ssync

import "github.com/gin-gonic/gin"

func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.POST("/clusters/:clusterId/k8s/sync", handler.SyncAll)
	api.POST("/clusters/:clusterId/k8s/sync/namespaces", handler.SyncNamespaces)
	api.POST("/clusters/:clusterId/k8s/sync/nodes", handler.SyncNodes)
	api.POST("/clusters/:clusterId/k8s/sync/pods", handler.SyncPods)
	api.POST("/clusters/:clusterId/k8s/sync/deployments", handler.SyncDeployments)
	api.POST("/clusters/:clusterId/k8s/sync/services", handler.SyncServices)
}
