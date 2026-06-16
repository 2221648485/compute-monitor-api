package main

import (
	"context"
	"log"

	"compute-monitor-api/internal/app"
	"compute-monitor-api/internal/config"
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
	resources := app.InitResources(context.Background(), cfg)
	defer resources.Close()

	stopSchedulers := app.StartSchedulers(context.Background(), cfg, resources)
	defer stopSchedulers()

	router := app.NewRouter(cfg, resources.DB, resources.Redis)
	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
