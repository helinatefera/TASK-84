package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/dto/response"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/repository"
)

type FavoriteHandler struct {
	favoriteRepo repository.FavoriteRepository
}

func NewFavoriteHandler(favoriteRepo repository.FavoriteRepository) *FavoriteHandler {
	return &FavoriteHandler{
		favoriteRepo: favoriteRepo,
	}
}

// List handles GET /favorites
func (h *FavoriteHandler) List(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	userID := middleware.GetUserID(c)
	pg := getPagination(c)

	favorites, total, err := h.favoriteRepo.ListByUser(c.Request.Context(), userID, pg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to list favorites"})
		return
	}

	c.JSON(http.StatusOK, response.NewPaginated(favorites, pg.Page, pg.PerPage, total))
}

// addFavoriteRequest is the JSON body for adding a favorite.
type addFavoriteRequest struct {
	ItemID uint64 `json:"item_id" binding:"required"`
}

// Add handles POST /favorites
func (h *FavoriteHandler) Add(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	var req addFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)

	exists, err := h.favoriteRepo.Exists(c.Request.Context(), userID, req.ItemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to check favorite"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"code": http.StatusConflict, "msg": "Item is already in favorites"})
		return
	}

	if err := h.favoriteRepo.Add(c.Request.Context(), userID, req.ItemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to add favorite"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": http.StatusCreated, "msg": "Favorite added"})
}

// Remove handles DELETE /favorites/:item_id
func (h *FavoriteHandler) Remove(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	itemID, ok := parseUintParam(c, "item_id")
	if !ok {
		return
	}

	userID := middleware.GetUserID(c)

	exists, err := h.favoriteRepo.Exists(c.Request.Context(), userID, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to check favorite"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Favorite not found"})
		return
	}

	if err := h.favoriteRepo.Remove(c.Request.Context(), userID, itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to remove favorite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "Favorite removed"})
}
