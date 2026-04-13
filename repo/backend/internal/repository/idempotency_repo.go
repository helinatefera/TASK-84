package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
)

type idempotencyKeyRepo struct {
	db *database.DB
}

// NewIdempotencyKeyRepository returns a new IdempotencyKeyRepository backed by MySQL.
func NewIdempotencyKeyRepository(db *database.DB) IdempotencyKeyRepository {
	return &idempotencyKeyRepo{db: db}
}

func (r *idempotencyKeyRepo) Get(ctx context.Context, keyHash string, userID uint64) (*model.IdempotencyKey, error) {
	const q = `SELECT id, key_hash, user_id, endpoint, response_code, response_body, created_at, expires_at
		FROM idempotency_keys WHERE key_hash = ? AND user_id = ?`

	var k model.IdempotencyKey
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &k, q, keyHash, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("idempotency key repo get: %w", err)
	}
	return &k, nil
}

func (r *idempotencyKeyRepo) Create(ctx context.Context, k *model.IdempotencyKey) error {
	const q = `
		INSERT INTO idempotency_keys (key_hash, user_id, endpoint, response_code, response_body, created_at, expires_at)
		VALUES (:key_hash, :user_id, :endpoint, :response_code, :response_body, :created_at, :expires_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, k)
	if err != nil {
		return fmt.Errorf("idempotency key repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("idempotency key repo create last insert id: %w", err)
	}
	k.ID = uint64(id)

	return nil
}

func (r *idempotencyKeyRepo) UpdateResponse(ctx context.Context, keyHash string, userID uint64, code uint16, body string) error {
	const q = `UPDATE idempotency_keys SET response_code = ?, response_body = ?
		WHERE key_hash = ? AND user_id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, code, body, keyHash, userID)
	if err != nil {
		return fmt.Errorf("idempotency key repo update response: %w", err)
	}
	return nil
}

func (r *idempotencyKeyRepo) DeleteExpired(ctx context.Context) (int64, error) {
	const q = `DELETE FROM idempotency_keys WHERE expires_at < NOW()`

	result, err := r.db.ExtContext(ctx).ExecContext(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("idempotency key repo delete expired: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("idempotency key repo delete expired rows affected: %w", err)
	}
	return rows, nil
}
