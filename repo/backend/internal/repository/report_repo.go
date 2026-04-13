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

type reportRepo struct {
	db *database.DB
}

// NewReportRepository returns a new ReportRepository backed by MySQL.
func NewReportRepository(db *database.DB) ReportRepository {
	return &reportRepo{db: db}
}

func (r *reportRepo) Create(ctx context.Context, rpt *model.Report) error {
	const q = `
		INSERT INTO reports (uuid, reporter_id, target_type, target_id, category, description,
			status, priority, assigned_to, resolved_at, resolution_note, user_visible_note,
			idempotency_key, created_at, updated_at)
		VALUES (:uuid, :reporter_id, :target_type, :target_id, :category, :description,
			:status, :priority, :assigned_to, :resolved_at, :resolution_note, :user_visible_note,
			:idempotency_key, :created_at, :updated_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, rpt)
	if err != nil {
		return fmt.Errorf("report repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("report repo create last insert id: %w", err)
	}
	rpt.ID = uint64(id)

	return nil
}

func (r *reportRepo) GetByID(ctx context.Context, id uint64) (*model.Report, error) {
	const q = `SELECT id, uuid, reporter_id, target_type, target_id, category, description,
		status, priority, assigned_to, resolved_at, resolution_note, user_visible_note,
		idempotency_key, created_at, updated_at
		FROM reports WHERE id = ?`

	var rpt model.Report
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &rpt, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("report repo get by id: %w", err)
	}
	return &rpt, nil
}

func (r *reportRepo) GetByUUID(ctx context.Context, uuid string) (*model.Report, error) {
	const q = `SELECT id, uuid, reporter_id, target_type, target_id, category, description,
		status, priority, assigned_to, resolved_at, resolution_note, user_visible_note,
		idempotency_key, created_at, updated_at
		FROM reports WHERE uuid = ?`

	var rpt model.Report
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &rpt, q, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("report repo get by uuid: %w", err)
	}
	return &rpt, nil
}

func (r *reportRepo) Update(ctx context.Context, rpt *model.Report) error {
	const q = `UPDATE reports SET status = :status, priority = :priority, assigned_to = :assigned_to,
		resolved_at = :resolved_at, resolution_note = :resolution_note, user_visible_note = :user_visible_note,
		updated_at = :updated_at WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, rpt)
	if err != nil {
		return fmt.Errorf("report repo update: %w", err)
	}
	return nil
}

func (r *reportRepo) ListQueue(ctx context.Context, status string, targetType string, page Pagination) ([]*model.Report, int64, error) {
	countQ := `SELECT COUNT(*) FROM reports WHERE 1=1`
	listQ := `SELECT id, uuid, reporter_id, target_type, target_id, category, description,
		status, priority, assigned_to, resolved_at, resolution_note, user_visible_note,
		idempotency_key, created_at, updated_at
		FROM reports WHERE 1=1`

	var args []interface{}

	if status != "" {
		countQ += ` AND status = ?`
		listQ += ` AND status = ?`
		args = append(args, status)
	}

	if targetType != "" {
		countQ += ` AND target_type = ?`
		listQ += ` AND target_type = ?`
		args = append(args, targetType)
	}

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("report repo list queue count: %w", err)
	}

	listQ += ` ORDER BY status ASC, priority ASC, created_at ASC LIMIT ? OFFSET ?`
	listArgs := append(args, page.PerPage, page.Offset())

	var reports []*model.Report
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &reports, listQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("report repo list queue select: %w", err)
	}

	return reports, total, nil
}

func (r *reportRepo) ListByReporter(ctx context.Context, reporterID uint64, page Pagination) ([]*model.Report, int64, error) {
	const countQ = `SELECT COUNT(*) FROM reports WHERE reporter_id = ?`
	const listQ = `SELECT id, uuid, reporter_id, target_type, target_id, category, description,
		status, priority, assigned_to, resolved_at, resolution_note, user_visible_note,
		idempotency_key, created_at, updated_at
		FROM reports WHERE reporter_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ, reporterID)
	if err != nil {
		return nil, 0, fmt.Errorf("report repo list by reporter count: %w", err)
	}

	var reports []*model.Report
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &reports, listQ, reporterID, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("report repo list by reporter select: %w", err)
	}

	return reports, total, nil
}
