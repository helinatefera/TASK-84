package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
)

type DashboardHandler struct {
	db *database.DB
}

func NewDashboardHandler(db *database.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// generateToken returns a cryptographically random hex-encoded token of the
// given byte length (the resulting string is twice as long).
func generateToken(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// GetDashboard queries analytics_aggregates with optional filters and returns
// aggregated data. Supports filtering by item_id, date range, sentiment, and
// keywords.
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()

	var filter request.DashboardFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Build the dynamic query.
	whereClauses := []string{"1 = 1"}
	args := []interface{}{}

	if filter.ItemID != "" {
		whereClauses = append(whereClauses, "aa.item_id = (SELECT id FROM items WHERE uuid = ?)")
		args = append(args, filter.ItemID)
	}
	if filter.StartDate != "" {
		whereClauses = append(whereClauses, "aa.period_start >= ?")
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != "" {
		whereClauses = append(whereClauses, "aa.period_start <= ?")
		args = append(args, filter.EndDate)
	}

	joinClause := ""
	if filter.Sentiment != "" {
		// Join review_sentiment to filter by sentiment label for the same item.
		joinClause = `
			INNER JOIN reviews r ON r.item_id = aa.item_id
			INNER JOIN review_sentiment rs ON rs.review_id = r.id`
		whereClauses = append(whereClauses, "rs.sentiment_label = ?")
		args = append(args, filter.Sentiment)
	}

	if filter.Keywords != "" {
		// Search for keyword matches in review text via the reviews table.
		if joinClause == "" {
			joinClause = " INNER JOIN reviews r ON r.item_id = aa.item_id"
		}
		whereClauses = append(whereClauses, "r.body LIKE ?")
		args = append(args, "%"+filter.Keywords+"%")
	}

	where := strings.Join(whereClauses, " AND ")

	// Count total.
	countQuery := fmt.Sprintf(
		`SELECT COUNT(DISTINCT aa.id) FROM analytics_aggregates aa %s WHERE %s`,
		joinClause, where)
	var total int64
	err := h.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to count aggregates")
		return
	}

	// Fetch page of data.
	dataQuery := fmt.Sprintf(
		`SELECT DISTINCT aa.* FROM analytics_aggregates aa %s WHERE %s
		ORDER BY aa.period_start DESC
		LIMIT ? OFFSET ?`,
		joinClause, where)
	dataArgs := append(args, perPage, offset)

	var aggregates []model.AnalyticsAggregate
	err = h.db.SelectContext(ctx, &aggregates, dataQuery, dataArgs...)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query aggregates")
		return
	}

	if aggregates == nil {
		aggregates = []model.AnalyticsAggregate{}
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        aggregates,
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	})
}

// ListSavedViews returns all saved views belonging to the authenticated user.
func (h *DashboardHandler) ListSavedViews(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := c.Request.Context()
	var views []model.SavedView
	err := h.db.SelectContext(ctx, &views,
		"SELECT * FROM saved_views WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query saved views")
		return
	}

	if views == nil {
		views = []model.SavedView{}
	}

	c.JSON(http.StatusOK, gin.H{"data": views})
}

// CreateSavedView creates a new saved view for the current user.
func (h *DashboardHandler) CreateSavedView(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req request.CreateSavedViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	ctx := c.Request.Context()
	viewUUID := uuid.New().String()

	_, err := h.db.ExecContext(ctx,
		`INSERT INTO saved_views (uuid, user_id, name, filter_config)
		VALUES (?, ?, ?, ?)`,
		viewUUID, userID, req.Name, req.FilterConfig)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create saved view")
		return
	}

	var view model.SavedView
	err = h.db.GetContext(ctx, &view,
		"SELECT * FROM saved_views WHERE uuid = ?", viewUUID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch created view")
		return
	}

	c.JSON(http.StatusCreated, view)
}

