package handler

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
)

type ExperimentHandler struct {
	db *database.DB
}

func NewExperimentHandler(db *database.DB) *ExperimentHandler {
	return &ExperimentHandler{db: db}
}

// experimentWithVariants is the combined response for experiment endpoints.
type experimentWithVariants struct {
	model.Experiment
	Variants []model.ExperimentVariant `json:"variants"`
}

// assignVariant deterministically assigns a user to a variant using FNV-1a
// hashing. The bucket space is 10000, and each variant occupies a portion
// proportional to its traffic_pct.
func assignVariant(hashSalt string, userID uint64, variants []model.ExperimentVariant) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%s:%d", hashSalt, userID)))
	bucket := h.Sum32() % 10000
	var cumulative uint32
	for i, v := range variants {
		cumulative += uint32(v.TrafficPct * 100)
		if bucket < cumulative {
			return i
		}
	}
	return 0
}

// GetByID returns a single experiment by UUID with its variants.
func (h *ExperimentHandler) GetByID(c *gin.Context) {
	expUUID := c.Param("id")
	ctx := c.Request.Context()

	var exp model.Experiment
	err := h.db.GetContext(ctx, &exp, "SELECT * FROM experiments WHERE uuid = ?", expUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Experiment not found"})
		return
	}

	var variants []model.ExperimentVariant
	_ = h.db.SelectContext(ctx, &variants, "SELECT * FROM experiment_variants WHERE experiment_id = ? ORDER BY id", exp.ID)

	c.JSON(http.StatusOK, gin.H{
		"id":              exp.UUID,
		"name":            exp.Name,
		"description":     exp.Description,
		"status":          exp.Status,
		"hash_salt":       exp.HashSalt,
		"min_sample_size": exp.MinSampleSize,
		"started_at":      exp.StartedAt,
		"ended_at":        exp.EndedAt,
		"created_at":      exp.CreatedAt,
		"variants":        variants,
	})
}

