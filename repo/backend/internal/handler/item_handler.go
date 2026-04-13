package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/errs"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
	"github.com/localinsights/portal/internal/repository"
)

type ItemHandler struct {
	itemRepo repository.ItemRepository
	db       *database.DB
}

func NewItemHandler(itemRepo repository.ItemRepository, db *database.DB) *ItemHandler {
	return &ItemHandler{
		itemRepo: itemRepo,
		db:       db,
	}
}

func (h *ItemHandler) List(c *gin.Context) {
	pg := getPagination(c)
	search := c.Query("search")
	category := c.Query("category")

	items, total, err := h.itemRepo.ListPublished(c.Request.Context(), search, category, pg)
	if err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResponse(items, pg, total))
}

func (h *ItemHandler) GetByID(c *gin.Context) {
	itemUUID := c.Param("id")
	if itemUUID == "" {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Item ID is required"))
		return
	}

	if _, err := uuid.Parse(itemUUID); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Invalid item ID format"))
		return
	}

	item, err := h.itemRepo.GetByUUID(c.Request.Context(), itemUUID)
	if err != nil {
		respondAppError(c, err)
		return
	}

	agg, _ := h.itemRepo.GetRatingAggregate(c.Request.Context(), item.ID)

	resp := gin.H{
		"id":              item.UUID,
		"title":           item.Title,
		"description":     item.Description,
		"category":        item.Category,
		"lifecycle_state": item.LifecycleState,
		"published_at":    item.PublishedAt,
		"archived_at":     item.ArchivedAt,
		"created_at":      item.CreatedAt,
		"updated_at":      item.UpdatedAt,
	}

	if agg != nil {
		resp["rating"] = gin.H{
			"avg_rating":   agg.AvgRating,
			"rating_count": agg.RatingCount,
			"rating_1":     agg.Rating1,
			"rating_2":     agg.Rating2,
			"rating_3":     agg.Rating3,
			"rating_4":     agg.Rating4,
			"rating_5":     agg.Rating5,
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ItemHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondAppError(c, errs.ErrUnauthorized)
		return
	}

	role := middleware.GetUserRole(c)
	if role != string(model.RoleAdmin) && role != string(model.RoleProductAnalyst) {
		respondAppError(c, errs.ErrForbidden)
		return
	}

	var req request.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, err.Error()))
		return
	}

	item := &model.Item{
		UUID:           uuid.New().String(),
		Title:          req.Title,
		LifecycleState: model.LifecycleStateDraft,
		CreatedBy:      userID,
	}

	if req.Description != "" {
		item.Description = &req.Description
	}
	if req.Category != "" {
		item.Category = &req.Category
	}

	if err := h.itemRepo.Create(c.Request.Context(), item); err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":              item.UUID,
		"title":           item.Title,
		"description":     item.Description,
		"category":        item.Category,
		"lifecycle_state": item.LifecycleState,
		"created_at":      item.CreatedAt,
		"updated_at":      item.UpdatedAt,
	})
}

func (h *ItemHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondAppError(c, errs.ErrUnauthorized)
		return
	}

	itemUUID := c.Param("id")
	if itemUUID == "" {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Item ID is required"))
		return
	}

	var req request.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, err.Error()))
		return
	}

	item, err := h.itemRepo.GetByUUID(c.Request.Context(), itemUUID)
	if err != nil {
		respondAppError(c, err)
		return
	}

	role := middleware.GetUserRole(c)
	if item.CreatedBy != userID && role != string(model.RoleAdmin) {
		respondAppError(c, errs.ErrForbidden)
		return
	}

	if req.Title != "" {
		item.Title = req.Title
	}
	if req.Description != "" {
		item.Description = &req.Description
	}
	if req.Category != "" {
		item.Category = &req.Category
	}

	if err := h.itemRepo.Update(c.Request.Context(), item); err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              item.UUID,
		"title":           item.Title,
		"description":     item.Description,
		"category":        item.Category,
		"lifecycle_state": item.LifecycleState,
		"published_at":    item.PublishedAt,
		"archived_at":     item.ArchivedAt,
		"created_at":      item.CreatedAt,
		"updated_at":      item.UpdatedAt,
	})
}

func (h *ItemHandler) Transition(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondAppError(c, errs.ErrUnauthorized)
		return
	}

	itemUUID := c.Param("id")
	if itemUUID == "" {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Item ID is required"))
		return
	}

	var req request.TransitionItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, err.Error()))
		return
	}

	item, err := h.itemRepo.GetByUUID(c.Request.Context(), itemUUID)
	if err != nil {
		respondAppError(c, err)
		return
	}

	role := middleware.GetUserRole(c)
	if item.CreatedBy != userID && role != string(model.RoleAdmin) {
		respondAppError(c, errs.ErrForbidden)
		return
	}

	targetState := model.LifecycleState(req.State)
	if !isValidTransition(item.LifecycleState, targetState) {
		respondAppError(c, errs.WithMessage(errs.ErrValidation,
			"Invalid state transition from "+string(item.LifecycleState)+" to "+string(targetState)))
		return
	}

	now := time.Now().UTC()
	item.LifecycleState = targetState
	switch targetState {
	case model.LifecycleStatePublished:
		item.PublishedAt = &now
		item.ArchivedAt = nil
	case model.LifecycleStateArchived:
		item.ArchivedAt = &now
	case model.LifecycleStateDraft:
		item.PublishedAt = nil
		item.ArchivedAt = nil
	}

	if err := h.itemRepo.Update(c.Request.Context(), item); err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              item.UUID,
		"title":           item.Title,
		"description":     item.Description,
		"category":        item.Category,
		"lifecycle_state": item.LifecycleState,
		"published_at":    item.PublishedAt,
		"archived_at":     item.ArchivedAt,
		"created_at":      item.CreatedAt,
		"updated_at":      item.UpdatedAt,
	})
}

// isValidTransition validates allowed lifecycle state transitions.
// draft -> published, published -> archived, archived -> draft
var validTransitions = map[model.LifecycleState][]model.LifecycleState{
	model.LifecycleStateDraft:     {model.LifecycleStatePublished},
	model.LifecycleStatePublished: {model.LifecycleStateArchived},
	model.LifecycleStateArchived:  {model.LifecycleStateDraft},
}

func isValidTransition(from, to model.LifecycleState) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