// UpdateSavedView updates a saved view by UUID after verifying ownership.
func (h *DashboardHandler) UpdateSavedView(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	viewUUID := c.Param("id")
	if viewUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "View ID is required")
		return
	}

	var req request.UpdateSavedViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	ctx := c.Request.Context()

	// Resolve and verify ownership.
	var view model.SavedView
	err := h.db.GetContext(ctx, &view,
		"SELECT * FROM saved_views WHERE uuid = ?", viewUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Saved view not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query saved view")
		return
	}

	if view.UserID != userID {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "You do not own this saved view")
		return
	}

	// Build dynamic update.
	setClauses := []string{}
	setArgs := []interface{}{}

	if req.Name != "" {
		setClauses = append(setClauses, "name = ?")
		setArgs = append(setArgs, req.Name)
	}
	if req.FilterConfig != nil {
		setClauses = append(setClauses, "filter_config = ?")
		setArgs = append(setArgs, req.FilterConfig)
	}

	if len(setClauses) == 0 {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "No fields to update")
		return
	}

	setArgs = append(setArgs, view.ID)
	query := fmt.Sprintf("UPDATE saved_views SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	_, err = h.db.ExecContext(ctx, query, setArgs...)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update saved view")
		return
	}

	// Return the updated view.
	var updated model.SavedView
	err = h.db.GetContext(ctx, &updated,
		"SELECT * FROM saved_views WHERE id = ?", view.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch updated view")
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteSavedView deletes a saved view by UUID after verifying ownership.
func (h *DashboardHandler) DeleteSavedView(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	viewUUID := c.Param("id")
	if viewUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "View ID is required")
		return
	}

	ctx := c.Request.Context()

	var view model.SavedView
	err := h.db.GetContext(ctx, &view,
		"SELECT * FROM saved_views WHERE uuid = ?", viewUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Saved view not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query saved view")
		return
	}

	if view.UserID != userID {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "You do not own this saved view")
		return
	}

	_, err = h.db.ExecContext(ctx,
		"DELETE FROM saved_views WHERE id = ?", view.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete saved view")
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Saved view deleted"})
}

// CreateShareLink generates a random 64-character token for sharing a saved
// view and inserts a share_links row with a 7-day expiry.
func (h *DashboardHandler) CreateShareLink(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	viewUUID := c.Param("id")
	if viewUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "View ID is required")
		return
	}

	ctx := c.Request.Context()

	// Resolve saved view.
	var view model.SavedView
	err := h.db.GetContext(ctx, &view,
		"SELECT * FROM saved_views WHERE uuid = ?", viewUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Saved view not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query saved view")
		return
	}

	if view.UserID != userID {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "You do not own this saved view")
		return
	}

	token := generateToken(32) // 32 bytes = 64 hex characters
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)

	_, err = h.db.ExecContext(ctx,
		`INSERT INTO share_links (token, saved_view_id, created_by, expires_at)
		VALUES (?, ?, ?, ?)`,
		token, view.ID, userID, expiresAt)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create share link")
		return
	}

	var link model.ShareLink
	err = h.db.GetContext(ctx, &link,
		"SELECT * FROM share_links WHERE token = ?", token)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch share link")
		return
	}

	c.JSON(http.StatusCreated, link)
}

// RevokeShareLink marks all active share links for a saved view as revoked.
func (h *DashboardHandler) RevokeShareLink(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	viewUUID := c.Param("id")
	if viewUUID == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "View ID is required")
		return
	}

	ctx := c.Request.Context()

	var view model.SavedView
	err := h.db.GetContext(ctx, &view,
		"SELECT * FROM saved_views WHERE uuid = ?", viewUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Saved view not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query saved view")
		return
	}

	if view.UserID != userID {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "You do not own this saved view")
		return
	}

	_, err = h.db.ExecContext(ctx,
		"UPDATE share_links SET is_revoked = 1 WHERE saved_view_id = ? AND is_revoked = 0",
		view.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to revoke share link")
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Share link revoked"})
}

// cloneRequest is the request body for CloneSavedView.
type cloneRequest struct {
	SourceToken string `json:"source_token" binding:"required"`
}

