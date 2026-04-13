package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/dto/response"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/repository"
)

type NotificationHandler struct {
	notifRepo repository.NotificationRepository
}

func NewNotificationHandler(notifRepo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		notifRepo: notifRepo,
	}
}

// List handles GET /notifications
func (h *NotificationHandler) List(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	userID := middleware.GetUserID(c)
	pg := getPagination(c)

	unreadOnly := false
	if v := c.Query("unread_only"); v == "true" || v == "1" {
		unreadOnly = true
	}

	notifications, total, err := h.notifRepo.ListByUser(c.Request.Context(), userID, unreadOnly, pg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to list notifications"})
		return
	}

	c.JSON(http.StatusOK, response.NewPaginated(notifications, pg.Page, pg.PerPage, total))
}

// GetByID handles GET /notifications/:id
func (h *NotificationHandler) GetByID(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	notification, err := h.notifRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve notification"})
		return
	}
	if notification == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Notification not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if notification.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// UnreadCount handles GET /notifications/unread-count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	userID := middleware.GetUserID(c)

	count, err := h.notifRepo.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkRead handles PUT /notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	notification, err := h.notifRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve notification"})
		return
	}
	if notification == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Notification not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if notification.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Notification not found"})
		return
	}

	if err := h.notifRepo.MarkRead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "Notification marked as read"})
}

// MarkAllRead handles PUT /notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.notifRepo.MarkAllRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "All notifications marked as read"})
}
