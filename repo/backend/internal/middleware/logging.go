package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging records the start time, calls c.Next(), then logs request details
// including method, path, status code, latency, client IP, and request ID.
// Requests that result in a 5xx status are logged at error level.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID, _ := c.Get("request_id")

		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
			"request_id", requestID,
		}

		if status >= 500 {
			slog.Error("request completed with server error", attrs...)
		} else {
			slog.Info("request completed", attrs...)
		}
	}
}
