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

type answerRepo struct {
	db *database.DB
}

// NewAnswerRepository returns a new AnswerRepository backed by MySQL.
func NewAnswerRepository(db *database.DB) AnswerRepository {
	return &answerRepo{db: db}
}

func (r *answerRepo) Create(ctx context.Context, a *model.Answer) error {
	const q = `
		INSERT INTO answers (uuid, question_id, user_id, body, is_deleted, created_at, updated_at)
		VALUES (:uuid, :question_id, :user_id, :body, :is_deleted, :created_at, :updated_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, a)
	if err != nil {
		return fmt.Errorf("answer repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("answer repo create last insert id: %w", err)
	}
	a.ID = uint64(id)

	return nil
}

func (r *answerRepo) GetByID(ctx context.Context, id uint64) (*model.Answer, error) {
	const q = `SELECT id, uuid, question_id, user_id, body, is_deleted, created_at, updated_at
		FROM answers WHERE id = ?`

	var a model.Answer
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &a, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("answer repo get by id: %w", err)
	}
	return &a, nil
}

func (r *answerRepo) GetByUUID(ctx context.Context, uuid string) (*model.Answer, error) {
	const q = `SELECT id, uuid, question_id, user_id, body, is_deleted, created_at, updated_at
		FROM answers WHERE uuid = ?`

	var a model.Answer
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &a, q, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("answer repo get by uuid: %w", err)
	}
	return &a, nil
}

func (r *answerRepo) ListByQuestion(ctx context.Context, questionID uint64, page Pagination) ([]*model.Answer, int64, error) {
	const countQ = `SELECT COUNT(*) FROM answers WHERE question_id = ? AND is_deleted = 0`
	const listQ = `SELECT id, uuid, question_id, user_id, body, is_deleted, created_at, updated_at
		FROM answers WHERE question_id = ? AND is_deleted = 0
		ORDER BY created_at ASC LIMIT ? OFFSET ?`

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ, questionID)
	if err != nil {
		return nil, 0, fmt.Errorf("answer repo list by question count: %w", err)
	}

	var answers []*model.Answer
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &answers, listQ, questionID, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("answer repo list by question select: %w", err)
	}

	return answers, total, nil
}

func (r *answerRepo) Update(ctx context.Context, a *model.Answer) error {
	const q = `UPDATE answers SET body = :body, updated_at = :updated_at WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, a)
	if err != nil {
		return fmt.Errorf("answer repo update: %w", err)
	}
	return nil
}

func (r *answerRepo) SoftDelete(ctx context.Context, id uint64) error {
	const q = `UPDATE answers SET is_deleted = 1 WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("answer repo soft delete: %w", err)
	}
	return nil
}
