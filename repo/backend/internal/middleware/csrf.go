package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/errs"
)

// CSRF implements the double-submit cookie pattern for CSRF protection.
//
// GET requests to /api/v1/csrf generate a random token, set it as a cookie
// (Secure, HttpOnly=false so JavaScript can read it, SameSite=Strict), and
// return the token in a JSON response.
//
// Mutating requests (POST, PUT, PATCH, DELETE) must include the token in
// the X-CSRF-Token header matching the csrf_token cookie. If the enabled
// flag is false the middleware is a no-op.
func CSRF(enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		// Token issuance endpoint.
		if c.Request.Method == http.MethodGet && c.Request.URL.Path == "/api/v1/csrf" {
			token, err := generateCSRFToken()
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    errs.ErrInternal.Code,
					"msg": errs.ErrInternal.Message,
				})
				return
			}

			c.SetCookie("csrf_token", token, 0, "/", "", true, false)
			c.JSON(http.StatusOK, gin.H{"csrf_token": token})
			c.Abort()
			return
		}

		// For mutating methods, validate the double-submit cookie.
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			headerToken := c.GetHeader("X-CSRF-Token")
			cookieToken, err := c.Cookie("csrf_token")

			if err != nil || headerToken == "" || headerToken != cookieToken {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"code":    errs.ErrCSRFInvalid.Code,
					"msg": errs.ErrCSRFInvalid.Message,
				})
				return
			}
		}

		c.Next()
	}
}

// generateCSRFToken returns a cryptographically random 32-byte hex-encoded
// string.
func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
