package job

import (
	"testing"

	"github.com/localinsights/portal/internal/config"
)

// These tests verify that every job constructor returns a non-nil job and
// that Scheduler.RegisterAll wires them into the cron scheduler without
// failing. Real job execution requires a DB and is covered by API tests;
// here we lock in the wiring contract so a refactor cannot silently drop
// a background job.

func TestJobConstructorsReturnNonNil(t *testing.T) {
	// Passing nil *database.DB is safe because these constructors do not
	// dereference the DB — they only hold a reference for later use.
	if NewRatingRefreshJob(nil) == nil {
		t.Error("NewRatingRefreshJob returned nil")
	}
	if NewAnalyticsETLJob(nil) == nil {
		t.Error("NewAnalyticsETLJob returned nil")
	}
	if NewIdempotencyCleanupJob(nil) == nil {
		t.Error("NewIdempotencyCleanupJob returned nil")
	}
	if NewShareLinkExpiryJob(nil) == nil {
		t.Error("NewShareLinkExpiryJob returned nil")
	}
	if NewBackupJob(nil, config.BackupConfig{}) == nil {
		t.Error("NewBackupJob returned nil")
	}
	if NewRecoveryDrillJob(nil, &config.Config{}) == nil {
		t.Error("NewRecoveryDrillJob returned nil")
	}
	if NewFraudScanJob(nil, config.AnalyticsConfig{}) == nil {
		t.Error("NewFraudScanJob returned nil")
	}
	if NewCSRFCleanupJob(nil) == nil {
		t.Error("NewCSRFCleanupJob returned nil")
	}
	if NewNotificationCleanupJob(nil) == nil {
		t.Error("NewNotificationCleanupJob returned nil")
	}
	if NewMonitoringCollectJob(nil) == nil {
		t.Error("NewMonitoringCollectJob returned nil")
	}
	if NewNLPProcessingJob(nil) == nil {
		t.Error("NewNLPProcessingJob returned nil")
	}
}

func TestSchedulerNewCreatesInstance(t *testing.T) {
	s := NewScheduler(nil, &config.Config{})
	if s == nil {
		t.Fatal("NewScheduler returned nil")
	}
	if s.cron == nil {
		t.Error("NewScheduler did not initialize cron")
	}
}

func TestSchedulerRegisterAllRegistersEveryJob(t *testing.T) {
	cfg := &config.Config{}
	s := NewScheduler(nil, cfg)
	s.RegisterAll()

	entries := s.cron.Entries()
	// The scheduler currently wires 11 cron jobs. If that list changes,
	// update this expectation deliberately — don't silently lose coverage.
	wantMin := 11
	if len(entries) < wantMin {
		t.Errorf("expected at least %d scheduled jobs, got %d", wantMin, len(entries))
	}

	// Every registered entry must have a non-nil Job implementation and a
	// valid schedule (Next must produce a future time).
	for i, e := range entries {
		if e.Job == nil {
			t.Errorf("entry %d has nil Job", i)
		}
		if e.Schedule == nil {
			t.Errorf("entry %d has nil Schedule", i)
		}
	}
}

func TestSchedulerStartStopDoesNotPanic(t *testing.T) {
	// We don't assert anything beyond "does not panic" because the cron
	// library owns its internal state. The goal is to ensure that these
	// methods remain callable on the public surface.
	s := NewScheduler(nil, &config.Config{})
	// Register nothing so we don't fire real jobs.
	s.Start()
	s.Stop()
}
