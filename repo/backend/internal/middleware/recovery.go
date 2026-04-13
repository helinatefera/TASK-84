package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/errs"
)

// Recovery recovers from panics, logs the error with structured logging
// (including the request_id from context), and returns a 500 JSON response.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Get("request_id")

				slog.Error("panic recovered",
					"error", r,
					"request_id", requestID,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"stack", string(debug.Stack()),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    errs.ErrInternal.Code,
					"msg": errs.ErrInternal.Message,
				})
			}
		}()

		c.Next()
	}
}
