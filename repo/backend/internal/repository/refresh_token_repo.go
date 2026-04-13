package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/pkg/database"
)

type refreshTokenRepo struct {
	db *database.DB
}

// NewRefreshTokenRepository returns a new RefreshTokenRepository backed by MySQL.
func NewRefreshTokenRepository(db *database.DB) RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(ctx context.Context, userID uint64, tokenHash string, expiresAt time.Time) error {
	const q = `INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked, created_at)
		VALUES (?, ?, ?, 0, ?)`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, userID, tokenHash, expiresAt, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("refresh token repo create: %w", err)
	}
	return nil
}

func (r *refreshTokenRepo) GetByHash(ctx context.Context, tokenHash string) (uint64, error) {
	const q = `SELECT user_id FROM refresh_tokens
		WHERE token_hash = ? AND revoked = 0 AND expires_at > NOW()`

	var userID uint64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &userID, q, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("refresh token not found or expired")
		}
		return 0, fmt.Errorf("refresh token repo get by hash: %w", err)
	}
	return userID, nil
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, tokenHash string) error {
	const q = `UPDATE refresh_tokens SET revoked = 1 WHERE token_hash = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, tokenHash)
	if err != nil {
		return fmt.Errorf("refresh token repo revoke: %w", err)
	}
	return nil
}

func (r *refreshTokenRepo) RevokeAllForUser(ctx context.Context, userID uint64) error {
	const q = `UPDATE refresh_tokens SET revoked = 1 WHERE user_id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("refresh token repo revoke all for user: %w", err)
	}
	return nil
}
