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

type sensitiveWordRuleRepo struct {
	db *database.DB
}

// NewSensitiveWordRuleRepository returns a new SensitiveWordRuleRepository backed by MySQL.
func NewSensitiveWordRuleRepository(db *database.DB) SensitiveWordRuleRepository {
	return &sensitiveWordRuleRepo{db: db}
}

func (r *sensitiveWordRuleRepo) Create(ctx context.Context, rule *model.SensitiveWordRule) error {
	const q = `
		INSERT INTO sensitive_word_rules (pattern, action, replacement, version, is_active, created_by, created_at, updated_at)
		VALUES (:pattern, :action, :replacement, :version, :is_active, :created_by, :created_at, :updated_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, rule)
	if err != nil {
		return fmt.Errorf("sensitive word rule repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("sensitive word rule repo create last insert id: %w", err)
	}
	rule.ID = uint64(id)

	return nil
}

func (r *sensitiveWordRuleRepo) GetByID(ctx context.Context, id uint64) (*model.SensitiveWordRule, error) {
	const q = `SELECT id, pattern, action, replacement, version, is_active, created_by, created_at, updated_at
		FROM sensitive_word_rules WHERE id = ?`

	var rule model.SensitiveWordRule
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &rule, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("sensitive word rule repo get by id: %w", err)
	}
	return &rule, nil
}

func (r *sensitiveWordRuleRepo) Update(ctx context.Context, rule *model.SensitiveWordRule) error {
	const q = `UPDATE sensitive_word_rules SET pattern = :pattern, action = :action,
		replacement = :replacement, version = :version, is_active = :is_active,
		updated_at = :updated_at WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, rule)
	if err != nil {
		return fmt.Errorf("sensitive word rule repo update: %w", err)
	}
	return nil
}

func (r *sensitiveWordRuleRepo) Delete(ctx context.Context, id uint64) error {
	const q = `DELETE FROM sensitive_word_rules WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("sensitive word rule repo delete: %w", err)
	}
	return nil
}

func (r *sensitiveWordRuleRepo) ListActive(ctx context.Context) ([]*model.SensitiveWordRule, error) {
	const q = `SELECT id, pattern, action, replacement, version, is_active, created_by, created_at, updated_at
		FROM sensitive_word_rules WHERE is_active = 1 ORDER BY id ASC`

	var rules []*model.SensitiveWordRule
	err := sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &rules, q)
	if err != nil {
		return nil, fmt.Errorf("sensitive word rule repo list active: %w", err)
	}
	return rules, nil
}
