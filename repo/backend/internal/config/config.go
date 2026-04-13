package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Security  SecurityConfig
	Storage   StorageConfig
	Backup    BackupConfig
	Analytics AnalyticsConfig
}

type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	TLSCert      string
	TLSKey       string
	RequireTLS   bool
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MigrationsPath  string
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type SecurityConfig struct {
	RateLimitPerMinute int
	CaptchaThreshold   int
	CaptchaWindow      time.Duration
	IdempotencyTTL     time.Duration
	CSRFEnabled        bool
}

type StorageConfig struct {
	ImagesDir     string
	QuarantineDir string
	MaxImageSize  int64
	AllowedTypes  []string
}

type BackupConfig struct {
	Dir           string
	RetentionDays int
}

type AnalyticsConfig struct {
	EventRateLimit           int
	DedupWindow              time.Duration
	DwellCapSeconds          int
	SessionInactivityTimeout time.Duration
	SequenceFraudThreshold   int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Addr:         envOrDefault("SERVER_ADDR", ":8080"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			TLSCert:      envOrDefault("TLS_CERT", ""),
			TLSKey:       envOrDefault("TLS_KEY", ""),
			RequireTLS:   envBoolOrDefault("REQUIRE_TLS", true),
		},
		Database: DatabaseConfig{
			DSN:             envOrDefault("DB_DSN", "root:rootpassword@tcp(localhost:3306)/local_insights?parseTime=true&charset=utf8mb4&loc=UTC&multiStatements=true"),
			MaxOpenConns:    envIntOrDefault("DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:    envIntOrDefault("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: 5 * time.Minute,
			MigrationsPath:  envOrDefault("MIGRATIONS_PATH", "./migrations"),
		},
		JWT: JWTConfig{
			Secret:     envOrDefault("JWT_SECRET", "default-dev-secret-must-change-in-production-64chars"),
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 168 * time.Hour,
		},
		Security: SecurityConfig{
			RateLimitPerMinute: envIntOrDefault("RATE_LIMIT_PER_MINUTE", 60),
			CaptchaThreshold:   5,
			CaptchaWindow:      15 * time.Minute,
			IdempotencyTTL:     10 * time.Minute,
			CSRFEnabled:        envBoolOrDefault("CSRF_ENABLED", true),
		},
		Storage: StorageConfig{
			ImagesDir:     envOrDefault("STORAGE_IMAGES_DIR", "/app/storage/images"),
			QuarantineDir: envOrDefault("STORAGE_QUARANTINE_DIR", "/app/storage/quarantine"),
			MaxImageSize:  5 * 1024 * 1024,
			AllowedTypes:  []string{"image/jpeg", "image/png", "image/webp"},
		},
		Backup: BackupConfig{
			Dir:           envOrDefault("BACKUP_DIR", "/app/storage/backups"),
			RetentionDays: 30,
		},
		Analytics: AnalyticsConfig{
			EventRateLimit:           300,
			DedupWindow:              2 * time.Second,
			DwellCapSeconds:          600,
			SessionInactivityTimeout: 30 * time.Minute,
			SequenceFraudThreshold:   3,
		},
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envBoolOrDefault(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
