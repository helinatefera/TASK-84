package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/errs"
	"github.com/localinsights/portal/internal/pkg/jwt"
)

// Auth extracts a Bearer token from the Authorization header and validates
// it using the provided jwt.Manager.
//
// If no token is present the request is allowed to proceed (for public
// routes) with "authenticated" set to false. If a token is present but
// invalid the request is aborted with ErrUnauthorized.
func Auth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.Set("authenticated", false)
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Set("authenticated", false)
			c.Next()
			return
		}

		tokenStr := parts[1]
		claims, err := jwtManager.ValidateAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errs.ErrUnauthorized.Code,
				"msg": errs.ErrUnauthorized.Message,
			})
			return
		}

		c.Set("authenticated", true)
		c.Set("user_id", claims.UserID)
		c.Set("user_uuid", claims.UserUUID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// GetUserID returns the authenticated user's ID from the gin context.
// Returns 0 if not authenticated.
func GetUserID(c *gin.Context) uint64 {
	v, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	id, ok := v.(uint64)
	if !ok {
		return 0
	}
	return id
}

// GetUserRole returns the authenticated user's role from the gin context.
// Returns an empty string if not authenticated.
func GetUserRole(c *gin.Context) string {
	v, exists := c.Get("user_role")
	if !exists {
		return ""
	}
	role, ok := v.(string)
	if !ok {
		return ""
	}
	return role
}

// GetUserUUID returns the authenticated user's UUID from the gin context.
// Returns an empty string if not authenticated.
func GetUserUUID(c *gin.Context) string {
	v, exists := c.Get("user_uuid")
	if !exists {
		return ""
	}
	uid, ok := v.(string)
	if !ok {
		return ""
	}
	return uid
}

// IsAuthenticated returns true if the request has a valid JWT token.
func IsAuthenticated(c *gin.Context) bool {
	v, exists := c.Get("authenticated")
	if !exists {
		return false
	}
	auth, ok := v.(bool)
	if !ok {
		return false
	}
	return auth
}
