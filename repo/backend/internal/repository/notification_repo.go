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

// --- NotificationRepository ---

type notificationRepo struct {
	db *database.DB
}

// NewNotificationRepository returns a new NotificationRepository backed by MySQL.
func NewNotificationRepository(db *database.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, n *model.Notification) error {
	const q = `
		INSERT INTO notifications (user_id, template_key, locale, rendered_subject, rendered_body,
			data, is_read, read_at, created_at)
		VALUES (:user_id, :template_key, :locale, :rendered_subject, :rendered_body,
			:data, :is_read, :read_at, :created_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, n)
	if err != nil {
		return fmt.Errorf("notification repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("notification repo create last insert id: %w", err)
	}
	n.ID = uint64(id)

	return nil
}

func (r *notificationRepo) GetByID(ctx context.Context, id uint64) (*model.Notification, error) {
	const q = `SELECT id, user_id, template_key, locale, rendered_subject, rendered_body,
		data, is_read, read_at, created_at
		FROM notifications WHERE id = ?`

	var n model.Notification
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &n, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("notification repo get by id: %w", err)
	}
	return &n, nil
}

func (r *notificationRepo) ListByUser(ctx context.Context, userID uint64, unreadOnly bool, page Pagination) ([]*model.Notification, int64, error) {
	countQ := `SELECT COUNT(*) FROM notifications WHERE user_id = ?`
	listQ := `SELECT id, user_id, template_key, locale, rendered_subject, rendered_body,
		data, is_read, read_at, created_at
		FROM notifications WHERE user_id = ?`

	var args []interface{}
	args = append(args, userID)

	if unreadOnly {
		countQ += ` AND is_read = 0`
		listQ += ` AND is_read = 0`
	}

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("notification repo list by user count: %w", err)
	}

	listQ += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, page.PerPage, page.Offset())

	var notifications []*model.Notification
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &notifications, listQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("notification repo list by user select: %w", err)
	}

	return notifications, total, nil
}

func (r *notificationRepo) MarkRead(ctx context.Context, id uint64) error {
	const q = `UPDATE notifications SET is_read = 1, read_at = ? WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("notification repo mark read: %w", err)
	}
	return nil
}

func (r *notificationRepo) MarkAllRead(ctx context.Context, userID uint64) error {
	const q = `UPDATE notifications SET is_read = 1, read_at = ? WHERE user_id = ? AND is_read = 0`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("notification repo mark all read: %w", err)
	}
	return nil
}

func (r *notificationRepo) UnreadCount(ctx context.Context, userID uint64) (int64, error) {
	const q = `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`

	var count int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &count, q, userID)
	if err != nil {
		return 0, fmt.Errorf("notification repo unread count: %w", err)
	}
	return count, nil
}

// --- NotificationTemplateRepository ---

type notificationTemplateRepo struct {
	db *database.DB
}

// NewNotificationTemplateRepository returns a new NotificationTemplateRepository backed by MySQL.
func NewNotificationTemplateRepository(db *database.DB) NotificationTemplateRepository {
	return &notificationTemplateRepo{db: db}
}

func (r *notificationTemplateRepo) GetByKeyAndLocale(ctx context.Context, key, locale string) (*model.NotificationTemplate, error) {
	const q = `SELECT id, template_key, locale, subject, body_template, created_at, updated_at
		FROM notification_templates WHERE template_key = ? AND locale = ?`

	var tmpl model.NotificationTemplate
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &tmpl, q, key, locale)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("notification template repo get by key and locale: %w", err)
	}
	return &tmpl, nil
}
