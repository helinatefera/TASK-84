package unit_tests_test

import (
	"testing"
	"time"

	"github.com/localinsights/portal/internal/config"
	"github.com/localinsights/portal/internal/pkg/jwt"
)

func newJWTManager() *jwt.Manager {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key-at-least-32-chars!!",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}
	return jwt.NewManager(cfg)
}

func TestGenerateAccessToken(t *testing.T) {
	m := newJWTManager()
	token, err := m.GenerateAccessToken(1, "uuid-123", "testuser", "admin")
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateAccessToken returned empty token")
	}
}

func TestValidateAccessToken(t *testing.T) {
	m := newJWTManager()

	userID := uint64(42)
	userUUID := "uuid-456"
	username := "alice"
	role := "editor"

	token, err := m.GenerateAccessToken(userID, userUUID, username, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := m.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.Username != username {
		t.Errorf("expected Username %q, got %q", username, claims.Username)
	}
	if claims.Role != role {
		t.Errorf("expected Role %q, got %q", role, claims.Role)
	}
	if claims.UserUUID != userUUID {
		t.Errorf("expected UserUUID %q, got %q", userUUID, claims.UserUUID)
	}
}

func TestExpiredToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key-at-least-32-chars!!",
		AccessTTL:  1 * time.Millisecond,
		RefreshTTL: time.Hour,
	}
	m := jwt.NewManager(cfg)

	token, err := m.GenerateAccessToken(1, "uuid-789", "expireduser", "viewer")
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	// Wait for the token to expire
	time.Sleep(50 * time.Millisecond)

	_, err = m.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("ValidateAccessToken should return error for expired token")
	}
}

func TestInvalidToken(t *testing.T) {
	m := newJWTManager()
	_, err := m.ValidateAccessToken("this.is.garbage")
	if err == nil {
		t.Fatal("ValidateAccessToken should return error for invalid token string")
	}
}

func TestRefreshToken(t *testing.T) {
	m := newJWTManager()
	userUUID := "refresh-uuid-001"

	token, expiresAt, err := m.GenerateRefreshToken(99, userUUID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateRefreshToken returned empty token")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("refresh token expiry should be in the future")
	}

	subject, err := m.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken returned error: %v", err)
	}
	if subject != userUUID {
		t.Errorf("expected subject %q, got %q", userUUID, subject)
	}
}
