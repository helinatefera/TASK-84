package unit_tests_test

import (
	"testing"
	"time"

	"github.com/localinsights/portal/internal/pkg/locale"
)

func TestFormatTimestampUTC(t *testing.T) {
	// 2024-06-15 14:30:00 UTC
	ts := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	result := locale.FormatTimestamp(ts, "en", "UTC")
	expected := "Jun 15, 2024 2:30 PM"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatTimestampTimezone(t *testing.T) {
	// 2024-06-15 14:30:00 UTC should be 10:30 AM in America/New_York (EDT, UTC-4)
	ts := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	result := locale.FormatTimestamp(ts, "en", "America/New_York")
	expected := "Jun 15, 2024 10:30 AM"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatTimestampSpanish(t *testing.T) {
	ts := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	enResult := locale.FormatTimestamp(ts, "en", "UTC")
	esResult := locale.FormatTimestamp(ts, "es", "UTC")

	if enResult == esResult {
		t.Errorf("expected different formats for en (%q) and es (%q)", enResult, esResult)
	}
}

func TestFormatISO(t *testing.T) {
	ts := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	result := locale.FormatISO(ts, "UTC")
	expected := "2024-06-15T14:30:00Z"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatDuration(t *testing.T) {
	t.Run("seconds", func(t *testing.T) {
		result := locale.FormatDuration(45)
		if result != "45s" {
			t.Errorf("expected %q, got %q", "45s", result)
		}
	})

	t.Run("minutes and seconds", func(t *testing.T) {
		result := locale.FormatDuration(125)
		if result != "2m 5s" {
			t.Errorf("expected %q, got %q", "2m 5s", result)
		}
	})

	t.Run("hours and minutes", func(t *testing.T) {
		result := locale.FormatDuration(3725)
		if result != "1h 2m" {
			t.Errorf("expected %q, got %q", "1h 2m", result)
		}
	})
}
