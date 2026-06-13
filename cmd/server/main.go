package main

import (
	"context"
	"log"
	"time"

	"compute-monitor-api/internal/app"
	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/store/mysql"
	redisstore "compute-monitor-api/internal/store/redis"
)

// @title Compute Monitor API
// @version 1.0
// @description 算力监控平台后端 API 文档。
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.Load()

	db, err := mysql.New(mysql.Options{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: 30 * time.Minute,
	})
	if err != nil {
		log.Printf("mysql init failed, database api will not be available: %v", err)
	}
	if db != nil {
		if err := mysql.Migrate(db); err != nil {
			log.Fatalf("mysql migrate failed: %v", err)
		}
		defer func() {
			if err := mysql.Close(db); err != nil {
				log.Printf("mysql close failed: %v", err)
			}
		}()
	}

	redisClient, err := redisstore.New(context.Background(), redisstore.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Printf("redis init failed, auth api will not be available: %v", err)
	}
	if redisClient != nil {
		defer func() {
			if err := redisstore.Close(redisClient); err != nil {
				log.Printf("redis close failed: %v", err)
			}
		}()
	}

	router := app.NewRouter(cfg, db, redisClient)
	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
