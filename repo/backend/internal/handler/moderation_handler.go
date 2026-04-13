package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
	"github.com/localinsights/portal/internal/repository"
)

type ModerationHandler struct {
	reportRepo   repository.ReportRepository
	appealRepo   repository.AppealRepository
	noteRepo     repository.ModerationNoteRepository
	wordRuleRepo repository.SensitiveWordRuleRepository
	imageRepo    repository.ImageRepository
	reviewRepo   repository.ReviewRepository
	db           *database.DB
}

func NewModerationHandler(
	reportRepo repository.ReportRepository,
	appealRepo repository.AppealRepository,
	noteRepo repository.ModerationNoteRepository,
	wordRuleRepo repository.SensitiveWordRuleRepository,
	imageRepo repository.ImageRepository,
	reviewRepo repository.ReviewRepository,
	db *database.DB,
) *ModerationHandler {
	return &ModerationHandler{
		reportRepo:   reportRepo,
		appealRepo:   appealRepo,
		noteRepo:     noteRepo,
		wordRuleRepo: wordRuleRepo,
		imageRepo:    imageRepo,
		reviewRepo:   reviewRepo,
		db:           db,
	}
}

// ListQueue returns the moderation queue filtered by optional status and target_type.
func (h *ModerationHandler) ListQueue(c *gin.Context) {
	status := c.Query("status")
	targetType := c.Query("target_type")
	pag := getPagination(c)

	reports, total, err := h.reportRepo.ListQueue(c.Request.Context(), status, targetType, pag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list moderation queue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     reports,
		"total":    total,
		"page":     pag.Page,
		"per_page": pag.PerPage,
	})
}

// CreateReport creates a new content report.
func (h *ModerationHandler) CreateReport(c *gin.Context) {
	var req request.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	reporterID := middleware.GetUserID(c)
	if reporterID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "authentication required"})
		return
	}

	targetID, err := strconv.ParseUint(req.TargetID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid target_id"})
		return
	}

	now := time.Now().UTC()
	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	report := &model.Report{
		UUID:       uuid.New().String(),
		ReporterID: reporterID,
		TargetType: req.TargetType,
		TargetID:   targetID,
		Category:   model.ReportCategory(req.Category),
		Description: desc,
		Status:     model.ReportStatusPending,
		Priority:   0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.reportRepo.Create(c.Request.Context(), report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to create report"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          report.ID,
		"uuid":        report.UUID,
		"target_type": report.TargetType,
		"target_id":   report.TargetID,
		"category":    report.Category,
		"status":      report.Status,
		"created_at":  report.CreatedAt,
	})
}

// ListMyReports returns the authenticated user's own reports.
func (h *ModerationHandler) ListMyReports(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "authentication required"})
		return
	}

	pag := getPagination(c)

	reports, total, err := h.reportRepo.ListByReporter(c.Request.Context(), userID, pag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     reports,
		"total":    total,
		"page":     pag.Page,
		"per_page": pag.PerPage,
	})
}

// UpdateReport updates an existing report's status and notes.
func (h *ModerationHandler) UpdateReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid report id"})
		return
	}

	var req request.UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	report, err := h.reportRepo.GetByID(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "report not found"})
		return
	}

	report.Status = model.ReportStatus(req.Status)
	report.UpdatedAt = time.Now().UTC()

	if req.ResolutionNote != "" {
		report.ResolutionNote = &req.ResolutionNote
	}
	if req.UserVisibleNote != "" {
		report.UserVisibleNote = &req.UserVisibleNote
	}

	if model.ReportStatus(req.Status) == model.ReportStatusResolved {
		now := time.Now().UTC()
		report.ResolvedAt = &now
	}

	if err := h.reportRepo.Update(c.Request.Context(), report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to update report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": report})
}

// AddNote adds a moderation note to a report.
func (h *ModerationHandler) AddNote(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid report id"})
		return
	}

	var req request.CreateModerationNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	authorID := middleware.GetUserID(c)
	if authorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "authentication required"})
		return
	}

	note := &model.ModerationNote{
		ReportID:   reportID,
		AuthorID:   authorID,
		Body:       req.Body,
		IsInternal: req.IsInternal,
		CreatedAt:  time.Now().UTC(),
	}

	if err := h.noteRepo.Create(c.Request.Context(), note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to add note"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": note})
}

// ListNotes returns all moderation notes for a report.
func (h *ModerationHandler) ListNotes(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid report id"})
		return
	}

	notes, err := h.noteRepo.ListByReport(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list notes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notes})
}

