package job

import (
	"log/slog"

	"github.com/robfig/cron/v3"
	"github.com/localinsights/portal/internal/config"
	"github.com/localinsights/portal/internal/pkg/database"
)

type Scheduler struct {
	cron *cron.Cron
	db   *database.DB
	cfg  *config.Config
}

func NewScheduler(db *database.DB, cfg *config.Config) *Scheduler {
	c := cron.New(cron.WithSeconds())
	return &Scheduler{cron: c, db: db, cfg: cfg}
}

func (s *Scheduler) RegisterAll() {
	// Rating refresh - every 30 seconds
	s.register("*/30 * * * * *", "rating_refresh", NewRatingRefreshJob(s.db))

	// Analytics ETL - every 5 minutes
	s.register("0 */5 * * * *", "analytics_etl", NewAnalyticsETLJob(s.db))

	// Idempotency cleanup - every 2 minutes
	s.register("0 */2 * * * *", "idempotency_cleanup", NewIdempotencyCleanupJob(s.db))

	// Share link expiry - every hour
	s.register("0 0 * * * *", "share_link_expiry", NewShareLinkExpiryJob(s.db))

	// Nightly backup - 2:00 AM
	s.register("0 0 2 * * *", "nightly_backup", NewBackupJob(s.db, s.cfg.Backup))

	// Weekly recovery drill - Sunday 3:00 AM
	s.register("0 0 3 * * 0", "weekly_recovery_drill", NewRecoveryDrillJob(s.db, s.cfg))

	// Fraud scan - every 10 minutes
	s.register("0 */10 * * * *", "fraud_scan", NewFraudScanJob(s.db, s.cfg.Analytics))

	// CSRF cleanup - every hour
	s.register("0 0 * * * *", "csrf_cleanup", NewCSRFCleanupJob(s.db))

	// Notification cleanup - daily at 4 AM
	s.register("0 0 4 * * *", "notification_cleanup", NewNotificationCleanupJob(s.db))

	// Monitoring metrics - every minute
	s.register("0 * * * * *", "monitoring_collect", NewMonitoringCollectJob(s.db))

	// NLP processing - every 5 minutes
	s.register("0 */5 * * * *", "nlp_processing", NewNLPProcessingJob(s.db))
}

func (s *Scheduler) register(spec, name string, job cron.Job) {
	_, err := s.cron.AddJob(spec, job)
	if err != nil {
		slog.Error("Failed to register job", "name", name, "error", err)
		return
	}
	slog.Info("Registered background job", "name", name, "schedule", spec)
}

func (s *Scheduler) Start() {
	slog.Info("Starting background job scheduler")
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	slog.Info("Stopping background job scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
}
