package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/errs"
)

// RequireRole returns a middleware that ensures the authenticated user has
// one of the specified roles. If the user is not authenticated or does not
// have a matching role the request is aborted.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		if !IsAuthenticated(c) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errs.ErrUnauthorized.Code,
				"msg": errs.ErrUnauthorized.Message,
			})
			return
		}

		role := GetUserRole(c)
		if _, ok := allowed[role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    errs.ErrForbidden.Code,
				"msg": errs.ErrForbidden.Message,
			})
			return
		}

		c.Next()
	}
}

// RequireAuth returns a middleware that ensures the user is authenticated.
// No role check is performed.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAuthenticated(c) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errs.ErrUnauthorized.Code,
				"msg": errs.ErrUnauthorized.Message,
			})
			return
		}

		c.Next()
	}
}
