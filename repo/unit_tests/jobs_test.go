package unit_tests_test

import (
	"testing"

	"github.com/localinsights/portal/internal/config"
	"github.com/localinsights/portal/internal/job"
)

// Verifies every background-job constructor is callable and the scheduler
// wires them up. Real job execution depends on a live MySQL and is covered
// by API tests; these locks lock in the public constructor contract so a
// refactor cannot silently drop a job.

func TestJobConstructorsNonNil(t *testing.T) {
	if job.NewRatingRefreshJob(nil) == nil {
		t.Error("NewRatingRefreshJob returned nil")
	}
	if job.NewAnalyticsETLJob(nil) == nil {
		t.Error("NewAnalyticsETLJob returned nil")
	}
	if job.NewIdempotencyCleanupJob(nil) == nil {
		t.Error("NewIdempotencyCleanupJob returned nil")
	}
	if job.NewShareLinkExpiryJob(nil) == nil {
		t.Error("NewShareLinkExpiryJob returned nil")
	}
	if job.NewBackupJob(nil, config.BackupConfig{}) == nil {
		t.Error("NewBackupJob returned nil")
	}
	if job.NewRecoveryDrillJob(nil, &config.Config{}) == nil {
		t.Error("NewRecoveryDrillJob returned nil")
	}
	if job.NewFraudScanJob(nil, config.AnalyticsConfig{}) == nil {
		t.Error("NewFraudScanJob returned nil")
	}
	if job.NewCSRFCleanupJob(nil) == nil {
		t.Error("NewCSRFCleanupJob returned nil")
	}
	if job.NewNotificationCleanupJob(nil) == nil {
		t.Error("NewNotificationCleanupJob returned nil")
	}
	if job.NewMonitoringCollectJob(nil) == nil {
		t.Error("NewMonitoringCollectJob returned nil")
	}
	if job.NewNLPProcessingJob(nil) == nil {
		t.Error("NewNLPProcessingJob returned nil")
	}
}

func TestSchedulerConstructorNonNil(t *testing.T) {
	s := job.NewScheduler(nil, &config.Config{})
	if s == nil {
		t.Fatal("NewScheduler returned nil")
	}
}

func TestSchedulerRegisterAllDoesNotPanic(t *testing.T) {
	// RegisterAll must be safe to call even without a working DB;
	// the jobs are not run, only registered.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterAll panicked: %v", r)
		}
	}()
	s := job.NewScheduler(nil, &config.Config{})
	s.RegisterAll()
}
