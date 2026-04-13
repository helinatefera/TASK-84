package unit_tests_test

import (
	"testing"

	"github.com/localinsights/portal/internal/pkg/audit"
)

func TestMaskPassword(t *testing.T) {
	details := map[string]any{
		"password": "supersecret123",
	}
	masked := audit.MaskDetails(details)
	if masked["password"] != "***REDACTED***" {
		t.Errorf("expected password to be redacted, got %v", masked["password"])
	}
}

func TestMaskToken(t *testing.T) {
	details := map[string]any{
		"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
	}
	masked := audit.MaskDetails(details)
	if masked["token"] != "***REDACTED***" {
		t.Errorf("expected token to be redacted, got %v", masked["token"])
	}
}

func TestMaskEmail(t *testing.T) {
	details := map[string]any{
		"email": "john@example.com",
	}
	masked := audit.MaskDetails(details)
	expected := "j***@example.com"
	if masked["email"] != expected {
		t.Errorf("expected email %q, got %v", expected, masked["email"])
	}
}

func TestPreserveNormalFields(t *testing.T) {
	details := map[string]any{
		"action":   "login",
		"username": "alice",
		"count":    42,
	}
	masked := audit.MaskDetails(details)

	if masked["action"] != "login" {
		t.Errorf("expected action %q, got %v", "login", masked["action"])
	}
	if masked["username"] != "alice" {
		t.Errorf("expected username %q, got %v", "alice", masked["username"])
	}
	if masked["count"] != 42 {
		t.Errorf("expected count %d, got %v", 42, masked["count"])
	}
}

func TestMaskNilDetails(t *testing.T) {
	masked := audit.MaskDetails(nil)
	if masked != nil {
		t.Errorf("expected nil for nil input, got %v", masked)
	}
}
