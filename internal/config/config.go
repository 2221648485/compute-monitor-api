package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const defaultEnv = "dev"

// Config 是应用启动时加载的总配置。
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Database   DatabaseConfig   `mapstructure:"mysql"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Auth       AuthConfig       `mapstructure:"auth"`
	K8s        K8sConfig        `mapstructure:"k8s"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Scheduler  SchedulerConfig  `mapstructure:"scheduler"`
}

// AppConfig 是 HTTP 服务自身配置。
type AppConfig struct {
	Env  string `mapstructure:"env"`
	Port string `mapstructure:"port"`
}

// DatabaseConfig 是 MySQL 连接池配置。
type DatabaseConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// RedisConfig 是 Redis 连接配置。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AuthConfig 是认证相关配置。
type AuthConfig struct {
	JWT            JWTConfig            `mapstructure:"jwt"`
	BootstrapAdmin BootstrapAdminConfig `mapstructure:"bootstrap_admin"`
}

// JWTConfig 是 JWT 签发配置。
type JWTConfig struct {
	Secret                 string `mapstructure:"secret"`
	Issuer                 string `mapstructure:"issuer"`
	AccessTokenTTLSeconds  int64  `mapstructure:"access_token_ttl_seconds"`
	RefreshTokenTTLSeconds int64  `mapstructure:"refresh_token_ttl_seconds"`
}

// BootstrapAdminConfig 是开发环境初始化管理员账号配置。
type BootstrapAdminConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	DisplayName string `mapstructure:"display_name"`
	Email       string `mapstructure:"email"`
	Role        string `mapstructure:"role"`
	Status      int    `mapstructure:"status"`
}

// K8sConfig 是 Kubernetes API 连接配置。
type K8sConfig struct {
	Mode           string `mapstructure:"mode"`
	ApiServer      string `mapstructure:"api_server"`
	KubeconfigPath string `mapstructure:"kubeconfig_path"`
}

// PrometheusConfig 是 Prometheus HTTP API 连接配置。
type PrometheusConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

// SchedulerConfig 是后台定时任务配置。
type SchedulerConfig struct {
	K8sSync K8sSyncSchedulerConfig `mapstructure:"k8s_sync"`
}

// K8sSyncSchedulerConfig 是 Kubernetes 资源同步任务配置。
type K8sSyncSchedulerConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	IntervalSeconds int64  `mapstructure:"interval_seconds"`
	TimeoutSeconds  int64  `mapstructure:"timeout_seconds"`
	Namespace       string `mapstructure:"namespace"`
}

// Load 根据 APP_ENV 或 GO_ENV 加载对应配置文件。
func Load() Config {
	cfg, err := LoadFromFile(ConfigFilePath(Env()))
	if err != nil {
		return defaultConfig()
	}
	return cfg
}

// Env 返回当前运行环境，默认 dev。
func Env() string {
	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if env == "" {
		env = strings.TrimSpace(os.Getenv("GO_ENV"))
	}
	if env == "" {
		return defaultEnv
	}
	return strings.ToLower(env)
}

// ConfigFilePath 根据环境名生成配置文件路径。
func ConfigFilePath(env string) string {
	env = strings.TrimSpace(strings.ToLower(env))
	if env == "" {
		env = defaultEnv
	}
	return fmt.Sprintf("configs/config.%s.yaml", env)
}

// LoadFromFile 从指定 yaml 文件加载配置。
func LoadFromFile(path string) (Config, error) {
	v := newViper()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !os.IsNotExist(err) {
			return Config{}, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func newViper() *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")

	v.SetDefault("app.env", defaultEnv)
	v.SetDefault("app.port", "8080")
	v.SetDefault("mysql.dsn", "root:123456@tcp(127.0.0.1:3306)/compute_monitor?charset=utf8mb4&parseTime=True&loc=Local")
	v.SetDefault("mysql.max_open_conns", 20)
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("auth.jwt.secret", "change-me")
	v.SetDefault("auth.jwt.issuer", "compute-monitor-api")
	v.SetDefault("auth.jwt.access_token_ttl_seconds", 7200)
	v.SetDefault("auth.jwt.refresh_token_ttl_seconds", 604800)
	v.SetDefault("auth.bootstrap_admin.enabled", true)
	v.SetDefault("auth.bootstrap_admin.username", "admin")
	v.SetDefault("auth.bootstrap_admin.password", "Admin@123456")
	v.SetDefault("auth.bootstrap_admin.display_name", "System Admin")
	v.SetDefault("auth.bootstrap_admin.email", "admin@example.com")
	v.SetDefault("auth.bootstrap_admin.role", "admin")
	v.SetDefault("auth.bootstrap_admin.status", 1)
	v.SetDefault("k8s.mode", "kubeconfig")
	v.SetDefault("k8s.api_server", "https://127.0.0.1:6443")
	v.SetDefault("k8s.kubeconfig_path", "configs/kubeconfig-demo.yaml")
	v.SetDefault("prometheus.base_url", "http://127.0.0.1:9090")
	v.SetDefault("scheduler.k8s_sync.enabled", true)
	v.SetDefault("scheduler.k8s_sync.interval_seconds", 60)
	v.SetDefault("scheduler.k8s_sync.timeout_seconds", 30)
	v.SetDefault("scheduler.k8s_sync.namespace", "")

	return v
}

func defaultConfig() Config {
	return Config{
		App: AppConfig{Env: defaultEnv, Port: "8080"},
		Database: DatabaseConfig{
			DSN:          "root:123456@tcp(127.0.0.1:3306)/compute_monitor?charset=utf8mb4&parseTime=True&loc=Local",
			MaxOpenConns: 20,
			MaxIdleConns: 10,
		},
		Redis: RedisConfig{Addr: "127.0.0.1:6379", DB: 0},
		Auth: AuthConfig{
			JWT: JWTConfig{Secret: "change-me", Issuer: "compute-monitor-api", AccessTokenTTLSeconds: 7200, RefreshTokenTTLSeconds: 604800},
			BootstrapAdmin: BootstrapAdminConfig{
				Enabled:     true,
				Username:    "admin",
				Password:    "Admin@123456",
				DisplayName: "System Admin",
				Email:       "admin@example.com",
				Role:        "admin",
				Status:      1,
			},
		},
		K8s:        K8sConfig{Mode: "kubeconfig", ApiServer: "https://127.0.0.1:6443", KubeconfigPath: "configs/kubeconfig-demo.yaml"},
		Prometheus: PrometheusConfig{BaseURL: "http://127.0.0.1:9090"},
		Scheduler: SchedulerConfig{
			K8sSync: K8sSyncSchedulerConfig{
				Enabled:         true,
				IntervalSeconds: 60,
				TimeoutSeconds:  30,
			},
		},
	}
}
