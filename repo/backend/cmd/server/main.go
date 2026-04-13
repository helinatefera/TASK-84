package main

import (
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/localinsights/portal/internal/config"
	"github.com/localinsights/portal/internal/handler"
	"github.com/localinsights/portal/internal/job"
	"github.com/localinsights/portal/internal/pkg/captcha"
	"github.com/localinsights/portal/internal/pkg/database"
	"github.com/localinsights/portal/internal/pkg/jwt"
	"github.com/localinsights/portal/internal/repository"
	"github.com/localinsights/portal/internal/router"
	"github.com/localinsights/portal/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Load()
	slog.Info("Configuration loaded", "addr", cfg.Server.Addr)

	// Database connection with retry
	var db *database.DB
	var err error
	for i := range 30 {
		db, err = database.NewMySQL(cfg.Database)
		if err == nil {
			break
		}
		slog.Warn("Database not ready, retrying...", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}
	defer db.Close()
	slog.Info("Database connected")

	// Run migrations
	runMigrations(db, cfg.Database.MigrationsPath)

	// JWT Manager
	jwtMgr := jwt.NewManager(cfg.JWT)

	// CAPTCHA Store
	captchaStore := captcha.NewStore()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	prefsRepo := repository.NewUserPreferencesRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	itemRepo := repository.NewItemRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	imageRepo := repository.NewImageRepository(db)
	reviewImageRepo := repository.NewReviewImageRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	answerRepo := repository.NewAnswerRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	wishlistRepo := repository.NewWishlistRepository(db)
	reportRepo := repository.NewReportRepository(db)
	appealRepo := repository.NewAppealRepository(db)
	modNoteRepo := repository.NewModerationNoteRepository(db)
	wordRuleRepo := repository.NewSensitiveWordRuleRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	auditRepo := repository.NewAuditLogRepository(db)
	ipRuleRepo := repository.NewIPRuleRepository(db)

	// Services
	authSvc := service.NewAuthService(
		userRepo, prefsRepo, loginAttemptRepo, refreshTokenRepo, jwtMgr,
		cfg.Security.CaptchaThreshold, cfg.Security.CaptchaWindow, captchaStore,
	)
	contentFilter := service.NewContentFilter(wordRuleRepo)

	// Background Jobs
	scheduler := job.NewScheduler(db, cfg)
	scheduler.RegisterAll()
	scheduler.Start()
	defer scheduler.Stop()

	// Handlers
	handlers := &router.Handlers{
		Auth:         handler.NewAuthHandler(authSvc),
		User:         handler.NewUserHandler(userRepo, prefsRepo),
		Item:         handler.NewItemHandler(itemRepo, db),
		Review:       handler.NewReviewHandler(reviewRepo, imageRepo, reviewImageRepo, contentFilter, db, cfg.Storage),
		Image:        handler.NewImageHandler(imageRepo, cfg.Storage),
		QA:           handler.NewQAHandler(questionRepo, answerRepo, itemRepo, contentFilter, db),
		Favorite:     handler.NewFavoriteHandler(favoriteRepo),
		Wishlist:     handler.NewWishlistHandler(wishlistRepo),
		Moderation:   handler.NewModerationHandler(reportRepo, appealRepo, modNoteRepo, wordRuleRepo, imageRepo, reviewRepo, db),
		Analytics:    handler.NewAnalyticsHandler(db),
		Dashboard:    handler.NewDashboardHandler(db),
		Experiment:   handler.NewExperimentHandler(db),
		Notification: handler.NewNotificationHandler(notifRepo),
		Admin:        handler.NewAdminHandler(userRepo, auditRepo, ipRuleRepo, db, cfg),
		Captcha:      handler.NewCaptchaHandler(captchaStore),
	}

	// Router
	r := router.Setup(cfg, db, jwtMgr, handlers)

	// Server
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	if cfg.Server.TLSCert != "" && cfg.Server.TLSKey != "" {
		srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		go func() {
			slog.Info("Starting HTTPS server", "addr", cfg.Server.Addr)
			if err := srv.ListenAndServeTLS(cfg.Server.TLSCert, cfg.Server.TLSKey); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server failed: %v", err)
			}
		}()
	} else if cfg.Server.RequireTLS {
		log.Fatalf("REQUIRE_TLS is enabled but TLS_CERT and TLS_KEY are not configured. " +
			"Set REQUIRE_TLS=false for local development or provide TLS material.")
	} else {
		slog.Warn("TLS not configured — starting plaintext HTTP (development mode only)")
		go func() {
			slog.Info("Starting HTTP server", "addr", cfg.Server.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server failed: %v", err)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}
	slog.Info("Server exited")
}

func runMigrations(db *database.DB, migrationsPath string) {
	driver, err := mysql.WithInstance(db.DB.DB, &mysql.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "mysql", driver)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}
	slog.Info("Database migrations applied successfully")
}