// CreateAppeal creates an appeal for a report, if one does not already exist.
func (h *ModerationHandler) CreateAppeal(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid report id"})
		return
	}

	var req request.CreateAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "authentication required"})
		return
	}

	// Verify the caller is the reporter (object-level authorization)
	report, err := h.reportRepo.GetByID(c.Request.Context(), reportID)
	if err != nil || report == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "report not found"})
		return
	}
	if report.ReporterID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "You can only appeal your own reports"})
		return
	}

	// Check if an appeal already exists for this report.
	existing, _ := h.appealRepo.GetByReportID(c.Request.Context(), reportID)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"code": http.StatusConflict, "msg": "an appeal already exists for this report"})
		return
	}

	appeal := &model.Appeal{
		UUID:      uuid.New().String(),
		ReportID:  reportID,
		UserID:    userID,
		Body:      req.Body,
		Status:    model.AppealStatusPending,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.appealRepo.Create(c.Request.Context(), appeal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to create appeal"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": appeal})
}

// ResubmitAppeal lets the appeal owner revise the body and reset the status to
// pending. Only allowed when the current status is needs_edit.
func (h *ModerationHandler) ResubmitAppeal(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid report id"})
		return
	}

	var req request.CreateAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "authentication required"})
		return
	}

	existing, _ := h.appealRepo.GetByReportID(c.Request.Context(), reportID)
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "no appeal found for this report"})
		return
	}
	if existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "you can only edit your own appeals"})
		return
	}
	if existing.Status != model.AppealStatusNeedsEdit {
		c.JSON(http.StatusConflict, gin.H{"code": http.StatusConflict, "msg": "appeal can only be resubmitted when status is needs_edit"})
		return
	}

	if err := h.appealRepo.Resubmit(c.Request.Context(), existing.ID, req.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to resubmit appeal"})
		return
	}

	existing.Body = req.Body
	existing.Status = model.AppealStatusPending
	existing.ReviewedBy = nil
	existing.ReviewedAt = nil
	c.JSON(http.StatusOK, gin.H{"data": existing})
}

// HandleAppeal updates an appeal's status (accepted/rejected) with reviewer info.
func (h *ModerationHandler) HandleAppeal(c *gin.Context) {
	appealID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid appeal id"})
		return
	}

	var req request.HandleAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "authentication required"})
		return
	}

	// Load the existing appeal to get the report ID for the note.
	var existingAppeal model.Appeal
	if err := h.db.GetContext(c.Request.Context(), &existingAppeal, "SELECT * FROM appeals WHERE id = ?", appealID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "appeal not found"})
		return
	}

	now := time.Now().UTC()
	appeal := &model.Appeal{
		ID:     appealID,
		Status: model.AppealStatus(req.Status),
	}
	// Only stamp reviewer info on terminal statuses; needs_edit is
	// a request for the user to revise, not a final resolution.
	if req.Status != string(model.AppealStatusNeedsEdit) {
		appeal.ReviewedBy = &reviewerID
		appeal.ReviewedAt = &now
	}

	if err := h.appealRepo.Update(c.Request.Context(), appeal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to update appeal"})
		return
	}

	// Persist the moderator note on the associated report.
	note := &model.ModerationNote{
		ReportID:  existingAppeal.ReportID,
		AuthorID:  reviewerID,
		Body:      req.Note,
		IsInternal: false,
		CreatedAt: now,
	}
	_ = h.noteRepo.Create(c.Request.Context(), note)

	c.JSON(http.StatusOK, gin.H{"data": appeal})
}

// ListQuarantined returns paginated quarantined images.
func (h *ModerationHandler) ListQuarantined(c *gin.Context) {
	pag := getPagination(c)

	images, total, err := h.imageRepo.ListQuarantined(c.Request.Context(), pag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list quarantined images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     images,
		"total":    total,
		"page":     pag.Page,
		"per_page": pag.PerPage,
	})
}

// HandleQuarantine processes a quarantined image (approve, reject, or keep).
func (h *ModerationHandler) HandleQuarantine(c *gin.Context) {
	imageID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid image id"})
		return
	}

	var req request.QuarantineActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	var newStatus string
	var reason *string
	switch req.Action {
	case "approve":
		newStatus = "approved"
	case "reject":
		newStatus = "rejected"
		r := "rejected by moderator"
		reason = &r
	case "keep":
		newStatus = "quarantined"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid action"})
		return
	}

	if err := h.imageRepo.UpdateStatus(c.Request.Context(), imageID, newStatus, reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to update image status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "image status updated", "status": newStatus})
}

