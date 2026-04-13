package handler

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math"
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

type AnalyticsHandler struct {
	db *database.DB
}

func NewAnalyticsHandler(db *database.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

// computeDedupHash produces a SHA-256 hash from stable event identity fields
// scoped to a 2-second time bucket so that truly identical events collapse
// while distinct events in the same batch do not.
func computeDedupHash(userID uint64, sessionID uint64, eventType string, itemID uint64, ts time.Time) string {
	bucket := ts.Unix() / 2
	raw := fmt.Sprintf("%d:%d:%s:%d:%d", userID, sessionID, eventType, itemID, bucket)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// IngestEvents accepts a batch of behavior events, deduplicates them, and
// persists each one together with an increment of the per-user hourly event
// counter.
func (h *AnalyticsHandler) IngestEvents(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req request.BatchEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	ctx := c.Request.Context()

	// Resolve session internal ID from the provided UUID, scoped to the
	// authenticated user to prevent cross-user event injection.
	var session model.AnalyticsSession
	err := h.db.GetContext(ctx, &session,
		"SELECT id, session_uuid FROM analytics_sessions WHERE session_uuid = ? AND user_id = ?",
		req.SessionUUID, userID)
	if err != nil {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "Session not found")
		return
	}

	ingested := 0
	now := time.Now().UTC()

	for _, ev := range req.Events {
		clientTS, err := time.Parse(time.RFC3339Nano, ev.ClientTS)
		if err != nil {
			clientTS = now
		}

		var itemID uint64
		if ev.ItemID != "" {
			parsed, err := strconv.ParseUint(ev.ItemID, 10, 64)
			if err == nil {
				itemID = parsed
			}
		}

		dedupHash := computeDedupHash(userID, session.ID, ev.EventType, itemID, clientTS)

		var dwellPtr *uint16
		if ev.DwellSeconds > 0 {
			dw := uint16(ev.DwellSeconds)
			dwellPtr = &dw
		}

		var itemIDPtr *uint64
		if itemID > 0 {
			itemIDPtr = &itemID
		}

		// INSERT with dedup: if the hash already exists, skip.
		result, err := h.db.ExecContext(ctx,
			`INSERT INTO behavior_events
				(session_id, user_id, event_type, item_id, dwell_seconds, event_data, client_ts, server_ts, dedup_hash)
			VALUES (?, ?, ?, ?, ?, ?, ?, NOW(3), ?)
			ON DUPLICATE KEY UPDATE id = id`,
			session.ID, userID, ev.EventType, itemIDPtr, dwellPtr, ev.EventData, clientTS, dedupHash)
		if err != nil {
			continue
		}

		rows, _ := result.RowsAffected()
		if rows > 0 {
			ingested++

			// Only increment the hourly event counter when the event was
			// actually inserted (not deduplicated) to avoid inflating
			// fraud-detection rate counts.
			_, _ = h.db.ExecContext(ctx,
				`INSERT INTO user_event_counts (user_id, hour_bucket, event_count)
				VALUES (?, DATE_FORMAT(NOW(), '%Y-%m-%d %H:00:00'), 1)
				ON DUPLICATE KEY UPDATE event_count = event_count + 1`,
				userID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"ingested": ingested})
}

// CreateSession inserts a new analytics session and returns the generated UUID.
func (h *AnalyticsHandler) CreateSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req request.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	ctx := c.Request.Context()
	sessionUUID := uuid.New().String()

	var itemIDPtr *uint64
	if req.ItemID != "" {
		parsed, err := strconv.ParseUint(req.ItemID, 10, 64)
		if err == nil {
			itemIDPtr = &parsed
		}
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddr := c.ClientIP()

	_, err := h.db.ExecContext(ctx,
		`INSERT INTO analytics_sessions
			(session_uuid, user_id, item_id, started_at, last_active_at, user_agent, ip_address)
		VALUES (?, ?, ?, NOW(3), NOW(3), ?, ?)`,
		sessionUUID, userID, itemIDPtr, userAgent, ipAddr)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create session")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"session_id": sessionUUID})
}

// Heartbeat touches the last_active_at timestamp for a session. The update is
// scoped to the authenticated user to prevent cross-user session manipulation.
func (h *AnalyticsHandler) Heartbeat(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	sessionUUID := c.Param("id")
	if sessionUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Session ID is required")
		return
	}

	ctx := c.Request.Context()
	result, err := h.db.ExecContext(ctx,
		"UPDATE analytics_sessions SET last_active_at = NOW(3) WHERE session_uuid = ? AND user_id = ?",
		sessionUUID, userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update session")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "Session not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

// GetSession returns session details by UUID.
func (h *AnalyticsHandler) GetSession(c *gin.Context) {
	sessionUUID := c.Param("id")
	if sessionUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Session ID is required")
		return
	}

	ctx := c.Request.Context()
	var session model.AnalyticsSession
	err := h.db.GetContext(ctx, &session,
		"SELECT * FROM analytics_sessions WHERE session_uuid = ?", sessionUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Session not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query session")
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetSessionTimeline returns all behavior events for a session ordered by
// server timestamp.
func (h *AnalyticsHandler) GetSessionTimeline(c *gin.Context) {
	sessionUUID := c.Param("id")
	if sessionUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Session ID is required")
		return
	}

	ctx := c.Request.Context()

	// Resolve session ID.
	var sessionID uint64
	err := h.db.GetContext(ctx, &sessionID,
		"SELECT id FROM analytics_sessions WHERE session_uuid = ?", sessionUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Session not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query session")
		return
	}

	var events []model.BehaviorEvent
	err = h.db.SelectContext(ctx, &events,
		"SELECT * FROM behavior_events WHERE session_id = ? ORDER BY server_ts ASC", sessionID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query events")
		return
	}

	if events == nil {
		events = []model.BehaviorEvent{}
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

// ListAggregateSessions returns analytics sessions that fall within a given
// aggregate period for an item. Expects query params item_id and period_start.
func (h *AnalyticsHandler) ListAggregateSessions(c *gin.Context) {
	itemUUID := c.Query("item_id")
	periodStart := c.Query("period_start")
	if itemUUID == "" || periodStart == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "item_id and period_start are required")
		return
	}

	// Resolve item identifier — accept UUID or numeric ID.
	ctx := c.Request.Context()
	var itemID uint64
	if numID, parseErr := strconv.ParseUint(itemUUID, 10, 64); parseErr == nil {
		itemID = numID
	} else if err := h.db.GetContext(ctx, &itemID, "SELECT id FROM items WHERE uuid = ?", itemUUID); err != nil {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "Item not found")
		return
	}

	startTime, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		// Try full timestamp format as well.
		startTime, err = time.Parse(time.RFC3339, periodStart)
		if err != nil {
			respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid period_start format")
			return
		}
	}
	endTime := startTime.AddDate(0, 0, 1)

	var sessions []model.AnalyticsSession
	err = h.db.SelectContext(ctx, &sessions,
		`SELECT * FROM analytics_sessions
		WHERE item_id = ? AND started_at >= ? AND started_at < ?
		ORDER BY started_at DESC
		LIMIT 50`,
		itemID, startTime, endTime)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query sessions")
		return
	}

	if sessions == nil {
		sessions = []model.AnalyticsSession{}
	}

	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

// GetScoringWeights returns the active scoring weights configuration.
func (h *AnalyticsHandler) GetScoringWeights(c *gin.Context) {
	ctx := c.Request.Context()

	var weights model.ScoringWeights
	err := h.db.GetContext(ctx, &weights,
		"SELECT * FROM scoring_weights WHERE is_active = 1 LIMIT 1")
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "No active scoring weights found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query scoring weights")
		return
	}

	c.JSON(http.StatusOK, weights)
}

