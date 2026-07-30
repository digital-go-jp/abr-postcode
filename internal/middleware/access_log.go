package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLog returns a Gin middleware that reports one record per request
// through slog, so request logs share the destination and format of the rest
// of the application's output.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		}
		if err := c.Errors.ByType(gin.ErrorTypePrivate).String(); err != "" {
			attrs = append(attrs, "error", err)
		}

		slog.Info("request", attrs...)
	}
}
