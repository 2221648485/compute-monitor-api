package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLog 记录简洁请求日志，主要用于观察调度器实际请求了哪些接口。
func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		slog.Info("http request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"raw_path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
