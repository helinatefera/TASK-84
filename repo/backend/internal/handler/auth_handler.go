package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/errs"
	"github.com/localinsights/portal/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         user.UUID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ip := c.ClientIP()
	accessToken, refreshToken, user, err := h.authService.Login(c.Request.Context(), &req, ip)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
		"user": gin.H{
			"id":        user.UUID,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role,
			"is_active": user.IsActive,
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	accessToken, user, err := h.authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"expires_in":   900,
		"user": gin.H{
			"id":       user.UUID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req request.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"msg": "Logged out"})
		return
	}

	_ = h.authService.Logout(c.Request.Context(), req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"msg": "Logged out"})
}

// handleError is the standard error handler used across all handlers.
// It maps AppError types to proper HTTP codes and never exposes internals.
func handleError(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, gin.H{"code": appErr.HTTPStatus, "msg": appErr.Message})
		return
	}
	// Never expose internal error details
	c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "An internal error occurred"})
}
