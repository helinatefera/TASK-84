package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/errs"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/repository"
)

type UserHandler struct {
	userRepo  repository.UserRepository
	prefsRepo repository.UserPreferencesRepository
}

func NewUserHandler(userRepo repository.UserRepository, prefsRepo repository.UserPreferencesRepository) *UserHandler {
	return &UserHandler{
		userRepo:  userRepo,
		prefsRepo: prefsRepo,
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondAppError(c, errs.ErrUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		respondAppError(c, err)
		return
	}

	prefs, _ := h.prefsRepo.GetByUserID(c.Request.Context(), userID)

	resp := gin.H{
		"id":         user.UUID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	if prefs != nil {
		resp["preferences"] = gin.H{
			"locale":                prefs.Locale,
			"timezone":              prefs.Timezone,
			"notification_settings": prefs.NotificationSettings,
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondAppError(c, errs.ErrUnauthorized)
		return
	}

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, err.Error()))
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		respondAppError(c, err)
		return
	}

	if req.Username != "" {
		existing, err := h.userRepo.GetByUsername(c.Request.Context(), req.Username)
		if err == nil && existing != nil && existing.ID != userID {
			respondAppError(c, errs.WithMessage(errs.ErrConflict, "Username already taken"))
			return
		}
		user.Username = req.Username
	}

	if req.Email != "" {
		existing, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
		if err == nil && existing != nil && existing.ID != userID {
			respondAppError(c, errs.WithMessage(errs.ErrConflict, "Email already taken"))
			return
		}
		user.Email = req.Email
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.UUID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

func (h *UserHandler) UpdatePreferences(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondAppError(c, errs.ErrUnauthorized)
		return
	}

	var req request.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, err.Error()))
		return
	}

	prefs, _ := h.prefsRepo.GetByUserID(c.Request.Context(), userID)
	now := time.Now().UTC()
	if prefs == nil {
		prefs = &model.UserPreferences{
			UserID:    userID,
			CreatedAt: now,
		}
	}
	prefs.UpdatedAt = now

	if req.Locale != "" {
		prefs.Locale = req.Locale
	}
	if req.Timezone != "" {
		prefs.Timezone = req.Timezone
	}

	if err := h.prefsRepo.Upsert(c.Request.Context(), prefs); err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"locale":                prefs.Locale,
		"timezone":              prefs.Timezone,
		"notification_settings": prefs.NotificationSettings,
	})
}