// CloneSavedView copies a shared view's filter_config into a new saved view
// owned by the current user. The source is located via a share token.
func (h *DashboardHandler) CloneSavedView(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req cloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	ctx := c.Request.Context()

	// Look up the share link.
	var link model.ShareLink
	err := h.db.GetContext(ctx, &link,
		"SELECT * FROM share_links WHERE token = ?", req.SourceToken)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Share link not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query share link")
		return
	}

	if link.IsRevoked {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "Share link has been revoked")
		return
	}
	if time.Now().UTC().After(link.ExpiresAt) {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "Share link has expired")
		return
	}

	// Fetch source view.
	var sourceView model.SavedView
	err = h.db.GetContext(ctx, &sourceView,
		"SELECT * FROM saved_views WHERE id = ?", link.SavedViewID)
	if err != nil {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "Source saved view not found")
		return
	}

	// Create cloned view.
	viewUUID := uuid.New().String()
	clonedName := sourceView.Name + " (cloned)"

	_, err = h.db.ExecContext(ctx,
		`INSERT INTO saved_views (uuid, user_id, name, filter_config)
		VALUES (?, ?, ?, ?)`,
		viewUUID, userID, clonedName, sourceView.FilterConfig)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to clone saved view")
		return
	}

	var cloned model.SavedView
	err = h.db.GetContext(ctx, &cloned,
		"SELECT * FROM saved_views WHERE uuid = ?", viewUUID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch cloned view")
		return
	}

	c.JSON(http.StatusCreated, cloned)
}

// GetSharedView looks up a share link by token, verifies it is not expired or
// revoked, and returns the associated saved view data in read-only form.
func (h *DashboardHandler) GetSharedView(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Token is required")
		return
	}

	ctx := c.Request.Context()

	var link model.ShareLink
	err := h.db.GetContext(ctx, &link,
		"SELECT * FROM share_links WHERE token = ?", token)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Share link not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query share link")
		return
	}

	if link.IsRevoked {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "Share link has been revoked")
		return
	}
	if time.Now().UTC().After(link.ExpiresAt) {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "Share link has expired")
		return
	}

	var view model.SavedView
	err = h.db.GetContext(ctx, &view,
		"SELECT * FROM saved_views WHERE id = ?", link.SavedViewID)
	if err != nil {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "Saved view not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            view.UUID,
		"name":          view.Name,
		"filter_config": view.FilterConfig,
		"created_at":    view.CreatedAt,
		"expires_at":    link.ExpiresAt,
		"read_only":     true,
	})
}