// UpdateScoringWeights validates that the weights sum to ~1.0, increments the
// version, updates the active row, and logs a version history entry.
func (h *AnalyticsHandler) UpdateScoringWeights(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req request.UpdateScoringWeightsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	sum := req.ImpressionW + req.ClickW + req.DwellW + req.FavoriteW + req.ShareW + req.CommentW
	if math.Abs(sum-1.0) > 0.01 {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR",
			fmt.Sprintf("Weights must sum to 1.0 (got %.4f)", sum))
		return
	}

	ctx := c.Request.Context()

	// Fetch current active weights.
	var current model.ScoringWeights
	err := h.db.GetContext(ctx, &current,
		"SELECT * FROM scoring_weights WHERE is_active = 1 LIMIT 1")
	if err != nil {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "No active scoring weights found")
		return
	}

	newVersion := current.Version + 1

	err = h.db.WithTx(ctx, func(txCtx context.Context) error {
		ext := h.db.ExtContext(txCtx)

		// Update the active row.
		_, err := ext.ExecContext(txCtx,
			`UPDATE scoring_weights
			SET impression_w = ?, click_w = ?, dwell_w = ?, favorite_w = ?, share_w = ?, comment_w = ?,
				version = ?, updated_by = ?
			WHERE id = ?`,
			req.ImpressionW, req.ClickW, req.DwellW, req.FavoriteW, req.ShareW, req.CommentW,
			newVersion, userID, current.ID)
		if err != nil {
			return err
		}

		// Insert version history.
		_, err = ext.ExecContext(txCtx,
			`INSERT INTO scoring_weight_versions
				(weight_id, version, impression_w, click_w, dwell_w, favorite_w, share_w, comment_w, changed_by, effective_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3))`,
			current.ID, newVersion,
			req.ImpressionW, req.ClickW, req.DwellW, req.FavoriteW, req.ShareW, req.CommentW,
			userID)
		return err
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update scoring weights")
		return
	}

	// Return updated weights.
	var updated model.ScoringWeights
	err = h.db.GetContext(ctx, &updated,
		"SELECT * FROM scoring_weights WHERE id = ?", current.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch updated weights")
		return
	}

	c.JSON(http.StatusOK, updated)
}

