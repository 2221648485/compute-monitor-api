package app

import (
	"context"
	"log"
	"net/http"
	"time"

	_ "compute-monitor-api/docs/swagger"
	"compute-monitor-api/internal/auth"
	"compute-monitor-api/internal/cluster"
	"compute-monitor-api/internal/compat"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/middleware"
	"compute-monitor-api/internal/response"
	"compute-monitor-api/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// NewRouter 创建 HTTP 路由，并在这里完成各业务模块的依赖组装。
func NewRouter(cfg config.Config, db *gorm.DB, redisClient *redis.Client) *gin.Engine {
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

	// /api/v2 是调度器兼容接口，当前阶段先保留原有兼容模块。
	compatAPI := router.Group("/api/v2")
	registerCompatModules(compatAPI)

	// /api/admin 是后台管理接口，包含认证、用户管理、集群管理等后台能力。
	adminAPI := router.Group("/api/admin")
	registerAdminModules(adminAPI, cfg, db, redisClient)

	return router
}

func registerCompatModules(api gin.IRouter) {
	compatHandler := compat.NewHandler()
	compat.RegisterRoutes(api, compatHandler)
}

func registerAdminModules(api gin.IRouter, cfg config.Config, db *gorm.DB, redisClient *redis.Client) {
	if db == nil || redisClient == nil {
		api.Any("/*path", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, response.Body{
				Code:    http.StatusServiceUnavailable,
				Message: "database or redis is not available",
			})
		})
		return
	}

	ensureBootstrapAdmin(cfg, db)

	userRepository := user.NewMySQLRepository(db)
	passwordHasher := auth.NewPasswordHasher()
	tokenManager := newTokenManager(cfg)
	sessionStore := auth.NewRedisSessionStore(redisClient)
	authService := auth.NewService(userRepository, passwordHasher, tokenManager, sessionStore, refreshTokenTTL(cfg))

	// publicAPI 放不需要登录的接口，例如登录和 refresh token 换签。
	publicAPI := api

	// privateAPI 只负责“是否登录”，并校验 token 是否已因改密、禁用、角色变化而失效。
	privateAPI := api.Group("")
	privateAPI.Use(middleware.Auth(tokenManager, authService))

	// adminOnlyAPI 负责后台管理授权，普通登录用户不能管理用户和集群。
	adminOnlyAPI := privateAPI.Group("")
	adminOnlyAPI.Use(middleware.RequireRole(user.RoleAdmin))

	registerAuthModule(publicAPI, privateAPI, authService)
	registerUserModule(adminOnlyAPI, userRepository, passwordHasher)
	registerClusterModule(adminOnlyAPI, db)
}

func registerAuthModule(publicAPI gin.IRouter, privateAPI gin.IRouter, service *auth.Service) {
	handler := auth.NewHandler(service)
	auth.RegisterRoutes(publicAPI, privateAPI, handler)
}

func registerUserModule(api gin.IRouter, repository user.Repository, passwordHasher user.PasswordHasher) {
	service := user.NewService(repository, passwordHasher)
	handler := user.NewHandler(service)
	user.RegisterRoutes(api, handler)
}

func registerClusterModule(api gin.IRouter, db *gorm.DB) {
	repository := cluster.NewMySQLRepository(db)
	deleteRepository := cluster.NewMySQLDeleteRepository(db)
	service := cluster.NewService(repository, deleteRepository)
	handler := cluster.NewHandler(service)
	cluster.RegisterRoutes(api, handler)
}

func ensureBootstrapAdmin(cfg config.Config, db *gorm.DB) {
	repository := user.NewMySQLRepository(db)
	passwordHasher := auth.NewPasswordHasher()
	if err := user.EnsureBootstrapAdmin(context.Background(), repository, passwordHasher, user.BootstrapAdminOptions{
		Enabled:     cfg.Auth.BootstrapAdmin.Enabled,
		Username:    cfg.Auth.BootstrapAdmin.Username,
		Password:    cfg.Auth.BootstrapAdmin.Password,
		DisplayName: cfg.Auth.BootstrapAdmin.DisplayName,
		Email:       cfg.Auth.BootstrapAdmin.Email,
		Role:        cfg.Auth.BootstrapAdmin.Role,
		Status:      cfg.Auth.BootstrapAdmin.Status,
	}); err != nil {
		log.Printf("bootstrap admin skipped: %v", err)
	}
}

func newTokenManager(cfg config.Config) auth.TokenManager {
	return auth.NewTokenManager(auth.TokenOptions{
		Secret:          cfg.Auth.JWT.Secret,
		Issuer:          cfg.Auth.JWT.Issuer,
		AccessTokenTTL:  time.Duration(cfg.Auth.JWT.AccessTokenTTLSeconds) * time.Second,
		RefreshTokenTTL: time.Duration(cfg.Auth.JWT.RefreshTokenTTLSeconds) * time.Second,
	})
}

func refreshTokenTTL(cfg config.Config) time.Duration {
	ttl := time.Duration(cfg.Auth.JWT.RefreshTokenTTLSeconds) * time.Second
	if ttl <= 0 {
		return 7 * 24 * time.Hour
	}
	return ttl
}
