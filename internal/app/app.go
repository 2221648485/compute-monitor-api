package app

import (
	"net/http"

	_ "compute-monitor-api/docs/swagger"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/middleware"
	"compute-monitor-api/internal/response"
	"compute-monitor-api/internal/user"

	"github.com/gin-gonic/gin"
	redisclient "github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// NewRouter 创建 HTTP 路由入口。
// app 层负责全局中间件、基础路由和模块注册，不直接编写业务逻辑。
func NewRouter(cfg config.Config, db *gorm.DB, redisClient *redisclient.Client) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.RequestLog())

	registerHealthRoute(router)
	registerSwaggerRoute(router, cfg)
	registerModules(router, NewAppContext(cfg, db, redisClient))

	return router
}

// registerHealthRoute 注册服务探活接口，给 Docker、Kubernetes 或负载均衡使用。
func registerHealthRoute(router *gin.Engine) {
	router.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{
			"service": "compute-monitor-api",
			"status":  "ok",
		})
	})
}

// registerSwaggerRoute 非生产环境开放 Swagger UI。
func registerSwaggerRoute(router *gin.Engine, cfg config.Config) {
	if cfg.App.Env == "prod" {
		return
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// registerModules 创建顶层路由组，并把模块装配委托给 ModuleRegistry。
func registerModules(router *gin.Engine, appCtx *AppContext) {
	adminAPI := router.Group("/api/admin")
	compatAPI := router.Group("/api/v2")

	if !appCtx.Ready() {
		registerUnavailableRoutes(adminAPI, compatAPI)
		return
	}

	registry := NewModuleRegistry(appCtx)
	registry.EnsureBootstrapAdmin()
	registry.RegisterCompat(compatAPI)

	publicAPI := adminAPI
	privateAPI := adminAPI.Group("")
	privateAPI.Use(middleware.Auth(appCtx.TokenManager, appCtx.AuthService))

	adminOnlyAPI := privateAPI.Group("")
	adminOnlyAPI.Use(middleware.RequireRole(user.RoleAdmin))

	registry.RegisterAuth(publicAPI, privateAPI)
	registry.RegisterUser(adminOnlyAPI)
	registry.RegisterCluster(adminOnlyAPI)
	registry.RegisterK8sSync(adminOnlyAPI)
}

// registerUnavailableRoutes 在基础设施不可用时统一返回 503，避免启动期空指针 panic。
func registerUnavailableRoutes(routers ...gin.IRouter) {
	handler := func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, response.Body{
			Code:    http.StatusServiceUnavailable,
			Message: "database or redis is not available",
		})
	}
	for _, router := range routers {
		router.Any("/*path", handler)
	}
}
