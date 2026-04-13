package unit_tests_test

import (
	"os"
	"testing"

	"github.com/localinsights/portal/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.Load()

	if cfg.Server.Addr != ":8080" {
		t.Errorf("expected default Server.Addr %q, got %q", ":8080", cfg.Server.Addr)
	}
	if cfg.JWT.Secret == "" {
		t.Fatal("expected non-empty default JWT secret")
	}
	if cfg.Database.MaxOpenConns <= 0 {
		t.Errorf("expected positive MaxOpenConns, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.JWT.AccessTTL <= 0 {
		t.Error("expected positive AccessTTL")
	}
	if cfg.JWT.RefreshTTL <= 0 {
		t.Error("expected positive RefreshTTL")
	}
	if cfg.Storage.MaxImageSize <= 0 {
		t.Error("expected positive MaxImageSize")
	}
}

func TestEnvOverride(t *testing.T) {
	customDSN := "user:pass@tcp(db.example.com:3306)/testdb"
	os.Setenv("DB_DSN", customDSN)
	t.Cleanup(func() {
		os.Unsetenv("DB_DSN")
	})

	cfg := config.Load()
	if cfg.Database.DSN != customDSN {
		t.Errorf("expected DB_DSN %q, got %q", customDSN, cfg.Database.DSN)
	}
}

func TestServerAddr(t *testing.T) {
	customAddr := ":9090"
	os.Setenv("SERVER_ADDR", customAddr)
	t.Cleanup(func() {
		os.Unsetenv("SERVER_ADDR")
	})

	cfg := config.Load()
	if cfg.Server.Addr != customAddr {
		t.Errorf("expected SERVER_ADDR %q, got %q", customAddr, cfg.Server.Addr)
	}
}
