package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/localinsights/portal/internal/config"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/errs"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
	"github.com/localinsights/portal/internal/repository"
	"github.com/localinsights/portal/internal/service"
)

type ReviewHandler struct {
	reviewRepo    repository.ReviewRepository
	imageRepo     repository.ImageRepository
	riRepo        repository.ReviewImageRepository
	contentFilter *service.ContentFilter
	db            *database.DB
	storageCfg    config.StorageConfig
}

func NewReviewHandler(
	reviewRepo repository.ReviewRepository,
	imageRepo repository.ImageRepository,
	riRepo repository.ReviewImageRepository,
	contentFilter *service.ContentFilter,
	db *database.DB,
	storageCfg config.StorageConfig,
) *ReviewHandler {
	return &ReviewHandler{
		reviewRepo:    reviewRepo,
		imageRepo:     imageRepo,
		riRepo:        riRepo,
		contentFilter: contentFilter,
		db:            db,
		storageCfg:    storageCfg,
	}
}

func (h *ReviewHandler) ListByItem(c *gin.Context) {
	itemUUID := c.Param("id")
	if itemUUID == "" {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Item ID is required"))
		return
	}

	if _, err := uuid.Parse(itemUUID); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Invalid item ID format"))
		return
	}

	// Resolve the item UUID to an internal ID.
	var itemID uint64
	row := h.db.QueryRowContext(c.Request.Context(), "SELECT id FROM items WHERE uuid = ?", itemUUID)
	if err := row.Scan(&itemID); err != nil {
		respondAppError(c, errs.ErrNotFound)
		return
	}

	pg := getPagination(c)

	reviews, total, err := h.reviewRepo.ListByItem(c.Request.Context(), itemID, pg)
	if err != nil {
		respondAppError(c, err)
		return
	}

	// Enrich each review with its linked images
	type reviewWithImages struct {
		*model.Review
		Images []gin.H `json:"images"`
	}
	enriched := make([]reviewWithImages, 0, len(reviews))
	for _, r := range reviews {
		rwi := reviewWithImages{Review: r, Images: []gin.H{}}
		ris, _ := h.riRepo.ListByReview(c.Request.Context(), r.ID)
		for _, ri := range ris {
			img, _ := h.imageRepo.GetByID(c.Request.Context(), ri.ImageID)
			if img != nil {
				rwi.Images = append(rwi.Images, gin.H{
					"image_id":  img.ID,
					"hash":      img.SHA256Hash,
					"mime_type":  img.MimeType,
					"file_size":  img.FileSize,
				})
			}
		}
		enriched = append(enriched, rwi)
	}

	c.JSON(http.StatusOK, paginatedResponse(enriched, pg, total))
}

