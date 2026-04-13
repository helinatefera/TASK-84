package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
)

type auditLogRepo struct {
	db *database.DB
}

// NewAuditLogRepository returns a new AuditLogRepository backed by MySQL.
func NewAuditLogRepository(db *database.DB) AuditLogRepository {
	return &auditLogRepo{db: db}
}

func (r *auditLogRepo) Create(ctx context.Context, log *model.AuditLog) error {
	const q = `
		INSERT INTO audit_logs (actor_id, actor_role, action, target_type, target_id,
			ip_address, request_id, details, created_at)
		VALUES (:actor_id, :actor_role, :action, :target_type, :target_id,
			:ip_address, :request_id, :details, :created_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, log)
	if err != nil {
		return fmt.Errorf("audit log repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("audit log repo create last insert id: %w", err)
	}
	log.ID = uint64(id)

	return nil
}

func (r *auditLogRepo) List(ctx context.Context, page Pagination) ([]*model.AuditLog, int64, error) {
	const countQ = `SELECT COUNT(*) FROM audit_logs`
	const listQ = `SELECT id, actor_id, actor_role, action, target_type, target_id,
		ip_address, request_id, details, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ)
	if err != nil {
		return nil, 0, fmt.Errorf("audit log repo list count: %w", err)
	}

	var logs []*model.AuditLog
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &logs, listQ, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("audit log repo list select: %w", err)
	}

	return logs, total, nil
}
