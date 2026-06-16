package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const defaultEnv = "dev"

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Database   DatabaseConfig   `mapstructure:"mysql"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Auth       AuthConfig       `mapstructure:"auth"`
	K8s        K8sConfig        `mapstructure:"k8s"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
}

type AppConfig struct {
	Env  string `mapstructure:"env"`
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	JWT            JWTConfig            `mapstructure:"jwt"`
	BootstrapAdmin BootstrapAdminConfig `mapstructure:"bootstrap_admin"`
}

type JWTConfig struct {
	Secret                 string `mapstructure:"secret"`
	Issuer                 string `mapstructure:"issuer"`
	AccessTokenTTLSeconds  int64  `mapstructure:"access_token_ttl_seconds"`
	RefreshTokenTTLSeconds int64  `mapstructure:"refresh_token_ttl_seconds"`
}

type BootstrapAdminConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	DisplayName string `mapstructure:"display_name"`
	Email       string `mapstructure:"email"`
	Role        string `mapstructure:"role"`
	Status      int    `mapstructure:"status"`
}

type K8sConfig struct {
	Mode           string `mapstructure:"mode"`
	ApiServer      string `mapstructure:"api_server"`
	KubeconfigPath string `mapstructure:"kubeconfig_path"`
}

type PrometheusConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

func Load() Config {
	cfg, err := LoadFromFile(ConfigFilePath(Env()))
	if err != nil {
		return defaultConfig()
	}

	return cfg
}

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

func ConfigFilePath(env string) string {
	env = strings.TrimSpace(strings.ToLower(env))
	if env == "" {
		env = defaultEnv
	}

	return fmt.Sprintf("configs/config.%s.yaml", env)
}

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
	v.SetDefault("auth.bootstrap_admin.display_name", "系统管理员")
	v.SetDefault("auth.bootstrap_admin.email", "")
	v.SetDefault("auth.bootstrap_admin.role", "admin")
	v.SetDefault("auth.bootstrap_admin.status", 1)
	v.SetDefault("k8s.mode", "kubeconfig")
	v.SetDefault("k8s.api_server", "https://127.0.0.1:6443")
	v.SetDefault("k8s.kubeconfig_path", "configs/kubeconfig-demo.yaml")
	v.SetDefault("prometheus.base_url", "http://127.0.0.1:9090")

	return v
}

func defaultConfig() Config {
	return Config{
		App: AppConfig{
			Env:  defaultEnv,
			Port: "8080",
		},
		Database: DatabaseConfig{
			DSN:          "root:123456@tcp(127.0.0.1:3306)/compute_monitor?charset=utf8mb4&parseTime=True&loc=Local",
			MaxOpenConns: 20,
			MaxIdleConns: 10,
		},
		Redis: RedisConfig{
			Addr: "127.0.0.1:6379",
			DB:   0,
		},
		Auth: AuthConfig{
			JWT: JWTConfig{
				Secret:                 "change-me",
				Issuer:                 "compute-monitor-api",
				AccessTokenTTLSeconds:  7200,
				RefreshTokenTTLSeconds: 604800,
			},
			BootstrapAdmin: BootstrapAdminConfig{
				Enabled:     true,
				Username:    "admin",
				Password:    "Admin@123456",
				DisplayName: "系统管理员",
				Role:        "admin",
				Status:      1,
			},
		},
		K8s:        K8sConfig{Mode: "kubeconfig", ApiServer: "https://127.0.0.1:6443", KubeconfigPath: "configs/kubeconfig-demo.yaml"},
		Prometheus: PrometheusConfig{BaseURL: "http://127.0.0.1:9090"},
	}
}
