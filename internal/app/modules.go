package app

import (
	"context"
	"log"
	"time"

	"compute-monitor-api/internal/alert"
	"compute-monitor-api/internal/audit"
	"compute-monitor-api/internal/auth"
	"compute-monitor-api/internal/cluster"
	"compute-monitor-api/internal/compat"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/gpu"
	"compute-monitor-api/internal/job"
	"compute-monitor-api/internal/k8s"
	"compute-monitor-api/internal/k8ssync"
	"compute-monitor-api/internal/metrics"
	"compute-monitor-api/internal/migration"
	"compute-monitor-api/internal/node"
	"compute-monitor-api/internal/prometheus"
	"compute-monitor-api/internal/resource"
	"compute-monitor-api/internal/server"
	"compute-monitor-api/internal/user"
	"compute-monitor-api/internal/workload"

	"github.com/gin-gonic/gin"
	redisclient "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AppContext 保存 app 装配层需要复用的基础依赖和共享服务。
// 它不是业务上下文，只在启动阶段用于组装 repository、service、handler 和 router。
type AppContext struct {
	Config       config.Config
	DB           *gorm.DB
	Redis        *redisclient.Client
	TokenManager auth.TokenManager
	AuthService  *auth.Service
	K8sFactory   k8s.ClientFactory
	PromFactory  prometheus.ClientFactory
}

// NewAppContext 创建模块注册时使用的应用上下文。
func NewAppContext(cfg config.Config, db *gorm.DB, redisClient *redisclient.Client) *AppContext {
	userRepository := user.NewRepository(db)
	clusterRepository := cluster.NewRepository(db)
	passwordHasher := auth.NewPasswordHasher()
	tokenManager := newTokenManager(cfg)
	sessionStore := auth.NewRedisSessionStore(redisClient)

	return &AppContext{
		Config:       cfg,
		DB:           db,
		Redis:        redisClient,
		TokenManager: tokenManager,
		AuthService:  auth.NewService(userRepository, passwordHasher, tokenManager, sessionStore, refreshTokenTTL(cfg)),
		K8sFactory:   newClusterK8sClientFactory(clusterRepository, cfg.K8s),
		PromFactory:  newClusterPrometheusClientFactory(clusterRepository),
	}
}

// Ready 判断基础设施是否足够支撑业务接口运行。
func (c *AppContext) Ready() bool {
	return c != nil && c.DB != nil && c.Redis != nil
}

// ModuleRegistry 负责完成每个业务模块的依赖组装和路由注册。
type ModuleRegistry struct {
	ctx *AppContext
}

// NewModuleRegistry 创建模块注册器。
func NewModuleRegistry(ctx *AppContext) *ModuleRegistry {
	return &ModuleRegistry{ctx: ctx}
}

// EnsureBootstrapAdmin 根据配置初始化默认管理员账号。
func (r *ModuleRegistry) EnsureBootstrapAdmin() {
	repository := user.NewRepository(r.ctx.DB)
	passwordHasher := auth.NewPasswordHasher()
	cfg := r.ctx.Config.Auth.BootstrapAdmin

	if err := user.EnsureBootstrapAdmin(context.Background(), repository, passwordHasher, user.BootstrapAdminOptions{
		Enabled:     cfg.Enabled,
		Username:    cfg.Username,
		Password:    cfg.Password,
		DisplayName: cfg.DisplayName,
		Email:       cfg.Email,
		Role:        cfg.Role,
		Status:      cfg.Status,
	}); err != nil {
		log.Printf("app bootstrap admin skipped: error=%v", err)
	}
}

// RegisterCompat 注册兼容旧版前端的 /api/v2 接口。
func (r *ModuleRegistry) RegisterCompat(api gin.IRouter) {
	k8sRepository := k8s.NewRepository(r.ctx.DB)
	clusterRepository := cluster.NewRepository(r.ctx.DB)
	handler := compat.NewHandler(r.ctx.K8sFactory, k8sRepository, r.ctx.PromFactory, clusterRepository)
	compat.RegisterRoutes(api, handler)
}

// RegisterAuth 注册认证接口，包括登录、刷新 token 和退出登录。
func (r *ModuleRegistry) RegisterAuth(publicAPI gin.IRouter, privateAPI gin.IRouter) {
	handler := auth.NewHandler(r.ctx.AuthService)
	auth.RegisterRoutes(publicAPI, privateAPI, handler)
}

// RegisterUser 注册后台用户管理接口。
func (r *ModuleRegistry) RegisterUser(api gin.IRouter) {
	repository := user.NewRepository(r.ctx.DB)
	passwordHasher := auth.NewPasswordHasher()
	service := user.NewService(repository, passwordHasher)
	handler := user.NewHandler(service)
	user.RegisterRoutes(api, handler)
}

// RegisterCluster 注册集群配置管理接口。
func (r *ModuleRegistry) RegisterCluster(api gin.IRouter) {
	repository := cluster.NewRepository(r.ctx.DB)
	deleteRepository := cluster.NewDeleteRepository(r.ctx.DB)
	service := cluster.NewServiceWithConfig(repository, deleteRepository, r.ctx.K8sFactory, r.ctx.Config.K8s)
	handler := cluster.NewHandler(service)
	cluster.RegisterRoutes(api, handler)
}

// RegisterK8sSync 注册 Kubernetes 数据同步接口。
func (r *ModuleRegistry) RegisterK8sSync(api gin.IRouter) {
	repository := k8s.NewRepository(r.ctx.DB)
	service := k8ssync.NewService(r.ctx.K8sFactory, repository)
	handler := k8ssync.NewHandler(service)
	k8ssync.RegisterRoutes(api, handler)
}

// RegisterNode 注册节点查询接口，数据从数据库缓存读取。
func (r *ModuleRegistry) RegisterNode(api gin.IRouter) {
	repository := k8s.NewRepository(r.ctx.DB)
	service := node.NewService(repository)
	handler := node.NewHandler(service)
	node.RegisterRoutes(api, handler)
}

// RegisterWorkload 注册工作负载接口。
func (r *ModuleRegistry) RegisterWorkload(api gin.IRouter) {
	repository := k8s.NewRepository(r.ctx.DB)
	service := workload.NewService(r.ctx.K8sFactory, repository)
	handler := workload.NewHandler(service)
	workload.RegisterRoutes(api, handler)
}

// RegisterMetrics 注册监控指标接口。
func (r *ModuleRegistry) RegisterMetrics(api gin.IRouter) {
	repository := metrics.NewRepository(r.ctx.DB)
	service := metrics.NewService(r.ctx.PromFactory, repository)
	handler := metrics.NewHandler(service)
	metrics.RegisterRoutes(api, handler)
}

// RegisterGPU 注册 GPU 查询接口。
func (r *ModuleRegistry) RegisterGPU(api gin.IRouter) {
	repository := k8s.NewRepository(r.ctx.DB)
	service := gpu.NewService(repository, r.ctx.PromFactory)
	handler := gpu.NewHandler(service)
	gpu.RegisterRoutes(api, handler)
}

// RegisterResource 注册算力、网络、存储和能源等多维资源感知接口。
func (r *ModuleRegistry) RegisterResource(api gin.IRouter) {
	clusterRepository := cluster.NewRepository(r.ctx.DB)
	k8sRepository := k8s.NewRepository(r.ctx.DB)
	service := resource.NewService(clusterRepository, k8sRepository)
	handler := resource.NewHandler(service)
	resource.RegisterRoutes(api, handler)
}

// RegisterMigration 注册迁移计划、迁移决策和迁移任务生命周期接口。
func (r *ModuleRegistry) RegisterMigration(api gin.IRouter) {
	clusterRepository := cluster.NewRepository(r.ctx.DB)
	k8sRepository := k8s.NewRepository(r.ctx.DB)
	resourceService := resource.NewService(clusterRepository, k8sRepository)
	repository := migration.NewRepository(r.ctx.DB)
	service := migration.NewService(repository, r.ctx.K8sFactory, k8sRepository, resourceService)
	handler := migration.NewHandler(service)
	migration.RegisterRoutes(api, handler)
}

// RegisterAlert 注册告警规则和告警事件接口。
func (r *ModuleRegistry) RegisterAlert(api gin.IRouter) {
	repository := alert.NewRepository(r.ctx.DB)
	service := alert.NewService(repository)
	handler := alert.NewHandler(service)
	alert.RegisterRoutes(api, handler)
}

// RegisterAudit 注册审计日志接口。
func (r *ModuleRegistry) RegisterAudit(api gin.IRouter) {
	repository := audit.NewRepository(r.ctx.DB)
	service := audit.NewService(repository)
	handler := audit.NewHandler(service)
	audit.RegisterRoutes(api, handler)
}

// RegisterJob 注册任务接口。
func (r *ModuleRegistry) RegisterJob(api gin.IRouter) {
	clusterRepository := cluster.NewRepository(r.ctx.DB)
	k8sRepository := k8s.NewRepository(r.ctx.DB)
	resourceService := resource.NewService(clusterRepository, k8sRepository)
	migrationRepository := migration.NewRepository(r.ctx.DB)
	migrationService := migration.NewService(migrationRepository, r.ctx.K8sFactory, k8sRepository, resourceService)
	job.RegisterRoutes(api, job.NewHandler(migrationService))
}

// RegisterServer 注册服务器概览接口。
func (r *ModuleRegistry) RegisterServer(api gin.IRouter) {
	server.RegisterRoutes(api, server.NewHandler())
}

// newTokenManager 创建 JWT 管理器，负责 access token 和 refresh token 的签发与解析。
func newTokenManager(cfg config.Config) auth.TokenManager {
	return auth.NewTokenManager(auth.TokenOptions{
		Secret:          cfg.Auth.JWT.Secret,
		Issuer:          cfg.Auth.JWT.Issuer,
		AccessTokenTTL:  time.Duration(cfg.Auth.JWT.AccessTokenTTLSeconds) * time.Second,
		RefreshTokenTTL: time.Duration(cfg.Auth.JWT.RefreshTokenTTLSeconds) * time.Second,
	})
}

// refreshTokenTTL 返回 refresh token 会话在 Redis 中的保存时间。
func refreshTokenTTL(cfg config.Config) time.Duration {
	ttl := time.Duration(cfg.Auth.JWT.RefreshTokenTTLSeconds) * time.Second
	if ttl <= 0 {
		return 7 * 24 * time.Hour
	}
	return ttl
}
