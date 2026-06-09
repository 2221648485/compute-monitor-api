package app

import (
	"net/http"
	"time"

	_ "compute-monitor-api/docs/swagger"
	"compute-monitor-api/internal/auth"
	"compute-monitor-api/internal/compat"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/middleware"
	"compute-monitor-api/internal/response"
	"compute-monitor-api/internal/user"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// NewRouter 创建 HTTP 路由，并在这里完成各业务模块的依赖组装。
func NewRouter(cfg config.Config, db *gorm.DB) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.RequestLog())

	// 健康检查接口，可给 Docker 或 K8s 探针使用。
	router.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{
			"service": "compute-monitor-api",
			"status":  "ok",
		})
	})

	// 非生产环境开放 Swagger UI，生产环境通常由网关或内网文档平台统一管理。
	if cfg.App.Env != "prod" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// /api/v2 是调度器兼容接口，当前阶段先返回 mock 数据。
	compatAPI := router.Group("/api/v2")
	registerCompatModules(compatAPI)

	// /api/admin 是后台管理接口，包含认证、用户管理等后台能力。
	adminAPI := router.Group("/api/admin")
	registerAdminModules(adminAPI, cfg, db)

	return router
}

func registerCompatModules(api gin.IRouter) {
	compatHandler := compat.NewHandler()
	compat.RegisterRoutes(api, compatHandler)
}

func registerAdminModules(api gin.IRouter, cfg config.Config, db *gorm.DB) {
	if db == nil {
		api.Any("/*path", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, response.Body{
				Code:    http.StatusServiceUnavailable,
				Message: "database is not available",
			})
		})
		return
	}

	tokenManager := auth.NewTokenManager(auth.TokenOptions{
		Secret:         cfg.Auth.JWT.Secret,
		Issuer:         cfg.Auth.JWT.Issuer,
		AccessTokenTTL: time.Duration(cfg.Auth.JWT.AccessTokenTTLSeconds) * time.Second,
	})

	userRepository := user.NewMySQLRepository(db)
	passwordHasher := auth.NewPasswordHasher()

	authService := auth.NewService(userRepository, passwordHasher, tokenManager)
	authHandler := auth.NewHandler(authService)

	userService := user.NewService(userRepository, passwordHasher)
	userHandler := user.NewHandler(userService)

	// publicAPI 放不需要登录的接口，例如登录。
	publicAPI := api

	// privateAPI 挂载鉴权中间件，后续后台管理接口默认放在这里。
	privateAPI := api.Group("")
	privateAPI.Use(middleware.Auth(tokenManager))

	auth.RegisterRoutes(publicAPI, privateAPI, authHandler)
	user.RegisterRoutes(privateAPI, userHandler)
}