func (h *ReviewHandler) Create(c *gin.Context) {
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

	if _, err := uuid.Parse(itemUUID); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Invalid item ID format"))
		return
	}

	var req request.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, err.Error()))
		return
	}

	// Resolve item UUID to internal ID.
	var itemID uint64
	row := h.db.QueryRowContext(c.Request.Context(), "SELECT id FROM items WHERE uuid = ?", itemUUID)
	if err := row.Scan(&itemID); err != nil {
		respondAppError(c, errs.ErrNotFound)
		return
	}

	// Run content through sensitive-word filter.
	filteredBody := req.Body
	if req.Body != "" {
		cleaned, blockReason, flagged := h.contentFilter.Apply(c.Request.Context(), req.Body)
		if blockReason != "" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": blockReason})
			return
		}
		filteredBody = cleaned
		if flagged {
			h.db.ExecContext(c.Request.Context(),
				`INSERT INTO audit_logs (actor_id, action, target_type, details, created_at)
				VALUES (?, 'content.flagged', 'review', '{"reason":"sensitive_word_match"}', NOW(3))`, userID)
		}
	}

	review := &model.Review{
		UUID:        uuid.New().String(),
		ItemID:      itemID,
		UserID:      userID,
		Rating:      uint8(req.Rating),
		FraudStatus: model.FraudStatusNormal,
	}

	if filteredBody != "" {
		review.Body = &filteredBody
	}

	if len(req.ImageIDs) > 6 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Maximum 6 images per review"})
		return
	}

	// Validate all images upfront before starting the transaction so we
	// don't hold a transaction open while doing ownership checks.
	type validatedImage struct {
		ID       uint64
		Hash     string
		MimeType string
	}
	var images []validatedImage
	for _, imgID := range req.ImageIDs {
		img, err := h.imageRepo.GetByID(c.Request.Context(), imgID)
		if err != nil || img == nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Image not found"})
			return
		}
		if img.UploadedBy != userID {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "Image does not belong to you"})
			return
		}
		images = append(images, validatedImage{ID: imgID, Hash: img.SHA256Hash, MimeType: img.MimeType})
	}

	// Create review + link images atomically.
	err := h.db.WithTx(c.Request.Context(), func(txCtx context.Context) error {
		if err := h.reviewRepo.Create(txCtx, review); err != nil {
			return err
		}
		for order, img := range images {
			if err := h.riRepo.Create(txCtx, &model.ReviewImage{
				ReviewID:  review.ID,
				ImageID:   img.ID,
				SortOrder: uint8(order),
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		respondAppError(c, err)
		return
	}

	linkedImages := make([]gin.H, 0, len(images))
	for _, img := range images {
		linkedImages = append(linkedImages, gin.H{
			"image_id":  img.ID,
			"hash":      img.Hash,
			"mime_type": img.MimeType,
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           review.UUID,
		"rating":       review.Rating,
		"body":         review.Body,
		"fraud_status": review.FraudStatus,
		"images":       linkedImages,
		"created_at":   review.CreatedAt,
		"updated_at":   review.UpdatedAt,
	})
}

func (h *ReviewHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondAppError(c, errs.ErrUnauthorized)
		return
	}

	reviewUUID := c.Param("reviewId")
	if reviewUUID == "" {
		reviewUUID = c.Param("id")
	}
	if reviewUUID == "" {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Review ID is required"))
		return
	}

	var req request.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, err.Error()))
		return
	}

	review, err := h.reviewRepo.GetByUUID(c.Request.Context(), reviewUUID)
	if err != nil || review == nil {
		respondAppError(c, errs.ErrNotFound)
		return
	}

	if review.UserID != userID {
		respondAppError(c, errs.WithMessage(errs.ErrForbidden, "You can only update your own reviews"))
		return
	}

	if req.Rating > 0 {
		review.Rating = uint8(req.Rating)
	}
	if req.Body != "" {
		cleaned, blockReason, flagged := h.contentFilter.Apply(c.Request.Context(), req.Body)
		if blockReason != "" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": blockReason})
			return
		}
		review.Body = &cleaned
		if flagged {
			h.db.ExecContext(c.Request.Context(),
				`INSERT INTO audit_logs (actor_id, action, target_type, details, created_at)
				VALUES (?, 'content.flagged', 'review', '{"reason":"sensitive_word_match"}', NOW(3))`, userID)
		}
	}

	if err := h.reviewRepo.Update(c.Request.Context(), review); err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           review.UUID,
		"rating":       review.Rating,
		"body":         review.Body,
		"fraud_status": review.FraudStatus,
		"created_at":   review.CreatedAt,
		"updated_at":   review.UpdatedAt,
	})
}

func (h *ReviewHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondAppError(c, errs.ErrUnauthorized)
		return
	}

	reviewUUID := c.Param("reviewId")
	if reviewUUID == "" {
		reviewUUID = c.Param("id")
	}
	if reviewUUID == "" {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Review ID is required"))
		return
	}

	review, err := h.reviewRepo.GetByUUID(c.Request.Context(), reviewUUID)
	if err != nil {
		respondAppError(c, err)
		return
	}

	role := middleware.GetUserRole(c)
	if review.UserID != userID && role != string(model.RoleAdmin) {
		respondAppError(c, errs.ErrForbidden)
		return
	}

	if err := h.reviewRepo.Delete(c.Request.Context(), review.ID); err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Review deleted"})
}