// List returns all experiments with pagination.
func (h *ExperimentHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int64
	err := h.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM experiments")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to count experiments")
		return
	}

	var experiments []model.Experiment
	err = h.db.SelectContext(ctx, &experiments,
		`SELECT * FROM experiments ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		perPage, offset)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query experiments")
		return
	}

	if experiments == nil {
		experiments = []model.Experiment{}
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        experiments,
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	})
}

// Create inserts a new experiment and its variants. It validates that the
// variant traffic percentages sum to exactly 100.
func (h *ExperimentHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req request.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	// Validate traffic percentages sum to 100.
	var trafficSum float64
	for _, v := range req.Variants {
		trafficSum += v.TrafficPct
	}
	if trafficSum != 100 {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR",
			fmt.Sprintf("Variant traffic_pct must sum to 100 (got %.2f)", trafficSum))
		return
	}

	ctx := c.Request.Context()
	expUUID := uuid.New().String()
	hashSalt := uuid.New().String()

	minSample := req.MinSampleSize
	if minSample == 0 {
		minSample = 100
	}

	var expID uint64

	err := h.db.WithTx(ctx, func(txCtx context.Context) error {
		ext := h.db.ExtContext(txCtx)

		var slugPtr *string
		if req.Slug != "" {
			slugPtr = &req.Slug
		}

		result, err := ext.ExecContext(txCtx,
			`INSERT INTO experiments (uuid, name, slug, description, status, hash_salt, min_sample_size, created_by)
			VALUES (?, ?, ?, ?, 'draft', ?, ?, ?)`,
			expUUID, req.Name, slugPtr, req.Description, hashSalt, minSample, userID)
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		expID = uint64(id)

		for _, v := range req.Variants {
			_, err := ext.ExecContext(txCtx,
				`INSERT INTO experiment_variants (experiment_id, name, traffic_pct, config)
				VALUES (?, ?, ?, ?)`,
				expID, v.Name, v.TrafficPct, v.Config)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create experiment")
		return
	}

	// Return the full experiment with variants.
	var exp model.Experiment
	err = h.db.GetContext(ctx, &exp,
		"SELECT * FROM experiments WHERE id = ?", expID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch created experiment")
		return
	}

	var variants []model.ExperimentVariant
	err = h.db.SelectContext(ctx, &variants,
		"SELECT * FROM experiment_variants WHERE experiment_id = ?", expID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch variants")
		return
	}

	c.JSON(http.StatusCreated, experimentWithVariants{
		Experiment: exp,
		Variants:   variants,
	})
}

// Update modifies the name and/or description of an experiment.
func (h *ExperimentHandler) Update(c *gin.Context) {
	expUUID := c.Param("id")
	if expUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Experiment ID is required")
		return
	}

	var req request.UpdateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	ctx := c.Request.Context()

	var exp model.Experiment
	err := h.db.GetContext(ctx, &exp,
		"SELECT * FROM experiments WHERE uuid = ?", expUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Experiment not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query experiment")
		return
	}

	if req.Name != "" {
		exp.Name = req.Name
	}
	if req.Description != "" {
		exp.Description = &req.Description
	}

	_, err = h.db.ExecContext(ctx,
		"UPDATE experiments SET name = ?, description = ? WHERE id = ?",
		exp.Name, exp.Description, exp.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update experiment")
		return
	}

	// Re-fetch.
	err = h.db.GetContext(ctx, &exp,
		"SELECT * FROM experiments WHERE id = ?", exp.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch updated experiment")
		return
	}

	c.JSON(http.StatusOK, exp)
}

// UpdateTraffic adjusts variant traffic percentages for a running or paused experiment.
func (h *ExperimentHandler) UpdateTraffic(c *gin.Context) {
	expUUID := c.Param("id")
	if expUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Experiment ID is required")
		return
	}

	var req request.UpdateCanaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	var trafficSum float64
	for _, v := range req.Variants {
		trafficSum += v.TrafficPct
	}
	if trafficSum != 100 {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR",
			fmt.Sprintf("Variant traffic_pct must sum to 100 (got %.2f)", trafficSum))
		return
	}

	ctx := c.Request.Context()

	var exp model.Experiment
	err := h.db.GetContext(ctx, &exp, "SELECT * FROM experiments WHERE uuid = ?", expUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Experiment not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query experiment")
		return
	}

	if exp.Status != model.ExperimentStatusRunning && exp.Status != model.ExperimentStatusPaused {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR",
			"Traffic can only be adjusted on running or paused experiments")
		return
	}

	err = h.db.WithTx(ctx, func(txCtx context.Context) error {
		ext := h.db.ExtContext(txCtx)
		for _, v := range req.Variants {
			_, err := ext.ExecContext(txCtx,
				"UPDATE experiment_variants SET traffic_pct = ? WHERE experiment_id = ? AND name = ?",
				v.TrafficPct, exp.ID, v.Name)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update traffic")
		return
	}

	var variants []model.ExperimentVariant
	_ = h.db.SelectContext(ctx, &variants,
		"SELECT * FROM experiment_variants WHERE experiment_id = ? ORDER BY id ASC", exp.ID)

	c.JSON(http.StatusOK, gin.H{"variants": variants})
}

// transitionExperiment is a helper for Start, Pause, Complete, and Rollback.
// It validates the current status, updates to the target status, records
// status history, and optionally sets started_at or ended_at.
func (h *ExperimentHandler) transitionExperiment(
	c *gin.Context,
	allowedFrom []model.ExperimentStatus,
	newStatus model.ExperimentStatus,
	setStarted bool,
	setEnded bool,
) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	expUUID := c.Param("id")
	if expUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Experiment ID is required")
		return
	}

	ctx := c.Request.Context()

	var exp model.Experiment
	err := h.db.GetContext(ctx, &exp,
		"SELECT * FROM experiments WHERE uuid = ?", expUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Experiment not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query experiment")
		return
	}

	// Validate current status is in allowed list.
	allowed := false
	for _, s := range allowedFrom {
		if exp.Status == s {
			allowed = true
			break
		}
	}
	if !allowed {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR",
			fmt.Sprintf("Cannot transition from %s to %s", exp.Status, newStatus))
		return
	}

	oldStatus := exp.Status

	err = h.db.WithTx(ctx, func(txCtx context.Context) error {
		ext := h.db.ExtContext(txCtx)

		// Build the status update.
		if setStarted {
			_, err := ext.ExecContext(txCtx,
				"UPDATE experiments SET status = ?, started_at = NOW(3) WHERE id = ?",
				newStatus, exp.ID)
			if err != nil {
				return err
			}
		} else if setEnded {
			_, err := ext.ExecContext(txCtx,
				"UPDATE experiments SET status = ?, ended_at = NOW(3) WHERE id = ?",
				newStatus, exp.ID)
			if err != nil {
				return err
			}
		} else {
			_, err := ext.ExecContext(txCtx,
				"UPDATE experiments SET status = ? WHERE id = ?",
				newStatus, exp.ID)
			if err != nil {
				return err
			}
		}

		// Insert status history.
		_, err := ext.ExecContext(txCtx,
			`INSERT INTO experiment_status_history
				(experiment_id, old_status, new_status, changed_by, changed_at)
			VALUES (?, ?, ?, ?, NOW(3))`,
			exp.ID, oldStatus, newStatus, userID)
		return err
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to transition experiment")
		return
	}

	// Re-fetch.
	err = h.db.GetContext(ctx, &exp,
		"SELECT * FROM experiments WHERE id = ?", exp.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch experiment")
		return
	}

	c.JSON(http.StatusOK, exp)
}

// Start transitions an experiment from draft to running.
func (h *ExperimentHandler) Start(c *gin.Context) {
	h.transitionExperiment(c,
		[]model.ExperimentStatus{model.ExperimentStatusDraft},
		model.ExperimentStatusRunning,
		true,  // setStarted
		false, // setEnded
	)
}

// Pause transitions an experiment from running to paused.
func (h *ExperimentHandler) Pause(c *gin.Context) {
	h.transitionExperiment(c,
		[]model.ExperimentStatus{model.ExperimentStatusRunning},
		model.ExperimentStatusPaused,
		false,
		false,
	)
}

// Complete transitions an experiment from running or paused to completed.
func (h *ExperimentHandler) Complete(c *gin.Context) {
	h.transitionExperiment(c,
		[]model.ExperimentStatus{model.ExperimentStatusRunning, model.ExperimentStatusPaused},
		model.ExperimentStatusCompleted,
		false,
		true, // setEnded
	)
}

// Rollback transitions an experiment from running or paused to rolled_back.
func (h *ExperimentHandler) Rollback(c *gin.Context) {
	h.transitionExperiment(c,
		[]model.ExperimentStatus{model.ExperimentStatusRunning, model.ExperimentStatusPaused},
		model.ExperimentStatusRolledBack,
		false,
		true, // setEnded
	)
}

// variantResult holds per-variant stats returned by GetResults.
type variantResult struct {
	Name       string `json:"name" db:"name"`
	SampleSize int64  `json:"sample_size"`
	Exposures  int64  `json:"exposures"`
}

// GetResults computes per-variant stats (assignments and exposures) and
// determines a confidence_state based on the experiment's min_sample_size.
func (h *ExperimentHandler) GetResults(c *gin.Context) {
	expUUID := c.Param("id")
	if expUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Experiment ID is required")
		return
	}

	ctx := c.Request.Context()

	var exp model.Experiment
	err := h.db.GetContext(ctx, &exp,
		"SELECT * FROM experiments WHERE uuid = ?", expUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Experiment not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query experiment")
		return
	}

	// Get all variants.
	var variants []model.ExperimentVariant
	err = h.db.SelectContext(ctx, &variants,
		"SELECT * FROM experiment_variants WHERE experiment_id = ?", exp.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query variants")
		return
	}

	results := make([]variantResult, 0, len(variants))
	var totalAssignments int64

	for _, v := range variants {
		var sampleSize int64
		err := h.db.GetContext(ctx, &sampleSize,
			"SELECT COUNT(*) FROM experiment_assignments WHERE experiment_id = ? AND variant_id = ?",
			exp.ID, v.ID)
		if err != nil {
			sampleSize = 0
		}

		var exposures int64
		err = h.db.GetContext(ctx, &exposures,
			"SELECT COUNT(*) FROM experiment_exposures WHERE experiment_id = ? AND variant_id = ?",
			exp.ID, v.ID)
		if err != nil {
			exposures = 0
		}

		totalAssignments += sampleSize

		results = append(results, variantResult{
			Name:       v.Name,
			SampleSize: sampleSize,
			Exposures:  exposures,
		})
	}

	// Determine confidence state.
	confidenceState := "insufficient_data"
	minSample := int64(exp.MinSampleSize)

	if totalAssignments >= minSample {
		// Check if all variants have at least some assignments.
		allHaveData := true
		for _, r := range results {
			if r.SampleSize == 0 {
				allHaveData = false
				break
			}
		}

		if !allHaveData {
			confidenceState = "monitoring"
		} else {
			// Simple heuristic: if the best-performing variant (by exposure
			// rate) is >10% better, recommend keeping. Otherwise recommend
			// rollback.
			var bestRate float64
			var worstRate float64 = 1.0
			for _, r := range results {
				if r.SampleSize > 0 {
					rate := float64(r.Exposures) / float64(r.SampleSize)
					if rate > bestRate {
						bestRate = rate
					}
					if rate < worstRate {
						worstRate = rate
					}
				}
			}

			if bestRate > worstRate*1.1 {
				confidenceState = "recommend_keep"
			} else {
				confidenceState = "recommend_rollback"
			}
		}
	} else if totalAssignments > 0 {
		confidenceState = "monitoring"
	}

	c.JSON(http.StatusOK, gin.H{
		"variants":         results,
		"confidence_state": confidenceState,
	})
}

// GetAssignment returns or creates a deterministic variant assignment for the
// current user. An existing assignment is returned if one exists; otherwise a
// new one is computed via FNV-1a and persisted.
func (h *ExperimentHandler) GetAssignment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	expID := c.Param("exp_id")
	if expID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Experiment ID is required")
		return
	}

	ctx := c.Request.Context()

	// Resolve by UUID first; fall back to slug for human-readable identifiers.
	var exp model.Experiment
	err := h.db.GetContext(ctx, &exp,
		"SELECT * FROM experiments WHERE uuid = ? OR slug = ?", expID, expID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Experiment not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query experiment")
		return
	}

	if exp.Status != model.ExperimentStatusRunning {
		respondError(c, http.StatusBadRequest, "EXPERIMENT_NOT_ACTIVE", "Experiment is not currently running")
		return
	}

	// Check for existing assignment.
	var existing model.ExperimentAssignment
	err = h.db.GetContext(ctx, &existing,
		"SELECT * FROM experiment_assignments WHERE experiment_id = ? AND user_id = ?",
		exp.ID, userID)
	if err == nil {
		// Assignment already exists. Return the variant.
		var variant model.ExperimentVariant
		err = h.db.GetContext(ctx, &variant,
			"SELECT * FROM experiment_variants WHERE id = ?", existing.VariantID)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch variant")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"experiment_id": exp.UUID,
			"variant":       variant,
			"assigned_at":   existing.AssignedAt,
		})
		return
	}

	// No assignment yet -- compute one.
	var variants []model.ExperimentVariant
	err = h.db.SelectContext(ctx, &variants,
		"SELECT * FROM experiment_variants WHERE experiment_id = ? ORDER BY id ASC", exp.ID)
	if err != nil || len(variants) == 0 {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch variants")
		return
	}

	idx := assignVariant(exp.HashSalt, userID, variants)
	chosen := variants[idx]

	_, err = h.db.ExecContext(ctx,
		`INSERT INTO experiment_assignments (experiment_id, user_id, variant_id, assigned_at)
		VALUES (?, ?, ?, NOW(3))
		ON DUPLICATE KEY UPDATE id = id`,
		exp.ID, userID, chosen.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to persist assignment")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"experiment_id": exp.UUID,
		"variant":       chosen,
		"assigned_at":   time.Now().UTC(),
	})
}

// RecordExposure inserts an experiment exposure event for the current user.
func (h *ExperimentHandler) RecordExposure(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	expUUID := c.Param("id")
	if expUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Experiment ID is required")
		return
	}

	ctx := c.Request.Context()

	var exp model.Experiment
	err := h.db.GetContext(ctx, &exp,
		"SELECT * FROM experiments WHERE uuid = ?", expUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Experiment not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query experiment")
		return
	}

	if exp.Status != model.ExperimentStatusRunning {
		respondError(c, http.StatusBadRequest, "EXPERIMENT_NOT_ACTIVE", "Experiment is not currently running")
		return
	}

	// The user must have an assignment.
	var assignment model.ExperimentAssignment
	err = h.db.GetContext(ctx, &assignment,
		"SELECT * FROM experiment_assignments WHERE experiment_id = ? AND user_id = ?",
		exp.ID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "No assignment found for this user; call GetAssignment first")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query assignment")
		return
	}

	_, err = h.db.ExecContext(ctx,
		`INSERT INTO experiment_exposures (experiment_id, user_id, variant_id, exposed_at)
		VALUES (?, ?, ?, NOW(3))`,
		exp.ID, userID, assignment.VariantID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record exposure")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"msg": "Exposure recorded"})
}