// scoringWeightVersion represents a single version row from the
// scoring_weight_versions table, used for the history endpoint.
type scoringWeightVersion struct {
	ID          uint64    `db:"id" json:"-"`
	WeightID    uint64    `db:"weight_id" json:"weight_id"`
	Version     uint32    `db:"version" json:"version"`
	ImpressionW float64   `db:"impression_w" json:"impression_w"`
	ClickW      float64   `db:"click_w" json:"click_w"`
	DwellW      float64   `db:"dwell_w" json:"dwell_w"`
	FavoriteW   float64   `db:"favorite_w" json:"favorite_w"`
	ShareW      float64   `db:"share_w" json:"share_w"`
	CommentW    float64   `db:"comment_w" json:"comment_w"`
	ChangedBy   uint64    `db:"changed_by" json:"changed_by"`
	EffectiveAt time.Time `db:"effective_at" json:"effective_at"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// GetScoringWeightsHistory returns version history for the active scoring
// weights, ordered by version descending with pagination.
func (h *AnalyticsHandler) GetScoringWeightsHistory(c *gin.Context) {
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
	err := h.db.GetContext(ctx, &total,
		"SELECT COUNT(*) FROM scoring_weight_versions")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to count versions")
		return
	}

	var versions []scoringWeightVersion
	err = h.db.SelectContext(ctx, &versions,
		`SELECT * FROM scoring_weight_versions
		ORDER BY version DESC
		LIMIT ? OFFSET ?`, perPage, offset)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query version history")
		return
	}

	if versions == nil {
		versions = []scoringWeightVersion{}
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        versions,
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	})
}
