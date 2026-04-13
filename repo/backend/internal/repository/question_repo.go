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

type questionRepo struct {
	db *database.DB
}

// NewQuestionRepository returns a new QuestionRepository backed by MySQL.
func NewQuestionRepository(db *database.DB) QuestionRepository {
	return &questionRepo{db: db}
}

func (r *questionRepo) Create(ctx context.Context, q *model.Question) error {
	const query = `
		INSERT INTO questions (uuid, item_id, user_id, body, is_deleted, created_at, updated_at)
		VALUES (:uuid, :item_id, :user_id, :body, :is_deleted, :created_at, :updated_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), query, q)
	if err != nil {
		return fmt.Errorf("question repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("question repo create last insert id: %w", err)
	}
	q.ID = uint64(id)

	return nil
}

func (r *questionRepo) GetByID(ctx context.Context, id uint64) (*model.Question, error) {
	const query = `SELECT id, uuid, item_id, user_id, body, is_deleted, created_at, updated_at
		FROM questions WHERE id = ?`

	var q model.Question
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &q, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("question repo get by id: %w", err)
	}
	return &q, nil
}

func (r *questionRepo) GetByUUID(ctx context.Context, uuid string) (*model.Question, error) {
	const query = `SELECT id, uuid, item_id, user_id, body, is_deleted, created_at, updated_at
		FROM questions WHERE uuid = ?`

	var q model.Question
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &q, query, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("question repo get by uuid: %w", err)
	}
	return &q, nil
}

func (r *questionRepo) ListByItem(ctx context.Context, itemID uint64, page Pagination) ([]*model.Question, int64, error) {
	const countQ = `SELECT COUNT(*) FROM questions WHERE item_id = ? AND is_deleted = 0`
	const listQ = `SELECT id, uuid, item_id, user_id, body, is_deleted, created_at, updated_at
		FROM questions WHERE item_id = ? AND is_deleted = 0
		ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ, itemID)
	if err != nil {
		return nil, 0, fmt.Errorf("question repo list by item count: %w", err)
	}

	var questions []*model.Question
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &questions, listQ, itemID, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("question repo list by item select: %w", err)
	}

	return questions, total, nil
}

func (r *questionRepo) Update(ctx context.Context, q *model.Question) error {
	const query = `UPDATE questions SET body = :body, updated_at = :updated_at WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), query, q)
	if err != nil {
		return fmt.Errorf("question repo update: %w", err)
	}
	return nil
}

func (r *questionRepo) SoftDelete(ctx context.Context, id uint64) error {
	const query = `UPDATE questions SET is_deleted = 1 WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("question repo soft delete: %w", err)
	}
	return nil
}
