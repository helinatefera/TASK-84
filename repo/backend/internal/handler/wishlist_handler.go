package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/repository"
)

type WishlistHandler struct {
	wishlistRepo repository.WishlistRepository
}

func NewWishlistHandler(wishlistRepo repository.WishlistRepository) *WishlistHandler {
	return &WishlistHandler{
		wishlistRepo: wishlistRepo,
	}
}

// List handles GET /wishlists
func (h *WishlistHandler) List(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	userID := middleware.GetUserID(c)

	wishlists, err := h.wishlistRepo.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to list wishlists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wishlists})
}

// createWishlistRequest is the JSON body for creating a wishlist.
type createWishlistRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// Create handles POST /wishlists
func (h *WishlistHandler) Create(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	var req createWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	now := time.Now().UTC()

	wishlist := &model.Wishlist{
		UUID:      uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.wishlistRepo.Create(c.Request.Context(), wishlist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to create wishlist"})
		return
	}

	c.JSON(http.StatusCreated, wishlist)
}

// Update handles PUT /wishlists/:id
func (h *WishlistHandler) Update(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	wishlistUUID := c.Param("id")
	wishlist, err := h.wishlistRepo.GetByUUID(c.Request.Context(), wishlistUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve wishlist"})
		return
	}
	if wishlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Wishlist not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if wishlist.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "You can only edit your own wishlists"})
		return
	}

	var req createWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	wishlist.Name = req.Name
	wishlist.UpdatedAt = time.Now().UTC()

	if err := h.wishlistRepo.Update(c.Request.Context(), wishlist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to update wishlist"})
		return
	}

	c.JSON(http.StatusOK, wishlist)
}

// Delete handles DELETE /wishlists/:id
func (h *WishlistHandler) Delete(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	wishlistUUID := c.Param("id")
	wishlist, err := h.wishlistRepo.GetByUUID(c.Request.Context(), wishlistUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve wishlist"})
		return
	}
	if wishlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Wishlist not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if wishlist.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "You can only delete your own wishlists"})
		return
	}

	if err := h.wishlistRepo.Delete(c.Request.Context(), wishlist.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to delete wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "Wishlist deleted"})
}

// wishlistAddItemRequest is the JSON body for adding an item to a wishlist.
type wishlistAddItemRequest struct {
	ItemID uint64 `json:"item_id" binding:"required"`
}

// AddItem handles POST /wishlists/:id/items
func (h *WishlistHandler) AddItem(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	wishlistUUID := c.Param("id")
	wishlist, err := h.wishlistRepo.GetByUUID(c.Request.Context(), wishlistUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve wishlist"})
		return
	}
	if wishlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Wishlist not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if wishlist.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "You can only modify your own wishlists"})
		return
	}

	var req wishlistAddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	if err := h.wishlistRepo.AddItem(c.Request.Context(), wishlist.ID, req.ItemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to add item to wishlist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": http.StatusCreated, "msg": "Item added to wishlist"})
}

// RemoveItem handles DELETE /wishlists/:id/items/:item_id
func (h *WishlistHandler) RemoveItem(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	wishlistUUID := c.Param("id")
	wishlist, err := h.wishlistRepo.GetByUUID(c.Request.Context(), wishlistUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve wishlist"})
		return
	}
	if wishlist == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Wishlist not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if wishlist.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "You can only modify your own wishlists"})
		return
	}

	itemID, ok := parseUintParam(c, "item_id")
	if !ok {
		return
	}

	if err := h.wishlistRepo.RemoveItem(c.Request.Context(), wishlist.ID, itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to remove item from wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "Item removed from wishlist"})
}
