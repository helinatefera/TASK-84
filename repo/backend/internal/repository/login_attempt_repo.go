package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
)

type loginAttemptRepo struct {
	db *database.DB
}

// NewLoginAttemptRepository returns a new LoginAttemptRepository backed by MySQL.
func NewLoginAttemptRepository(db *database.DB) LoginAttemptRepository {
	return &loginAttemptRepo{db: db}
}

func (r *loginAttemptRepo) Create(ctx context.Context, attempt *model.LoginAttempt) error {
	const q = `INSERT INTO login_attempts (ip_address, email, attempted_at, success)
		VALUES (:ip_address, :email, :attempted_at, :success)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, attempt)
	if err != nil {
		return fmt.Errorf("login attempt repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("login attempt repo create last insert id: %w", err)
	}
	attempt.ID = uint64(id)

	return nil
}

func (r *loginAttemptRepo) CountRecentFailed(ctx context.Context, email string, ip string, window time.Duration) (int, error) {
	const q = `SELECT COUNT(*) FROM login_attempts
		WHERE (email = ? OR ip_address = ?) AND success = 0 AND attempted_at > ?`

	cutoff := time.Now().UTC().Add(-window)

	var count int
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &count, q, email, ip, cutoff)
	if err != nil {
		return 0, fmt.Errorf("login attempt repo count recent failed: %w", err)
	}
	return count, nil
}
