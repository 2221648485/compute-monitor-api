package config

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"mysql"`
}

type AppConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

func Load() Config {
	cfg, err := LoadFromFile("configs/config.yaml")
	if err != nil {
		// 启动阶段不因为配置文件缺失直接崩溃；保底配置方便本地先跑通服务。
		return defaultConfig()
	}

	return cfg
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

	// 默认值用于本地快速启动；正式部署建议在 configs/config.yaml 或环境变量中显式配置。
	v.SetDefault("app.port", "8080")
	v.SetDefault("mysql.dsn", "root:123456@tcp(127.0.0.1:3306)/compute_monitor?charset=utf8mb4&parseTime=True&loc=Local")
	v.SetDefault("mysql.max_open_conns", 20)
	v.SetDefault("mysql.max_idle_conns", 10)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	bindEnv(v, "app.port", "APP_PORT")
	bindEnv(v, "mysql.dsn", "MYSQL_DSN")
	bindEnv(v, "mysql.max_open_conns", "MYSQL_MAX_OPEN_CONNS")
	bindEnv(v, "mysql.max_idle_conns", "MYSQL_MAX_IDLE_CONNS")

	return v
}

func bindEnv(v *viper.Viper, key string, env string) {
	if err := v.BindEnv(key, env); err != nil {
		panic(err)
	}
}

func defaultConfig() Config {
	return Config{
		App: AppConfig{Port: "8080"},
		Database: DatabaseConfig{
			DSN:          "root:123456@tcp(127.0.0.1:3306)/compute_monitor?charset=utf8mb4&parseTime=True&loc=Local",
			MaxOpenConns: 20,
			MaxIdleConns: 10,
		},
	}
}
