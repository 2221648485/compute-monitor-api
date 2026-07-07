package alert

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册告警接口。
func RegisterRoutes(api gin.IRouter, handler *Handler) {
	api.POST("/alerts/rules", handler.CreateRule)
	api.GET("/alerts/rules", handler.ListRules)
	api.PUT("/alerts/rules/:ruleId", handler.UpdateRule)
	api.DELETE("/alerts/rules/:ruleId", handler.DeleteRule)
	api.GET("/alerts/events", handler.ListEvents)
}
