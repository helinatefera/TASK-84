package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/errs"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/captcha"
	"github.com/localinsights/portal/internal/pkg/hash"
	"github.com/localinsights/portal/internal/pkg/jwt"
	"github.com/localinsights/portal/internal/repository"
)

type AuthService struct {
	userRepo         repository.UserRepository
	prefsRepo        repository.UserPreferencesRepository
	loginAttemptRepo repository.LoginAttemptRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtManager       *jwt.Manager
	captchaThreshold int
	captchaWindow    time.Duration
	captchaStore     *captcha.Store
}

func NewAuthService(
	userRepo repository.UserRepository,
	prefsRepo repository.UserPreferencesRepository,
	loginAttemptRepo repository.LoginAttemptRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtManager *jwt.Manager,
	captchaThreshold int,
	captchaWindow time.Duration,
	captchaStore *captcha.Store,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		prefsRepo:        prefsRepo,
		loginAttemptRepo: loginAttemptRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		captchaThreshold: captchaThreshold,
		captchaWindow:    captchaWindow,
		captchaStore:     captchaStore,
	}
}

func (s *AuthService) Register(ctx context.Context, req *request.RegisterRequest) (*model.User, error) {
	existing, _ := s.userRepo.GetByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errs.WithMessage(errs.ErrConflict, "Username already taken")
	}

	existing, _ = s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errs.WithMessage(errs.ErrConflict, "Email already registered")
	}

	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		UUID:         uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         model.RoleRegularUser,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	prefs := &model.UserPreferences{
		UserID:   user.ID,
		Locale:   "en",
		Timezone: "UTC",
	}
	_ = s.prefsRepo.Upsert(ctx, prefs)

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *request.LoginRequest, ip string) (string, string, *model.User, error) {
	failedCount, _ := s.loginAttemptRepo.CountRecentFailed(ctx, req.Username, ip, s.captchaWindow)
	if failedCount >= s.captchaThreshold {
		if req.CaptchaID == "" || req.CaptchaToken == "" {
			return "", "", nil, errs.ErrCaptchaRequired
		}
		if !s.captchaStore.Verify(req.CaptchaID, req.CaptchaToken) {
			return "", "", nil, errs.ErrCaptchaInvalid
		}
	}

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil || user == nil {
		s.recordLoginAttempt(ctx, req.Username, ip, false)
		return "", "", nil, errs.WithMessage(errs.ErrUnauthorized, "Invalid credentials")
	}

	if !user.IsActive {
		s.recordLoginAttempt(ctx, req.Username, ip, false)
		return "", "", nil, errs.WithMessage(errs.ErrForbidden, "Account is deactivated")
	}

	valid, err := hash.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		s.recordLoginAttempt(ctx, req.Username, ip, false)
		return "", "", nil, errs.WithMessage(errs.ErrUnauthorized, "Invalid credentials")
	}

	s.recordLoginAttempt(ctx, req.Username, ip, true)

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.UUID, user.Username, string(user.Role))
	if err != nil {
		return "", "", nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, expiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID, user.UUID)
	if err != nil {
		return "", "", nil, fmt.Errorf("generate refresh token: %w", err)
	}

	tokenHash := hashToken(refreshToken)
	if err := s.refreshTokenRepo.Create(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return "", "", nil, fmt.Errorf("store refresh token: %w", err)
	}

	return accessToken, refreshToken, user, nil
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, *model.User, error) {
	userUUID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", nil, errs.WithMessage(errs.ErrUnauthorized, "Invalid refresh token")
	}

	tokenHash := hashToken(refreshToken)
	userID, err := s.refreshTokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return "", nil, errs.WithMessage(errs.ErrUnauthorized, "Refresh token revoked or expired")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", nil, errs.WithMessage(errs.ErrUnauthorized, "User not found")
	}

	if user.UUID != userUUID {
		return "", nil, errs.WithMessage(errs.ErrUnauthorized, "Token mismatch")
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.UUID, user.Username, string(user.Role))
	if err != nil {
		return "", nil, fmt.Errorf("generate access token: %w", err)
	}

	return accessToken, user, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.refreshTokenRepo.Revoke(ctx, tokenHash)
}

func (s *AuthService) recordLoginAttempt(ctx context.Context, email, ip string, success bool) {
	attempt := &model.LoginAttempt{
		IPAddress:   ip,
		Email:       email,
		AttemptedAt: time.Now().UTC(),
		Success:     success,
	}
	_ = s.loginAttemptRepo.Create(ctx, attempt)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
