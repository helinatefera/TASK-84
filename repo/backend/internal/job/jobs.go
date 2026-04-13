package job

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/localinsights/portal/internal/config"
	"github.com/localinsights/portal/internal/pkg/database"
)

// AnalyticsETLJob aggregates raw behavior_events into analytics_aggregates
type AnalyticsETLJob struct{ db *database.DB }

func NewAnalyticsETLJob(db *database.DB) *AnalyticsETLJob { return &AnalyticsETLJob{db: db} }
func (j *AnalyticsETLJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	_, err := j.db.ExecContext(ctx, `
		INSERT INTO analytics_aggregates (item_id, period_start, impressions, clicks, avg_dwell_secs, favorites, shares, comments, computed_at)
		SELECT
			item_id,
			DATE(server_ts) as period_start,
			SUM(CASE WHEN event_type = 'impression' THEN 1 ELSE 0 END),
			SUM(CASE WHEN event_type = 'click' THEN 1 ELSE 0 END),
			COALESCE(AVG(CASE WHEN event_type = 'dwell' THEN dwell_seconds END), 0),
			SUM(CASE WHEN event_type = 'favorite' THEN 1 ELSE 0 END),
			SUM(CASE WHEN event_type = 'share' THEN 1 ELSE 0 END),
			SUM(CASE WHEN event_type = 'comment' THEN 1 ELSE 0 END),
			NOW(3)
		FROM behavior_events
		WHERE item_id IS NOT NULL
		GROUP BY item_id, DATE(server_ts)
		ON DUPLICATE KEY UPDATE
			impressions = VALUES(impressions),
			clicks = VALUES(clicks),
			avg_dwell_secs = VALUES(avg_dwell_secs),
			favorites = VALUES(favorites),
			shares = VALUES(shares),
			comments = VALUES(comments),
			computed_at = NOW(3)
	`)
	if err != nil {
		slog.Error("analytics_etl: failed", "error", err)
	}

	// Generate session sequence fingerprints for fraud detection
	_, _ = j.db.ExecContext(ctx, `
		INSERT IGNORE INTO session_sequence_fingerprints (session_id, user_id, sequence_hash, event_count, created_at)
		SELECT
			s.id,
			s.user_id,
			SHA2(GROUP_CONCAT(e.event_type ORDER BY e.server_ts SEPARATOR '->'), 256),
			COUNT(*),
			NOW(3)
		FROM analytics_sessions s
		JOIN behavior_events e ON e.session_id = s.id
		WHERE s.ended_at IS NOT NULL
		AND s.id NOT IN (SELECT session_id FROM session_sequence_fingerprints)
		GROUP BY s.id, s.user_id
	`)
}

// IdempotencyCleanupJob purges expired idempotency keys
type IdempotencyCleanupJob struct{ db *database.DB }

func NewIdempotencyCleanupJob(db *database.DB) *IdempotencyCleanupJob {
	return &IdempotencyCleanupJob{db: db}
}
func (j *IdempotencyCleanupJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := j.db.ExecContext(ctx, `DELETE FROM idempotency_keys WHERE expires_at < NOW() LIMIT 1000`)
	if err != nil {
		slog.Error("idempotency_cleanup: failed", "error", err)
		return
	}
	if n, _ := result.RowsAffected(); n > 0 {
		slog.Info("idempotency_cleanup: purged keys", "count", n)
	}
}

// ShareLinkExpiryJob deletes expired share links
type ShareLinkExpiryJob struct{ db *database.DB }

func NewShareLinkExpiryJob(db *database.DB) *ShareLinkExpiryJob {
	return &ShareLinkExpiryJob{db: db}
}
func (j *ShareLinkExpiryJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := j.db.ExecContext(ctx, `DELETE FROM share_links WHERE expires_at < NOW() OR is_revoked = 1`)
	if err != nil {
		slog.Error("share_link_expiry: failed", "error", err)
	}
}

// BackupJob runs mysqldump for nightly backups
type BackupJob struct {
	db  *database.DB
	cfg config.BackupConfig
}

func NewBackupJob(db *database.DB, cfg config.BackupConfig) *BackupJob {
	return &BackupJob{db: db, cfg: cfg}
}
func (j *BackupJob) Run() {
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(j.cfg.Dir, fmt.Sprintf("backup_%s.sql", timestamp))

	cmd := exec.Command("mysqldump",
		"--single-transaction", "--routines", "--triggers",
		"-h", dbHost(), "-u", dbUser(), fmt.Sprintf("-p%s", dbPassword()), dbName())

	outFile, err := os.Create(filename)
	if err != nil {
		slog.Error("nightly_backup: failed to create file", "error", err)
		return
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		slog.Error("nightly_backup: mysqldump failed", "error", err)
		os.Remove(filename)
		return
	}
	slog.Info("nightly_backup: completed", "file", filename)

	// Retention: remove backups older than configured days
	entries, err := os.ReadDir(j.cfg.Dir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -j.cfg.RetentionDays)
	for _, e := range entries {
		if info, err := e.Info(); err == nil && info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(j.cfg.Dir, e.Name()))
			slog.Info("nightly_backup: removed old backup", "file", e.Name())
		}
	}
}

