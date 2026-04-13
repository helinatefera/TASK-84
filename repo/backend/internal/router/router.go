package router

import (
	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/handler"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/pkg/database"
	"github.com/localinsights/portal/internal/pkg/jwt"
	"github.com/localinsights/portal/internal/config"
)

type Handlers struct {
	Auth       *handler.AuthHandler
	User       *handler.UserHandler
	Item       *handler.ItemHandler
	Review     *handler.ReviewHandler
	Image      *handler.ImageHandler
	QA         *handler.QAHandler
	Favorite   *handler.FavoriteHandler
	Wishlist   *handler.WishlistHandler
	Moderation *handler.ModerationHandler
	Analytics  *handler.AnalyticsHandler
	Dashboard  *handler.DashboardHandler
	Experiment *handler.ExperimentHandler
	Notification *handler.NotificationHandler
	Admin      *handler.AdminHandler
	Captcha    *handler.CaptchaHandler
}

func Setup(cfg *config.Config, db *database.DB, jwtMgr *jwt.Manager, h *Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Global middleware stack (ordered per plan)
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Logging())
	r.Use(middleware.NewIPFilter(db))
	r.Use(middleware.RateLimit(cfg.Security.RateLimitPerMinute))
	r.Use(middleware.CSRF(cfg.Security.CSRFEnabled))
	r.Use(middleware.Auth(jwtMgr))

	api := r.Group("/api/v1")

	// CSRF token endpoint
	api.GET("/csrf", func(c *gin.Context) {
		// Handled by CSRF middleware's GET handler
		c.JSON(200, gin.H{"message": "CSRF token set in cookie"})
	})

	// Health check
	api.GET("/health", func(c *gin.Context) {
		if err := db.HealthCheck(c.Request.Context()); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "healthy"})
	})

	idempotent := middleware.Idempotency(db, cfg.Security.IdempotencyTTL)

	// ========== Public routes ==========
	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
	}

	captcha := api.Group("/captcha")
	{
		captcha.GET("/generate", h.Captcha.Generate)
		captcha.POST("/verify", h.Captcha.Verify)
	}

	// Public item/review/QA browsing
	api.GET("/items", h.Item.List)
	api.GET("/items/:id", h.Item.GetByID)
	api.GET("/items/:id/reviews", h.Review.ListByItem)
	api.GET("/items/:id/questions", h.QA.ListQuestions)
	api.GET("/questions/:id/answers", h.QA.ListAnswers)
	api.GET("/images/:hash", h.Image.ServeByHash)

	// Shared analytics view (requires auth but no specific role)
	api.GET("/shared/:token", middleware.RequireAuth(), h.Dashboard.GetSharedView)
	api.GET("/shared/:token/data", middleware.RequireAuth(), h.Dashboard.GetSharedViewData)

	// ========== Authenticated routes ==========
	authed := api.Group("")
	authed.Use(middleware.RequireAuth())
	{
		authed.POST("/auth/logout", idempotent, h.Auth.Logout)
		authed.GET("/users/me", h.User.GetProfile)
		authed.PUT("/users/me", h.User.UpdateProfile)
		authed.PUT("/users/me/preferences", h.User.UpdatePreferences)

		// Reviews
		authed.POST("/items/:id/reviews", idempotent, h.Review.Create)
		authed.PUT("/reviews/:id", h.Review.Update)
		authed.DELETE("/reviews/:id", h.Review.Delete)

		// Images
		authed.POST("/images/upload", idempotent, h.Image.Upload)

		// Q&A
		authed.POST("/items/:id/questions", idempotent, h.QA.CreateQuestion)
		authed.PUT("/questions/:id", h.QA.UpdateQuestion)
		authed.DELETE("/questions/:id", h.QA.DeleteQuestion)
		authed.POST("/questions/:id/answers", idempotent, h.QA.CreateAnswer)
		authed.PUT("/answers/:id", h.QA.UpdateAnswer)
		authed.DELETE("/answers/:id", h.QA.DeleteAnswer)

		// Favorites
		authed.GET("/favorites", h.Favorite.List)
		authed.POST("/favorites", idempotent, h.Favorite.Add)
		authed.DELETE("/favorites/:item_id", h.Favorite.Remove)

		// Wishlists
		authed.GET("/wishlists", h.Wishlist.List)
		authed.POST("/wishlists", idempotent, h.Wishlist.Create)
		authed.PUT("/wishlists/:id", h.Wishlist.Update)
		authed.DELETE("/wishlists/:id", h.Wishlist.Delete)
		authed.POST("/wishlists/:id/items", idempotent, h.Wishlist.AddItem)
		authed.DELETE("/wishlists/:id/items/:item_id", h.Wishlist.RemoveItem)

		// Reports
		authed.POST("/reports", idempotent, h.Moderation.CreateReport)
		authed.GET("/reports/mine", h.Moderation.ListMyReports)
		authed.POST("/reports/:id/appeal", idempotent, h.Moderation.CreateAppeal)
		authed.PUT("/reports/:id/appeal", h.Moderation.ResubmitAppeal)

		// Analytics events
		authed.POST("/analytics/events", idempotent, h.Analytics.IngestEvents)
		authed.POST("/analytics/sessions", idempotent, h.Analytics.CreateSession)
		authed.PUT("/analytics/sessions/:id/heartbeat", h.Analytics.Heartbeat)

		// Experiment assignment & exposure
		authed.GET("/experiments/assignment/:exp_id", h.Experiment.GetAssignment)
		authed.POST("/experiments/:id/expose", idempotent, h.Experiment.RecordExposure)

		// Notifications
		authed.GET("/notifications", h.Notification.List)
		authed.GET("/notifications/:id", h.Notification.GetByID)
		authed.GET("/notifications/unread-count", h.Notification.UnreadCount)
		authed.PUT("/notifications/:id/read", h.Notification.MarkRead)
		authed.PUT("/notifications/read-all", h.Notification.MarkAllRead)

		// Frontend error capture
		authed.POST("/monitoring/frontend-errors", idempotent, h.Admin.CaptureError)
	}

	// ========== Moderator routes ==========
	mod := api.Group("/moderation")
	mod.Use(middleware.RequireAuth(), middleware.RequireRole("moderator", "admin"))
	{
		mod.GET("/queue", h.Moderation.ListQueue)
		mod.PUT("/reports/:id", h.Moderation.UpdateReport)
		mod.POST("/reports/:id/notes", idempotent, h.Moderation.AddNote)
		mod.GET("/reports/:id/notes", h.Moderation.ListNotes)
		mod.GET("/appeals", h.Moderation.ListAppeals)
		mod.PUT("/appeals/:id", h.Moderation.HandleAppeal)
		mod.GET("/quarantine", h.Moderation.ListQuarantined)
		mod.PUT("/quarantine/:id", h.Moderation.HandleQuarantine)
		mod.GET("/fraud", h.Moderation.ListFraudReviews)
		mod.PUT("/fraud/:review_id", h.Moderation.HandleFraud)
		mod.GET("/word-rules", h.Moderation.ListWordRules)
		mod.POST("/word-rules", idempotent, h.Moderation.CreateWordRule)
		mod.PUT("/word-rules/:id", h.Moderation.UpdateWordRule)
		mod.DELETE("/word-rules/:id", h.Moderation.DeleteWordRule)
	}

	// ========== Analyst routes ==========
	analyst := api.Group("")
	analyst.Use(middleware.RequireAuth(), middleware.RequireRole("product_analyst", "admin"))
	{
		analyst.GET("/analytics/dashboard", h.Dashboard.GetDashboard)
		analyst.GET("/analytics/keywords", h.Dashboard.GetKeywords)
		analyst.GET("/analytics/topics", h.Dashboard.GetTopics)
		analyst.GET("/analytics/cooccurrences", h.Dashboard.GetCooccurrences)
		analyst.GET("/analytics/sentiment", h.Dashboard.GetSentimentDistribution)
		analyst.GET("/analytics/aggregate-sessions", h.Analytics.ListAggregateSessions)
		analyst.GET("/analytics/sessions/:id", h.Analytics.GetSession)
		analyst.GET("/analytics/sessions/:id/timeline", h.Analytics.GetSessionTimeline)
		analyst.GET("/analytics/saved-views", h.Dashboard.ListSavedViews)
		analyst.POST("/analytics/saved-views", idempotent, h.Dashboard.CreateSavedView)
		analyst.PUT("/analytics/saved-views/:id", h.Dashboard.UpdateSavedView)
		analyst.DELETE("/analytics/saved-views/:id", h.Dashboard.DeleteSavedView)
		analyst.POST("/analytics/saved-views/:id/share", idempotent, h.Dashboard.CreateShareLink)
		analyst.DELETE("/analytics/saved-views/:id/share", h.Dashboard.RevokeShareLink)
		analyst.POST("/analytics/saved-views/clone", idempotent, h.Dashboard.CloneSavedView)

		analyst.GET("/experiments", h.Experiment.List)
		analyst.POST("/experiments", idempotent, h.Experiment.Create)
		analyst.GET("/experiments/:id", h.Experiment.GetByID)
		analyst.PUT("/experiments/:id", h.Experiment.Update)
		analyst.PUT("/experiments/:id/traffic", h.Experiment.UpdateTraffic)
		analyst.POST("/experiments/:id/start", idempotent, h.Experiment.Start)
		analyst.POST("/experiments/:id/pause", idempotent, h.Experiment.Pause)
		analyst.POST("/experiments/:id/complete", idempotent, h.Experiment.Complete)
		analyst.POST("/experiments/:id/rollback", idempotent, h.Experiment.Rollback)
		analyst.GET("/experiments/:id/results", h.Experiment.GetResults)

		analyst.GET("/scoring/weights", h.Analytics.GetScoringWeights)
		analyst.PUT("/scoring/weights", h.Analytics.UpdateScoringWeights)
		analyst.GET("/scoring/weights/history", h.Analytics.GetScoringWeightsHistory)
	}

	// ========== Admin routes ==========
	admin := api.Group("/admin")
	admin.Use(middleware.RequireAuth(), middleware.RequireRole("admin"))
	{
		admin.GET("/users", h.Admin.ListUsers)
		admin.PUT("/users/:id/role", h.Admin.UpdateUserRole)
		admin.PUT("/users/:id/status", h.Admin.UpdateUserStatus)
		admin.GET("/audit-logs", h.Admin.ListAuditLogs)
		admin.GET("/ip-rules", h.Admin.ListIPRules)
		admin.POST("/ip-rules", idempotent, h.Admin.CreateIPRule)
		admin.DELETE("/ip-rules/:id", h.Admin.DeleteIPRule)
		admin.POST("/backup/trigger", idempotent, h.Admin.TriggerBackup)
		admin.GET("/recovery-drills", h.Admin.ListRecoveryDrills)
		admin.POST("/recovery-drills/trigger", idempotent, h.Admin.TriggerRecoveryDrill)
		admin.POST("/analytics/rebuild", idempotent, h.Admin.RebuildAnalytics)
		admin.GET("/monitoring/performance", h.Admin.GetPerformanceMetrics)
		admin.GET("/monitoring/errors", h.Admin.GetErrorMetrics)
		admin.GET("/monitoring/health", h.Admin.GetSystemHealth)
	}

	return r
}
