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

type appealRepo struct {
	db *database.DB
}

// NewAppealRepository returns a new AppealRepository backed by MySQL.
func NewAppealRepository(db *database.DB) AppealRepository {
	return &appealRepo{db: db}
}

func (r *appealRepo) Create(ctx context.Context, a *model.Appeal) error {
	const q = `
		INSERT INTO appeals (uuid, report_id, user_id, body, status, reviewed_by, reviewed_at, created_at)
		VALUES (:uuid, :report_id, :user_id, :body, :status, :reviewed_by, :reviewed_at, :created_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, a)
	if err != nil {
		return fmt.Errorf("appeal repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("appeal repo create last insert id: %w", err)
	}
	a.ID = uint64(id)

	return nil
}

func (r *appealRepo) GetByReportID(ctx context.Context, reportID uint64) (*model.Appeal, error) {
	const q = `SELECT id, uuid, report_id, user_id, body, status, reviewed_by, reviewed_at, created_at
		FROM appeals WHERE report_id = ?`

	var a model.Appeal
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &a, q, reportID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("appeal repo get by report id: %w", err)
	}
	return &a, nil
}

func (r *appealRepo) Update(ctx context.Context, a *model.Appeal) error {
	const q = `UPDATE appeals SET status = :status, reviewed_by = :reviewed_by,
		reviewed_at = :reviewed_at WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, a)
	if err != nil {
		return fmt.Errorf("appeal repo update: %w", err)
	}
	return nil
}

func (r *appealRepo) Resubmit(ctx context.Context, id uint64, body string) error {
	const q = `UPDATE appeals SET body = ?, status = 'pending', reviewed_by = NULL, reviewed_at = NULL WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, body, id)
	if err != nil {
		return fmt.Errorf("appeal repo resubmit: %w", err)
	}
	return nil
}

func (r *appealRepo) List(ctx context.Context, status string, p Pagination) ([]*model.Appeal, int64, error) {
	where := "1 = 1"
	var args []interface{}
	if status != "" {
		where = "status = ?"
		args = append(args, status)
	}

	var total int64
	countQ := "SELECT COUNT(*) FROM appeals WHERE " + where
	err := r.db.ExtContext(ctx).QueryRowxContext(ctx, countQ, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("appeal repo list count: %w", err)
	}

	dataQ := fmt.Sprintf("SELECT * FROM appeals WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?", where)
	dataArgs := append(args, p.PerPage, p.Offset())
	var appeals []*model.Appeal
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &appeals, dataQ, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("appeal repo list: %w", err)
	}
	return appeals, total, nil
}