// RecoveryDrillJob restores latest backup to temp DB and validates
type RecoveryDrillJob struct {
	db  *database.DB
	cfg *config.Config
}

func NewRecoveryDrillJob(db *database.DB, cfg *config.Config) *RecoveryDrillJob {
	return &RecoveryDrillJob{db: db, cfg: cfg}
}
func (j *RecoveryDrillJob) Run() {
	slog.Info("recovery_drill: starting")

	// Find latest backup
	entries, err := os.ReadDir(j.cfg.Backup.Dir)
	if err != nil || len(entries) == 0 {
		slog.Error("recovery_drill: no backups found", "error", err)
		return
	}

	var latestBackup string
	var latestTime time.Time
	for _, e := range entries {
		if info, err := e.Info(); err == nil && info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestBackup = filepath.Join(j.cfg.Backup.Dir, e.Name())
		}
	}

	if latestBackup == "" {
		slog.Error("recovery_drill: could not find backup file")
		return
	}

	drillDB := fmt.Sprintf("local_insights_drill_%s", time.Now().Format("20060102"))

	// Create temp DB
	createCmd := exec.Command("mysql", "-h", dbHost(), "-u", dbUser(), fmt.Sprintf("-p%s", dbPassword()),
		"-e", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", drillDB))
	if err := createCmd.Run(); err != nil {
		slog.Error("recovery_drill: failed to create temp db", "error", err)
		return
	}

	// Restore backup
	inFile, err := os.Open(latestBackup)
	if err != nil {
		slog.Error("recovery_drill: failed to open backup", "error", err)
		return
	}
	defer inFile.Close()

	restoreCmd := exec.Command("mysql", "-h", dbHost(), "-u", dbUser(), fmt.Sprintf("-p%s", dbPassword()), drillDB)
	restoreCmd.Stdin = inFile
	if err := restoreCmd.Run(); err != nil {
		slog.Error("recovery_drill: restore failed", "error", err)
	} else {
		slog.Info("recovery_drill: restore successful", "backup", latestBackup, "db", drillDB)
	}

	// Drop temp DB
	dropCmd := exec.Command("mysql", "-h", dbHost(), "-u", dbUser(), fmt.Sprintf("-p%s", dbPassword()),
		"-e", fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", drillDB))
	dropCmd.Run()

	slog.Info("recovery_drill: completed")
}

// FraudScanJob checks event rate limits and sequence patterns
type FraudScanJob struct {
	db  *database.DB
	cfg config.AnalyticsConfig
}

func NewFraudScanJob(db *database.DB, cfg config.AnalyticsConfig) *FraudScanJob {
	return &FraudScanJob{db: db, cfg: cfg}
}
func (j *FraudScanJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Rate-based fraud: flag users exceeding event rate limit
	suspectedUserQuery := `
		SELECT user_id FROM user_event_counts
		WHERE hour_bucket >= DATE_FORMAT(NOW() - INTERVAL 1 HOUR, '%%Y-%%m-%%d %%H:00:00')
		AND event_count > ?`

	_, err := j.db.ExecContext(ctx, `
		UPDATE reviews SET fraud_status = 'suspected_fraud'
		WHERE user_id IN (`+suspectedUserQuery+`)
		AND fraud_status = 'normal'
	`, j.cfg.EventRateLimit)
	if err != nil {
		slog.Error("fraud_scan: rate-based review check failed", "error", err)
	}

	// Also flag the user accounts
	_, err = j.db.ExecContext(ctx, `
		UPDATE users SET fraud_status = 'suspected'
		WHERE id IN (`+suspectedUserQuery+`)
		AND fraud_status = 'clean'
	`, j.cfg.EventRateLimit)
	if err != nil {
		slog.Error("fraud_scan: rate-based account check failed", "error", err)
	}

	// Sequence-based fraud: flag users with repeated session patterns
	seqUserQuery := `
		SELECT user_id FROM session_sequence_fingerprints
		GROUP BY user_id, sequence_hash
		HAVING COUNT(*) >= ?`

	_, err = j.db.ExecContext(ctx, `
		UPDATE reviews SET fraud_status = 'suspected_fraud'
		WHERE user_id IN (`+seqUserQuery+`)
		AND fraud_status = 'normal'
	`, j.cfg.SequenceFraudThreshold)
	if err != nil {
		slog.Error("fraud_scan: sequence-based review check failed", "error", err)
	}

	// Also flag the user accounts
	_, err = j.db.ExecContext(ctx, `
		UPDATE users SET fraud_status = 'suspected'
		WHERE id IN (`+seqUserQuery+`)
		AND fraud_status = 'clean'
	`, j.cfg.SequenceFraudThreshold)
	if err != nil {
		slog.Error("fraud_scan: sequence-based account check failed", "error", err)
	}

	// Finalize expired sessions
	_, err = j.db.ExecContext(ctx, `
		UPDATE analytics_sessions SET ended_at = last_active_at
		WHERE ended_at IS NULL AND last_active_at < NOW() - INTERVAL 30 MINUTE
	`)
	if err != nil {
		slog.Error("fraud_scan: session finalization failed", "error", err)
	}
}

// CSRFCleanupJob purges expired CSRF tokens
type CSRFCleanupJob struct{ db *database.DB }

func NewCSRFCleanupJob(db *database.DB) *CSRFCleanupJob { return &CSRFCleanupJob{db: db} }
func (j *CSRFCleanupJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, _ = j.db.ExecContext(ctx, `DELETE FROM csrf_tokens WHERE expires_at < NOW()`)
}

// NotificationCleanupJob purges old notifications
type NotificationCleanupJob struct{ db *database.DB }

func NewNotificationCleanupJob(db *database.DB) *NotificationCleanupJob {
	return &NotificationCleanupJob{db: db}
}
func (j *NotificationCleanupJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := j.db.ExecContext(ctx, `DELETE FROM notifications WHERE created_at < NOW() - INTERVAL 90 DAY LIMIT 5000`)
	if err != nil {
		slog.Error("notification_cleanup: failed", "error", err)
		return
	}
	if n, _ := result.RowsAffected(); n > 0 {
		slog.Info("notification_cleanup: purged notifications", "count", n)
	}
}

// MonitoringCollectJob records performance metrics
type MonitoringCollectJob struct{ db *database.DB }

func NewMonitoringCollectJob(db *database.DB) *MonitoringCollectJob {
	return &MonitoringCollectJob{db: db}
}
func (j *MonitoringCollectJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats := j.db.Stats()
	_, _ = j.db.ExecContext(ctx, `
		INSERT INTO monitoring_metrics (metric_name, metric_value, recorded_at) VALUES
		('db_open_connections', ?, NOW(3)),
		('db_in_use', ?, NOW(3)),
		('db_idle', ?, NOW(3))
	`, stats.OpenConnections, stats.InUse, stats.Idle)
}

// NLPProcessingJob runs sentiment analysis and keyword extraction
type NLPProcessingJob struct{ db *database.DB }

func NewNLPProcessingJob(db *database.DB) *NLPProcessingJob {
	return &NLPProcessingJob{db: db}
}
func (j *NLPProcessingJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	// Find reviews not yet processed
	rows, err := j.db.QueryxContext(ctx, `
		SELECT r.id, r.body FROM reviews r
		LEFT JOIN review_sentiment s ON s.review_id = r.id
		WHERE s.id IS NULL AND r.body IS NOT NULL AND r.body != ''
		LIMIT 100
	`)
	if err != nil {
		slog.Error("nlp_processing: failed to find reviews", "error", err)
		return
	}
	defer rows.Close()

	processed := 0
	for rows.Next() {
		var id uint64
		var body string
		if err := rows.Scan(&id, &body); err != nil {
			continue
		}

		// Simple rule-based sentiment (placeholder for full NLP)
		sentiment, confidence := analyzeSentiment(body)

		_, err := j.db.ExecContext(ctx, `
			INSERT INTO review_sentiment (review_id, sentiment_label, confidence, processed_at)
			VALUES (?, ?, ?, NOW(3))
			ON DUPLICATE KEY UPDATE sentiment_label = VALUES(sentiment_label), confidence = VALUES(confidence), processed_at = NOW(3)
		`, id, sentiment, confidence)
		if err != nil {
			slog.Error("nlp_processing: failed to insert sentiment", "review_id", id, "error", err)
		}
		processed++
	}

	if processed > 0 {
		slog.Info("nlp_processing: processed reviews", "count", processed)
	}
}

func analyzeSentiment(text string) (string, float64) {
	// Simple keyword-based sentiment for offline operation
	positiveWords := []string{"great", "good", "excellent", "amazing", "love", "best", "wonderful", "fantastic", "awesome", "perfect"}
	negativeWords := []string{"bad", "terrible", "awful", "worst", "hate", "horrible", "poor", "disappointing", "useless", "broken"}

	posCount, negCount := 0, 0
	lower := text
	for _, w := range positiveWords {
		if contains(lower, w) {
			posCount++
		}
	}
	for _, w := range negativeWords {
		if contains(lower, w) {
			negCount++
		}
	}

	total := posCount + negCount
	if total == 0 {
		return "neutral", 0.5
	}
	if posCount > negCount {
		return "positive", float64(posCount) / float64(total)
	}
	if negCount > posCount {
		return "negative", float64(negCount) / float64(total)
	}
	return "neutral", 0.5
}

func contains(text, word string) bool {
	// Simple substring match
	for i := 0; i <= len(text)-len(word); i++ {
		match := true
		for j := 0; j < len(word); j++ {
			c := text[i+j]
			w := word[j]
			// Case-insensitive
			if c >= 'A' && c <= 'Z' {
				c += 32
			}
			if c != w {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func dbHost() string     { return envOrDefault("DB_HOST", "mysql") }
func dbUser() string     { return envOrDefault("DB_USER", "appuser") }
func dbPassword() string { return envOrDefault("DB_PASSWORD", "apppassword") }
func dbName() string     { return envOrDefault("DB_NAME", "local_insights") }
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
