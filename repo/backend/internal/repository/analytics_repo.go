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

// --- AnalyticsSessionRepository ---

type analyticsSessionRepo struct {
	db *database.DB
}

// NewAnalyticsSessionRepository returns a new AnalyticsSessionRepository backed by MySQL.
func NewAnalyticsSessionRepository(db *database.DB) AnalyticsSessionRepository {
	return &analyticsSessionRepo{db: db}
}

func (r *analyticsSessionRepo) Create(ctx context.Context, s *model.AnalyticsSession) error {
	const q = `
		INSERT INTO analytics_sessions (session_uuid, user_id, item_id, experiment_variant_id,
			started_at, ended_at, last_active_at, user_agent, ip_address)
		VALUES (:session_uuid, :user_id, :item_id, :experiment_variant_id,
			:started_at, :ended_at, :last_active_at, :user_agent, :ip_address)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, s)
	if err != nil {
		return fmt.Errorf("analytics session repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("analytics session repo create last insert id: %w", err)
	}
	s.ID = uint64(id)

	return nil
}

func (r *analyticsSessionRepo) GetByUUID(ctx context.Context, uuid string) (*model.AnalyticsSession, error) {
	const q = `SELECT id, session_uuid, user_id, item_id, experiment_variant_id,
		started_at, ended_at, last_active_at, user_agent, ip_address
		FROM analytics_sessions WHERE session_uuid = ?`

	var s model.AnalyticsSession
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &s, q, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("analytics session repo get by uuid: %w", err)
	}
	return &s, nil
}

func (r *analyticsSessionRepo) UpdateHeartbeat(ctx context.Context, id uint64) error {
	const q = `UPDATE analytics_sessions SET last_active_at = ? WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("analytics session repo update heartbeat: %w", err)
	}
	return nil
}

func (r *analyticsSessionRepo) EndSession(ctx context.Context, id uint64) error {
	const q = `UPDATE analytics_sessions SET ended_at = ? WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("analytics session repo end session: %w", err)
	}
	return nil
}

func (r *analyticsSessionRepo) GetExpiredSessions(ctx context.Context, inactiveThreshold time.Duration) ([]*model.AnalyticsSession, error) {
	const q = `SELECT id, session_uuid, user_id, item_id, experiment_variant_id,
		started_at, ended_at, last_active_at, user_agent, ip_address
		FROM analytics_sessions
		WHERE ended_at IS NULL AND last_active_at < ?`

	cutoff := time.Now().UTC().Add(-inactiveThreshold)

	var sessions []*model.AnalyticsSession
	err := sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &sessions, q, cutoff)
	if err != nil {
		return nil, fmt.Errorf("analytics session repo get expired sessions: %w", err)
	}
	return sessions, nil
}

// --- BehaviorEventRepository ---

type behaviorEventRepo struct {
	db *database.DB
}

// NewBehaviorEventRepository returns a new BehaviorEventRepository backed by MySQL.
func NewBehaviorEventRepository(db *database.DB) BehaviorEventRepository {
	return &behaviorEventRepo{db: db}
}

func (r *behaviorEventRepo) Create(ctx context.Context, e *model.BehaviorEvent) error {
	const q = `
		INSERT INTO behavior_events (session_id, user_id, event_type, item_id, dwell_seconds,
			event_data, client_ts, server_ts, dedup_hash)
		VALUES (:session_id, :user_id, :event_type, :item_id, :dwell_seconds,
			:event_data, :client_ts, :server_ts, :dedup_hash)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, e)
	if err != nil {
		return fmt.Errorf("behavior event repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("behavior event repo create last insert id: %w", err)
	}
	e.ID = uint64(id)

	return nil
}

func (r *behaviorEventRepo) CreateBatch(ctx context.Context, events []*model.BehaviorEvent) (int, error) {
	const q = `
		INSERT INTO behavior_events (session_id, user_id, event_type, item_id, dwell_seconds,
			event_data, client_ts, server_ts, dedup_hash)
		VALUES (:session_id, :user_id, :event_type, :item_id, :dwell_seconds,
			:event_data, :client_ts, :server_ts, :dedup_hash)
		ON DUPLICATE KEY UPDATE id = id`

	inserted := 0
	for _, e := range events {
		result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, e)
		if err != nil {
			return inserted, fmt.Errorf("behavior event repo create batch: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return inserted, fmt.Errorf("behavior event repo create batch rows affected: %w", err)
		}
		if rows > 0 {
			id, err := result.LastInsertId()
			if err == nil && id > 0 {
				e.ID = uint64(id)
			}
			inserted++
		}
	}

	return inserted, nil
}

func (r *behaviorEventRepo) ListBySession(ctx context.Context, sessionID uint64) ([]*model.BehaviorEvent, error) {
	const q = `SELECT id, session_id, user_id, event_type, item_id, dwell_seconds,
		event_data, client_ts, server_ts, dedup_hash
		FROM behavior_events WHERE session_id = ? ORDER BY server_ts ASC`

	var events []*model.BehaviorEvent
	err := sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &events, q, sessionID)
	if err != nil {
		return nil, fmt.Errorf("behavior event repo list by session: %w", err)
	}
	return events, nil
}

func (r *behaviorEventRepo) IncrementUserEventCount(ctx context.Context, userID uint64, hourBucket time.Time) error {
	const q = `INSERT INTO user_event_counts (user_id, hour_bucket, event_count)
		VALUES (?, ?, 1)
		ON DUPLICATE KEY UPDATE event_count = event_count + 1`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, userID, hourBucket)
	if err != nil {
		return fmt.Errorf("behavior event repo increment user event count: %w", err)
	}
	return nil
}

func (r *behaviorEventRepo) GetUserEventCount(ctx context.Context, userID uint64, hourBucket time.Time) (uint32, error) {
	const q = `SELECT COALESCE(event_count, 0) FROM user_event_counts
		WHERE user_id = ? AND hour_bucket = ?`

	var count uint32
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &count, q, userID, hourBucket)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("behavior event repo get user event count: %w", err)
	}
	return count, nil
}
