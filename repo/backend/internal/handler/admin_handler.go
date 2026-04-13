package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/config"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/job"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/audit"
	"github.com/localinsights/portal/internal/pkg/database"
	"github.com/localinsights/portal/internal/repository"
)

type AdminHandler struct {
	userRepo    repository.UserRepository
	auditRepo   repository.AuditLogRepository
	auditLogger *audit.Logger
	ipRuleRepo  repository.IPRuleRepository
	db          *database.DB
	cfg         *config.Config
}

func NewAdminHandler(
	userRepo repository.UserRepository,
	auditRepo repository.AuditLogRepository,
	ipRuleRepo repository.IPRuleRepository,
	db *database.DB,
	cfg *config.Config,
) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		auditRepo:   auditRepo,
		auditLogger: audit.NewLogger(auditRepo),
		ipRuleRepo:  ipRuleRepo,
		db:          db,
		cfg:         cfg,
	}
}

// ListUsers returns a paginated list of all users.
func (h *AdminHandler) ListUsers(c *gin.Context) {
	pag := getPagination(c)

	users, total, err := h.userRepo.List(c.Request.Context(), pag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     users,
		"total":    total,
		"page":     pag.Page,
		"per_page": pag.PerPage,
	})
}

// UpdateUserRole updates a user's role.
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid user id"})
		return
	}

	var req request.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	if err := h.userRepo.UpdateRole(c.Request.Context(), userID, model.Role(req.Role)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to update user role"})
		return
	}

	actorID := middleware.GetUserID(c)
	actorRole := middleware.GetUserRole(c)
	ip := c.ClientIP()
	reqID := c.GetHeader("X-Request-ID")
	targetType := "user"
	h.auditLogger.Log(c.Request.Context(), &actorID, &actorRole, "user.role_changed",
		&targetType, &userID, &ip, &reqID, map[string]any{"new_role": req.Role})

	c.JSON(http.StatusOK, gin.H{"msg": "user role updated"})
}

// UpdateUserStatus activates or deactivates a user.
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid user id"})
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	if err := h.userRepo.SetActive(c.Request.Context(), userID, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to update user status"})
		return
	}

	actorID := middleware.GetUserID(c)
	actorRole := middleware.GetUserRole(c)
	ip := c.ClientIP()
	reqID := c.GetHeader("X-Request-ID")
	targetType := "user"
	h.auditLogger.Log(c.Request.Context(), &actorID, &actorRole, "user.status_changed",
		&targetType, &userID, &ip, &reqID, map[string]any{"is_active": req.IsActive})

	c.JSON(http.StatusOK, gin.H{"msg": "user status updated"})
}

// ListAuditLogs returns a paginated list of audit log entries.
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	pag := getPagination(c)

	logs, total, err := h.auditRepo.List(c.Request.Context(), pag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     logs,
		"total":    total,
		"page":     pag.Page,
		"per_page": pag.PerPage,
	})
}

