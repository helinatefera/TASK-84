package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/localinsights/portal/internal/config"
)

type Claims struct {
	jwtlib.RegisteredClaims
	UserID   uint64 `json:"uid"`
	UserUUID string `json:"uuid"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(cfg config.JWTConfig) *Manager {
	return &Manager{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}
}

func (m *Manager) GenerateAccessToken(userID uint64, userUUID, username, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.accessTTL)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			Issuer:    "local-insights",
		},
		UserID:   userID,
		UserUUID: userUUID,
		Username: username,
		Role:     role,
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) GenerateRefreshToken(userID uint64, userUUID string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.refreshTTL)

	claims := jwtlib.RegisteredClaims{
		ExpiresAt: jwtlib.NewNumericDate(expiresAt),
		IssuedAt:  jwtlib.NewNumericDate(now),
		Subject:   userUUID,
		Issuer:    "local-insights-refresh",
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (m *Manager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (m *Manager) ValidateRefreshToken(tokenStr string) (string, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &jwtlib.RegisteredClaims{}, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*jwtlib.RegisteredClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	return claims.Subject, nil
}
