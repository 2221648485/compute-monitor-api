package app

import (
	"context"
	"log"
	"time"

	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/store/mysql"
	"compute-monitor-api/internal/store/redis"

	redisclient "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Resources 保存应用启动后需要复用的基础设施连接。
type Resources struct {
	DB    *gorm.DB
	Redis *redisclient.Client
}

// InitResources 初始化 MySQL、Redis 等基础设施资源。
func InitResources(ctx context.Context, cfg config.Config) Resources {
	resources := Resources{}

	db, err := mysql.New(mysql.Options{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: 30 * time.Minute,
	})
	if err != nil {
		log.Printf("mysql init failed, database api will not be available: %v", err)
	} else {
		if err := mysql.Migrate(db); err != nil {
			log.Fatalf("mysql migrate failed: %v", err)
		}
		resources.DB = db
	}

	redisClient, err := redis.New(ctx, redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Printf("redis init failed, auth api will not be available: %v", err)
	} else {
		resources.Redis = redisClient
	}

	return resources
}

// Close 关闭应用持有的基础设施连接。
func (r Resources) Close() {
	if err := mysql.Close(r.DB); err != nil {
		log.Printf("mysql close failed: %v", err)
	}
	if err := redis.Close(r.Redis); err != nil {
		log.Printf("redis close failed: %v", err)
	}
}