// ListFraudReviews returns reviews with suspected fraud status.
func (h *ModerationHandler) ListFraudReviews(c *gin.Context) {
	pag := getPagination(c)
	offset := pag.Offset()

	var reviews []model.Review
	err := h.db.SelectContext(
		c.Request.Context(),
		&reviews,
		"SELECT * FROM reviews WHERE fraud_status = 'suspected_fraud' ORDER BY updated_at DESC LIMIT ? OFFSET ?",
		pag.PerPage,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list fraud reviews"})
		return
	}

	var total int64
	_ = h.db.GetContext(
		c.Request.Context(),
		&total,
		"SELECT COUNT(*) FROM reviews WHERE fraud_status = 'suspected_fraud'",
	)

	c.JSON(http.StatusOK, gin.H{
		"data":     reviews,
		"total":    total,
		"page":     pag.Page,
		"per_page": pag.PerPage,
	})
}

// HandleFraud confirms or clears fraud status on a review.
func (h *ModerationHandler) HandleFraud(c *gin.Context) {
	reviewID, err := strconv.ParseUint(c.Param("review_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid review id"})
		return
	}

	var req request.FraudActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	var newStatus model.FraudStatus
	switch req.Action {
	case "confirm":
		newStatus = model.FraudStatusConfirmed
	case "clear":
		newStatus = model.FraudStatusCleared
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid action"})
		return
	}

	if err := h.reviewRepo.UpdateFraudStatus(c.Request.Context(), reviewID, newStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to update fraud status"})
		return
	}

	// Propagate fraud status to the account level.
	var reviewUserID uint64
	_ = h.db.GetContext(c.Request.Context(), &reviewUserID, "SELECT user_id FROM reviews WHERE id = ?", reviewID)
	if reviewUserID != 0 {
		var accountStatus model.UserFraudStatus
		if newStatus == model.FraudStatusConfirmed {
			accountStatus = model.UserFraudConfirmed
		} else if newStatus == model.FraudStatusCleared {
			accountStatus = model.UserFraudClean
		}
		if accountStatus != "" {
			_, _ = h.db.ExecContext(c.Request.Context(),
				"UPDATE users SET fraud_status = ? WHERE id = ?", accountStatus, reviewUserID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"msg": "fraud status updated", "fraud_status": string(newStatus)})
}

// ListAppeals returns a paginated list of appeals, optionally filtered by status.
func (h *ModerationHandler) ListAppeals(c *gin.Context) {
	pag := getPagination(c)
	status := c.Query("status")

	appeals, total, err := h.appealRepo.List(c.Request.Context(), status, pag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list appeals"})
		return
	}

	c.JSON(http.StatusOK, paginatedResponse(appeals, pag, total))
}

// ListWordRules returns all active sensitive word rules.
func (h *ModerationHandler) ListWordRules(c *gin.Context) {
	rules, err := h.wordRuleRepo.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list word rules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// CreateWordRule creates a new sensitive word rule.
func (h *ModerationHandler) CreateWordRule(c *gin.Context) {
	var req request.CreateSensitiveWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "authentication required"})
		return
	}

	now := time.Now().UTC()
	var replacement *string
	if req.Replacement != "" {
		replacement = &req.Replacement
	}

	rule := &model.SensitiveWordRule{
		Pattern:     req.Pattern,
		Action:      req.Action,
		Replacement: replacement,
		Version:     1,
		IsActive:    true,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.wordRuleRepo.Create(c.Request.Context(), rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to create word rule"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": rule})
}

// UpdateWordRule updates an existing sensitive word rule.
func (h *ModerationHandler) UpdateWordRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid rule id"})
		return
	}

	var req request.UpdateSensitiveWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	rule, err := h.wordRuleRepo.GetByID(c.Request.Context(), ruleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "word rule not found"})
		return
	}

	if req.Pattern != "" {
		rule.Pattern = req.Pattern
	}
	if req.Action != "" {
		rule.Action = req.Action
	}
	if req.Replacement != "" {
		rule.Replacement = &req.Replacement
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	rule.UpdatedAt = time.Now().UTC()

	if err := h.wordRuleRepo.Update(c.Request.Context(), rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to update word rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// DeleteWordRule deletes a sensitive word rule by ID.
func (h *ModerationHandler) DeleteWordRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid rule id"})
		return
	}

	if err := h.wordRuleRepo.Delete(c.Request.Context(), ruleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to delete word rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "word rule deleted"})
}
