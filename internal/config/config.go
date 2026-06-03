package config

import (
	"os"
	"strconv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
}

func Load() Config {
	return Config{
		App: AppConfig{
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("MYSQL_DSN", "root:123456@tcp(127.0.0.1:3306)/compute_monitor?charset=utf8mb4&parseTime=True&loc=Local"),
		},
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func GetEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
