package unit_tests_test

import (
	"encoding/hex"
	"testing"

	"github.com/localinsights/portal/internal/pkg/crypto"
)

func TestGenerateToken(t *testing.T) {
	t.Run("returns hex-encoded string of correct length", func(t *testing.T) {
		tok, err := crypto.GenerateToken(16)
		if err != nil {
			t.Fatalf("GenerateToken returned error: %v", err)
		}
		// 16 bytes -> 32 hex characters.
		if len(tok) != 32 {
			t.Fatalf("expected 32 hex chars, got %d (%q)", len(tok), tok)
		}
		if _, err := hex.DecodeString(tok); err != nil {
			t.Fatalf("token is not valid hex: %v", err)
		}
	})

	t.Run("returns different values on successive calls", func(t *testing.T) {
		a, _ := crypto.GenerateToken(16)
		b, _ := crypto.GenerateToken(16)
		if a == b {
			t.Fatalf("two random tokens should differ, both = %q", a)
		}
	})

	t.Run("length=0 still succeeds and returns empty string", func(t *testing.T) {
		tok, err := crypto.GenerateToken(0)
		if err != nil {
			t.Fatalf("GenerateToken(0) returned error: %v", err)
		}
		if tok != "" {
			t.Fatalf("expected empty string, got %q", tok)
		}
	})
}

func TestGenerateShareToken(t *testing.T) {
	tok, err := crypto.GenerateShareToken()
	if err != nil {
		t.Fatalf("GenerateShareToken returned error: %v", err)
	}
	// 32 bytes -> 64 hex chars, per GenerateShareToken doc comment.
	if len(tok) != 64 {
		t.Fatalf("expected 64 hex chars, got %d (%q)", len(tok), tok)
	}
	if _, err := hex.DecodeString(tok); err != nil {
		t.Fatalf("share token is not valid hex: %v", err)
	}
}

func TestGenerateCSRFToken(t *testing.T) {
	tok, err := crypto.GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken returned error: %v", err)
	}
	if len(tok) != 64 {
		t.Fatalf("expected 64 hex chars, got %d (%q)", len(tok), tok)
	}

	// Two separate CSRF tokens must differ.
	tok2, _ := crypto.GenerateCSRFToken()
	if tok == tok2 {
		t.Fatalf("CSRF tokens should be random, both = %q", tok)
	}
}
