package main

import (
	"log"
	"net/http"
	"time"

	"compute-monitor-api/internal/config"
	"compute-monitor-api/internal/store/mysql"
	"compute-monitor-api/internal/student"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := mysql.New(mysql.Options{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    config.GetEnvInt("MYSQL_MAX_OPEN_CONNS", 20),
		MaxIdleConns:    config.GetEnvInt("MYSQL_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: 30 * time.Minute,
	})
	if err != nil {
		log.Printf("mysql init failed, student api will not be available: %v", err)
	}
	if db != nil {
		defer db.Close()
	}

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	api := router.Group("/api/v1")
	if db != nil {
		studentRepository := student.NewMySQLRepository(db)
		studentService := student.NewService(studentRepository)
		studentHandler := student.NewHandler(studentService)
		student.RegisterRoutes(api, studentHandler)
	}

	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
