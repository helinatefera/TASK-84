package handler

import (
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/config"
	"github.com/localinsights/portal/internal/errs"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/imagepro"
	"github.com/localinsights/portal/internal/repository"
)

type ImageHandler struct {
	imageRepo  repository.ImageRepository
	storageCfg config.StorageConfig
}

func NewImageHandler(imageRepo repository.ImageRepository, storageCfg config.StorageConfig) *ImageHandler {
	return &ImageHandler{
		imageRepo:  imageRepo,
		storageCfg: storageCfg,
	}
}

func (h *ImageHandler) Upload(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondAppError(c, errs.ErrUnauthorized)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "No file uploaded"))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respondAppError(c, errs.ErrInternal)
		return
	}

	result, err := imagepro.ProcessUpload(data, h.storageCfg.ImagesDir, h.storageCfg.MaxImageSize)
	if err != nil {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, err.Error()))
		return
	}

	// Check for existing image with same hash (deduplication).
	existing, _ := h.imageRepo.GetByHash(c.Request.Context(), result.SHA256Hash)
	if existing != nil {
		c.JSON(http.StatusOK, gin.H{
			"image_id":      existing.ID,
			"hash":          existing.SHA256Hash,
			"original_name": existing.OriginalName,
			"mime_type":     existing.MimeType,
			"file_size":     existing.FileSize,
			"status":        existing.Status,
			"created_at":    existing.CreatedAt,
			"deduplicated":  true,
		})
		return
	}

	status := "active"
	var quarantineReason *string

	if result.IsSuspicious {
		status = "quarantined"
		reason := result.SuspiciousReason
		quarantineReason = &reason

		// Move file to quarantine directory.
		if moveErr := imagepro.MoveToQuarantine(h.storageCfg.ImagesDir, h.storageCfg.QuarantineDir, result.StoragePath); moveErr != nil {
			respondAppError(c, errs.ErrInternal)
			return
		}
	}

	img := &model.Image{
		SHA256Hash:       result.SHA256Hash,
		OriginalName:     header.Filename,
		MimeType:         result.MimeType,
		FileSize:         uint32(result.FileSize),
		StoragePath:      result.StoragePath,
		Status:           status,
		QuarantineReason: quarantineReason,
		UploadedBy:       userID,
	}

	if err := h.imageRepo.Create(c.Request.Context(), img); err != nil {
		respondAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"image_id":      img.ID,
		"hash":          img.SHA256Hash,
		"original_name": img.OriginalName,
		"mime_type":     img.MimeType,
		"file_size":     img.FileSize,
		"status":        img.Status,
		"created_at":    img.CreatedAt,
		"deduplicated":  false,
	})
}

func (h *ImageHandler) ServeByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		respondAppError(c, errs.WithMessage(errs.ErrValidation, "Image hash is required"))
		return
	}

	img, err := h.imageRepo.GetByHash(c.Request.Context(), hash)
	if err != nil {
		respondAppError(c, err)
		return
	}

	if img.Status == "quarantined" {
		respondAppError(c, errs.ErrImageQuarantined)
		return
	}

	fullPath := filepath.Join(h.storageCfg.ImagesDir, img.StoragePath)

	c.Header("Content-Type", img.MimeType)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.File(fullPath)
}
