package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
)

type userRepo struct {
	db *database.DB
}

// NewUserRepository returns a new UserRepository backed by MySQL.
func NewUserRepository(db *database.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	const q = `
		INSERT INTO users (uuid, username, email, password_hash, role, is_active, created_at, updated_at)
		VALUES (:uuid, :username, :email, :password_hash, :role, :is_active, :created_at, :updated_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, user)
	if err != nil {
		return fmt.Errorf("user repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("user repo create last insert id: %w", err)
	}
	user.ID = uint64(id)

	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	const q = `SELECT id, uuid, username, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE id = ?`

	var user model.User
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &user, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repo get by id: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByUUID(ctx context.Context, uuid string) (*model.User, error) {
	const q = `SELECT id, uuid, username, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE uuid = ?`

	var user model.User
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &user, q, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repo get by uuid: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	const q = `SELECT id, uuid, username, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE username = ?`

	var user model.User
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &user, q, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repo get by username: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `SELECT id, uuid, username, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE email = ?`

	var user model.User
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &user, q, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repo get by email: %w", err)
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	const q = `UPDATE users SET username = :username, email = :email, role = :role,
		is_active = :is_active, updated_at = :updated_at WHERE id = :id`

	user.UpdatedAt = time.Now().UTC()

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, user)
	if err != nil {
		return fmt.Errorf("user repo update: %w", err)
	}
	return nil
}

func (r *userRepo) List(ctx context.Context, page Pagination) ([]*model.User, int64, error) {
	const countQ = `SELECT COUNT(*) FROM users`
	const listQ = `SELECT id, uuid, username, email, password_hash, role, is_active, created_at, updated_at
		FROM users ORDER BY id ASC LIMIT ? OFFSET ?`

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ)
	if err != nil {
		return nil, 0, fmt.Errorf("user repo list count: %w", err)
	}

	var users []*model.User
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &users, listQ, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("user repo list select: %w", err)
	}

	return users, total, nil
}

func (r *userRepo) UpdateRole(ctx context.Context, id uint64, role model.Role) error {
	const q = `UPDATE users SET role = ?, updated_at = ? WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, role, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("user repo update role: %w", err)
	}
	return nil
}

func (r *userRepo) SetActive(ctx context.Context, id uint64, active bool) error {
	const q = `UPDATE users SET is_active = ?, updated_at = ? WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, active, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("user repo set active: %w", err)
	}
	return nil
}