// ListIPRules returns all IP rules as a JSON array.
func (h *AdminHandler) ListIPRules(c *gin.Context) {
	rules, err := h.ipRuleRepo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list IP rules"})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// CreateIPRule creates a new IP rule.
func (h *AdminHandler) CreateIPRule(c *gin.Context) {
	var req struct {
		CIDR        string `json:"cidr" binding:"required"`
		RuleType    string `json:"rule_type" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	createdBy := middleware.GetUserID(c)
	if createdBy == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "authentication required"})
		return
	}

	if err := h.ipRuleRepo.Create(c.Request.Context(), req.CIDR, req.RuleType, req.Description, createdBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to create IP rule"})
		return
	}

	actorRole := middleware.GetUserRole(c)
	ip := c.ClientIP()
	reqID := c.GetHeader("X-Request-ID")
	targetType := "ip_rule"
	h.auditLogger.Log(c.Request.Context(), &createdBy, &actorRole, "ip_rule.created",
		&targetType, nil, &ip, &reqID, map[string]any{"cidr": req.CIDR, "rule_type": req.RuleType})

	c.JSON(http.StatusCreated, gin.H{"msg": "IP rule created"})
}

// DeleteIPRule deletes an IP rule by numeric ID.
func (h *AdminHandler) DeleteIPRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "invalid rule id"})
		return
	}

	if err := h.ipRuleRepo.Delete(c.Request.Context(), ruleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to delete IP rule"})
		return
	}

	actorID := middleware.GetUserID(c)
	actorRole := middleware.GetUserRole(c)
	ip := c.ClientIP()
	reqID := c.GetHeader("X-Request-ID")
	targetType := "ip_rule"
	h.auditLogger.Log(c.Request.Context(), &actorID, &actorRole, "ip_rule.deleted",
		&targetType, &ruleID, &ip, &reqID, nil)

	c.JSON(http.StatusOK, gin.H{"msg": "IP rule deleted"})
}

// TriggerBackup runs a backup job asynchronously and audits the action.
func (h *AdminHandler) TriggerBackup(c *gin.Context) {
	go func() {
		backupJob := job.NewBackupJob(h.db, h.cfg.Backup)
		backupJob.Run()
	}()
	uid := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	ip := c.ClientIP()
	reqID := c.GetHeader("X-Request-ID")
	h.auditLogger.Log(c.Request.Context(), &uid, &role, "backup_triggered", nil, nil, &ip, &reqID, nil)
	c.JSON(http.StatusAccepted, gin.H{"msg": "Backup triggered and running in background"})
}

// ListRecoveryDrills returns paginated recovery drill records.
func (h *AdminHandler) ListRecoveryDrills(c *gin.Context) {
	pag := getPagination(c)
	offset := pag.Offset()

	var drills []model.RecoveryDrill
	err := h.db.SelectContext(
		c.Request.Context(),
		&drills,
		"SELECT * FROM recovery_drills ORDER BY started_at DESC LIMIT ? OFFSET ?",
		pag.PerPage,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to list recovery drills"})
		return
	}

	var total int64
	_ = h.db.GetContext(c.Request.Context(), &total, "SELECT COUNT(*) FROM recovery_drills")

	c.JSON(http.StatusOK, gin.H{
		"data":     drills,
		"total":    total,
		"page":     pag.Page,
		"per_page": pag.PerPage,
	})
}

// TriggerRecoveryDrill runs a recovery drill asynchronously and audits the action.
func (h *AdminHandler) TriggerRecoveryDrill(c *gin.Context) {
	go func() {
		drillJob := job.NewRecoveryDrillJob(h.db, h.cfg)
		drillJob.Run()
	}()
	uid := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	ip := c.ClientIP()
	reqID := c.GetHeader("X-Request-ID")
	h.auditLogger.Log(c.Request.Context(), &uid, &role, "recovery_drill_triggered", nil, nil, &ip, &reqID, nil)
	c.JSON(http.StatusAccepted, gin.H{"msg": "Recovery drill triggered and running in background"})
}

// RebuildAnalytics runs an analytics ETL rebuild asynchronously and audits the action.
func (h *AdminHandler) RebuildAnalytics(c *gin.Context) {
	go func() {
		etlJob := job.NewAnalyticsETLJob(h.db)
		etlJob.Run()
	}()
	uid := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	ip := c.ClientIP()
	reqID := c.GetHeader("X-Request-ID")
	h.auditLogger.Log(c.Request.Context(), &uid, &role, "analytics_rebuild_triggered", nil, nil, &ip, &reqID, nil)
	c.JSON(http.StatusAccepted, gin.H{"msg": "Analytics rebuild triggered and running in background"})
}

// GetPerformanceMetrics returns recent performance metrics from monitoring_metrics.
func (h *AdminHandler) GetPerformanceMetrics(c *gin.Context) {
	var metrics []model.MonitoringMetric
	err := h.db.SelectContext(
		c.Request.Context(),
		&metrics,
		"SELECT * FROM monitoring_metrics ORDER BY recorded_at DESC LIMIT 100",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to get performance metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": metrics})
}

// GetErrorMetrics returns recent error-related audit log entries.
func (h *AdminHandler) GetErrorMetrics(c *gin.Context) {
	var logs []model.AuditLog
	err := h.db.SelectContext(
		c.Request.Context(),
		&logs,
		"SELECT * FROM audit_logs WHERE action LIKE 'error%' ORDER BY created_at DESC LIMIT 100",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to get error metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// GetSystemHealth checks database connectivity and returns overall system health.
func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	dbStatus := "connected"
	if err := h.db.PingContext(c.Request.Context()); err != nil {
		dbStatus = "disconnected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"db":     dbStatus,
		"uptime": time.Since(startTime).String(),
	})
}

// startTime records the time the process started for uptime calculation.
var startTime = time.Now()

// CaptureError receives a frontend error report and persists it as an audit log entry.
func (h *AdminHandler) CaptureError(c *gin.Context) {
	var req struct {
		Error string `json:"error" binding:"required"`
		URL   string `json:"url"`
		Stack string `json:"stack"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	ip := c.ClientIP()
	reqID := c.GetHeader("X-Request-ID")

	targetType := "frontend"
	var actorIDPtr *uint64
	var rolePtr *string
	if userID != 0 {
		actorIDPtr = &userID
		role := middleware.GetUserRole(c)
		if role != "" {
			rolePtr = &role
		}
	}

	h.auditLogger.Log(c.Request.Context(), actorIDPtr, rolePtr, "error.frontend",
		&targetType, nil, &ip, &reqID, map[string]any{
			"error": req.Error,
			"url":   req.URL,
			"stack": req.Stack,
		})

	c.JSON(http.StatusCreated, gin.H{"msg": "error captured"})
}
