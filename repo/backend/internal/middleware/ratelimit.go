package middleware

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/errs"
	"golang.org/x/time/rate"
)

// RateLimit returns a middleware that enforces per-key rate limiting.
// The key is the authenticated user's ID when available, otherwise the
// client IP address. Each key is allowed requestsPerMinute tokens with a
// refill rate of requestsPerMinute/60 per second.
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	var limiters sync.Map

	rps := float64(requestsPerMinute) / 60.0
	burst := requestsPerMinute

	return func(c *gin.Context) {
		var key string
		if IsAuthenticated(c) {
			key = fmt.Sprintf("user:%d", GetUserID(c))
		} else {
			key = fmt.Sprintf("ip:%s", c.ClientIP())
		}

		v, _ := limiters.LoadOrStore(key, rate.NewLimiter(rate.Limit(rps), burst))
		limiter := v.(*rate.Limiter)

		if !limiter.Allow() {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    errs.ErrRateLimited.Code,
				"msg": errs.ErrRateLimited.Message,
			})
			return
		}

		c.Next()
	}
}