// GetSharedViewData validates a share token and returns all dashboard data
// (aggregates + visualizations) in a single response so that non-analyst users
// can view shared dashboards without requiring the analyst role.
func (h *DashboardHandler) GetSharedViewData(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Token is required")
		return
	}

	ctx := c.Request.Context()

	// Validate the share link.
	var link model.ShareLink
	err := h.db.GetContext(ctx, &link, "SELECT * FROM share_links WHERE token = ?", token)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Share link not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query share link")
		return
	}
	if link.IsRevoked {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "Share link has been revoked")
		return
	}
	if time.Now().UTC().After(link.ExpiresAt) {
		respondError(c, http.StatusForbidden, "FORBIDDEN", "Share link has expired")
		return
	}

	// Fetch the saved view to get filter_config.
	var view model.SavedView
	err = h.db.GetContext(ctx, &view, "SELECT * FROM saved_views WHERE id = ?", link.SavedViewID)
	if err != nil {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "Saved view not found")
		return
	}

	// Parse the filter config to build query params.
	type filterCfg struct {
		ItemID    string `json:"item_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Sentiment string `json:"sentiment"`
		Keywords  string `json:"keywords"`
	}
	var fc filterCfg
	_ = json.Unmarshal(view.FilterConfig, &fc)

	// --- Build dashboard aggregates query (mirrors GetDashboard) ---
	whereClauses := []string{"1 = 1"}
	args := []interface{}{}
	if fc.ItemID != "" {
		whereClauses = append(whereClauses, "aa.item_id = (SELECT id FROM items WHERE uuid = ?)")
		args = append(args, fc.ItemID)
	}
	if fc.StartDate != "" {
		whereClauses = append(whereClauses, "aa.period_start >= ?")
		args = append(args, fc.StartDate)
	}
	if fc.EndDate != "" {
		whereClauses = append(whereClauses, "aa.period_start <= ?")
		args = append(args, fc.EndDate)
	}
	joinClause := ""
	if fc.Sentiment != "" {
		joinClause = " INNER JOIN reviews r ON r.item_id = aa.item_id INNER JOIN review_sentiment rs ON rs.review_id = r.id"
		whereClauses = append(whereClauses, "rs.sentiment_label = ?")
		args = append(args, fc.Sentiment)
	}
	if fc.Keywords != "" {
		if joinClause == "" {
			joinClause = " INNER JOIN reviews r ON r.item_id = aa.item_id"
		}
		whereClauses = append(whereClauses, "r.body LIKE ?")
		args = append(args, "%"+fc.Keywords+"%")
	}
	where := strings.Join(whereClauses, " AND ")

	dataQuery := fmt.Sprintf(
		`SELECT DISTINCT aa.* FROM analytics_aggregates aa %s WHERE %s
		ORDER BY aa.period_start DESC LIMIT 100`, joinClause, where)

	var aggregates []model.AnalyticsAggregate
	_ = h.db.SelectContext(ctx, &aggregates, dataQuery, args...)
	if aggregates == nil {
		aggregates = []model.AnalyticsAggregate{}
	}

	// --- Build visualization queries ---
	// Keywords
	kwClauses := []string{"1 = 1"}
	var kwArgs []interface{}
	kwJoin := ""
	if fc.ItemID != "" {
		kwClauses = append(kwClauses, "r.item_id = (SELECT id FROM items WHERE uuid = ?)")
		kwArgs = append(kwArgs, fc.ItemID)
	}
	if fc.StartDate != "" {
		kwClauses = append(kwClauses, "r.created_at >= ?")
		kwArgs = append(kwArgs, fc.StartDate)
	}
	if fc.EndDate != "" {
		kwClauses = append(kwClauses, "r.created_at <= ?")
		kwArgs = append(kwArgs, fc.EndDate)
	}
	if fc.Sentiment != "" {
		kwJoin = " INNER JOIN review_sentiment rs_f ON rs_f.review_id = r.id"
		kwClauses = append(kwClauses, "rs_f.sentiment_label = ?")
		kwArgs = append(kwArgs, fc.Sentiment)
	}
	if fc.Keywords != "" {
		kwClauses = append(kwClauses, "r.body LIKE ?")
		kwArgs = append(kwArgs, "%"+fc.Keywords+"%")
	}
	kwWhere := strings.Join(kwClauses, " AND ")

	var keywords []keywordRow
	_ = h.db.SelectContext(ctx, &keywords, fmt.Sprintf(
		`SELECT rk.keyword, SUM(rk.weight) AS total_weight FROM review_keywords rk
		INNER JOIN reviews r ON r.id = rk.review_id %s WHERE %s
		GROUP BY rk.keyword ORDER BY total_weight DESC LIMIT 100`, kwJoin, kwWhere), kwArgs...)
	if keywords == nil {
		keywords = []keywordRow{}
	}

	var topics []topicRow
	_ = h.db.SelectContext(ctx, &topics, fmt.Sprintf(
		`SELECT rt.topic, AVG(rt.confidence) AS avg_confidence, COUNT(*) AS cnt FROM review_topics rt
		INNER JOIN reviews r ON r.id = rt.review_id %s WHERE %s
		GROUP BY rt.topic ORDER BY cnt DESC LIMIT 50`, kwJoin, kwWhere), kwArgs...)
	if topics == nil {
		topics = []topicRow{}
	}

	var sentiment []sentimentRow
	_ = h.db.SelectContext(ctx, &sentiment, fmt.Sprintf(
		`SELECT rs.sentiment_label, COUNT(*) AS cnt, AVG(rs.confidence) AS avg_confidence
		FROM review_sentiment rs INNER JOIN reviews r ON r.id = rs.review_id %s WHERE %s
		GROUP BY rs.sentiment_label ORDER BY cnt DESC`, kwJoin, kwWhere), kwArgs...)
	if sentiment == nil {
		sentiment = []sentimentRow{}
	}

	// Co-occurrence (simpler filter — no review join)
	coClauses := []string{"1 = 1"}
	var coArgs []interface{}
	if fc.ItemID != "" {
		coClauses = append(coClauses, "item_id = (SELECT id FROM items WHERE uuid = ?)")
		coArgs = append(coArgs, fc.ItemID)
	}
	if fc.StartDate != "" {
		coClauses = append(coClauses, "period_start >= ?")
		coArgs = append(coArgs, fc.StartDate)
	}
	if fc.EndDate != "" {
		coClauses = append(coClauses, "period_start <= ?")
		coArgs = append(coArgs, fc.EndDate)
	}

	var cooccurrences []model.CooccurrenceTerm
	_ = h.db.SelectContext(ctx, &cooccurrences, fmt.Sprintf(
		`SELECT * FROM cooccurrence_terms WHERE %s ORDER BY frequency DESC LIMIT 100`,
		strings.Join(coClauses, " AND ")), coArgs...)
	if cooccurrences == nil {
		cooccurrences = []model.CooccurrenceTerm{}
	}

	c.JSON(http.StatusOK, gin.H{
		"filter_config":  view.FilterConfig,
		"dashboard":      gin.H{"data": aggregates},
		"keywords":       gin.H{"data": keywords},
		"topics":         gin.H{"data": topics},
		"sentiment":      gin.H{"data": sentiment},
		"cooccurrences":  gin.H{"data": cooccurrences},
	})
}

// --- Visualization endpoints ---

// vizFilter builds a WHERE clause and optional joins against a reviews table
// aliased as "r", applying the same filter set as the dashboard. It returns
// the join clause, WHERE clause (always starts with "1 = 1"), and args.
func vizFilter(c *gin.Context) (joins string, where string, args []interface{}) {
	clauses := []string{"1 = 1"}

	if v := c.Query("item_id"); v != "" {
		clauses = append(clauses, "r.item_id = (SELECT id FROM items WHERE uuid = ?)")
		args = append(args, v)
	}
	if v := c.Query("start_date"); v != "" {
		clauses = append(clauses, "r.created_at >= ?")
		args = append(args, v)
	}
	if v := c.Query("end_date"); v != "" {
		clauses = append(clauses, "r.created_at <= ?")
		args = append(args, v)
	}
	if v := c.Query("sentiment"); v != "" {
		joins = " INNER JOIN review_sentiment rs_f ON rs_f.review_id = r.id"
		clauses = append(clauses, "rs_f.sentiment_label = ?")
		args = append(args, v)
	}
	if v := c.Query("keywords"); v != "" {
		clauses = append(clauses, "r.body LIKE ?")
		args = append(args, "%"+v+"%")
	}

	where = strings.Join(clauses, " AND ")
	return
}

// keywordRow represents an aggregated keyword with its total weight.
type keywordRow struct {
	Keyword string  `db:"keyword" json:"keyword"`
	Weight  float64 `db:"total_weight" json:"weight"`
}

// GetKeywords returns aggregated review keywords for the word cloud visualization.
// Accepts the same filters as the dashboard: item_id, start_date, end_date, sentiment, keywords.
func (h *DashboardHandler) GetKeywords(c *gin.Context) {
	ctx := c.Request.Context()
	extraJoin, where, args := vizFilter(c)

	q := fmt.Sprintf(
		`SELECT rk.keyword, SUM(rk.weight) AS total_weight
		FROM review_keywords rk
		INNER JOIN reviews r ON r.id = rk.review_id
		%s
		WHERE %s
		GROUP BY rk.keyword
		ORDER BY total_weight DESC
		LIMIT 100`, extraJoin, where)

	var rows []keywordRow
	err := h.db.SelectContext(ctx, &rows, q, args...)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query keywords")
		return
	}
	if rows == nil {
		rows = []keywordRow{}
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// topicRow represents an aggregated topic with its average confidence.
type topicRow struct {
	Topic      string  `db:"topic" json:"topic"`
	Confidence float64 `db:"avg_confidence" json:"confidence"`
	Count      int64   `db:"cnt" json:"count"`
}

// GetTopics returns aggregated review topics for topic distribution visualization.
// Accepts the same filters as the dashboard.
func (h *DashboardHandler) GetTopics(c *gin.Context) {
	ctx := c.Request.Context()
	extraJoin, where, args := vizFilter(c)

	q := fmt.Sprintf(
		`SELECT rt.topic, AVG(rt.confidence) AS avg_confidence, COUNT(*) AS cnt
		FROM review_topics rt
		INNER JOIN reviews r ON r.id = rt.review_id
		%s
		WHERE %s
		GROUP BY rt.topic
		ORDER BY cnt DESC
		LIMIT 50`, extraJoin, where)

	var rows []topicRow
	err := h.db.SelectContext(ctx, &rows, q, args...)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query topics")
		return
	}
	if rows == nil {
		rows = []topicRow{}
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GetCooccurrences returns term co-occurrence pairs for network visualization.
// Accepts item_id, start_date, end_date filters (period_start scoped).
func (h *DashboardHandler) GetCooccurrences(c *gin.Context) {
	ctx := c.Request.Context()

	clauses := []string{"1 = 1"}
	var args []interface{}

	if v := c.Query("item_id"); v != "" {
		clauses = append(clauses, "item_id = (SELECT id FROM items WHERE uuid = ?)")
		args = append(args, v)
	}
	if v := c.Query("start_date"); v != "" {
		clauses = append(clauses, "period_start >= ?")
		args = append(args, v)
	}
	if v := c.Query("end_date"); v != "" {
		clauses = append(clauses, "period_start <= ?")
		args = append(args, v)
	}

	q := fmt.Sprintf(
		`SELECT * FROM cooccurrence_terms
		WHERE %s
		ORDER BY frequency DESC
		LIMIT 100`, strings.Join(clauses, " AND "))

	var rows []model.CooccurrenceTerm
	err := h.db.SelectContext(ctx, &rows, q, args...)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query co-occurrences")
		return
	}
	if rows == nil {
		rows = []model.CooccurrenceTerm{}
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// sentimentRow represents aggregated sentiment distribution.
type sentimentRow struct {
	SentimentLabel string  `db:"sentiment_label" json:"sentiment_label"`
	Count          int64   `db:"cnt" json:"count"`
	AvgConfidence  float64 `db:"avg_confidence" json:"avg_confidence"`
}

// GetSentimentDistribution returns sentiment distribution for the heatmap visualization.
// Accepts the same filters as the dashboard.
func (h *DashboardHandler) GetSentimentDistribution(c *gin.Context) {
	ctx := c.Request.Context()

	clauses := []string{"1 = 1"}
	joins := ""
	var args []interface{}

	if v := c.Query("item_id"); v != "" {
		clauses = append(clauses, "r.item_id = (SELECT id FROM items WHERE uuid = ?)")
		args = append(args, v)
	}
	if v := c.Query("start_date"); v != "" {
		clauses = append(clauses, "r.created_at >= ?")
		args = append(args, v)
	}
	if v := c.Query("end_date"); v != "" {
		clauses = append(clauses, "r.created_at <= ?")
		args = append(args, v)
	}
	if v := c.Query("keywords"); v != "" {
		clauses = append(clauses, "r.body LIKE ?")
		args = append(args, "%"+v+"%")
	}

	q := fmt.Sprintf(
		`SELECT rs.sentiment_label, COUNT(*) AS cnt, AVG(rs.confidence) AS avg_confidence
		FROM review_sentiment rs
		INNER JOIN reviews r ON r.id = rs.review_id
		%s
		WHERE %s
		GROUP BY rs.sentiment_label
		ORDER BY cnt DESC`, joins, strings.Join(clauses, " AND "))

	var rows []sentimentRow
	err := h.db.SelectContext(ctx, &rows, q, args...)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query sentiment")
		return
	}
	if rows == nil {
		rows = []sentimentRow{}
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

