package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
)

type moderationNoteRepo struct {
	db *database.DB
}

// NewModerationNoteRepository returns a new ModerationNoteRepository backed by MySQL.
func NewModerationNoteRepository(db *database.DB) ModerationNoteRepository {
	return &moderationNoteRepo{db: db}
}

func (r *moderationNoteRepo) Create(ctx context.Context, n *model.ModerationNote) error {
	const q = `
		INSERT INTO moderation_notes (report_id, author_id, body, is_internal, created_at)
		VALUES (:report_id, :author_id, :body, :is_internal, :created_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, n)
	if err != nil {
		return fmt.Errorf("moderation note repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("moderation note repo create last insert id: %w", err)
	}
	n.ID = uint64(id)

	return nil
}

func (r *moderationNoteRepo) ListByReport(ctx context.Context, reportID uint64) ([]*model.ModerationNote, error) {
	const q = `SELECT id, report_id, author_id, body, is_internal, created_at
		FROM moderation_notes WHERE report_id = ? ORDER BY created_at ASC`

	var notes []*model.ModerationNote
	err := sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &notes, q, reportID)
	if err != nil {
		return nil, fmt.Errorf("moderation note repo list by report: %w", err)
	}
	return notes, nil
}
