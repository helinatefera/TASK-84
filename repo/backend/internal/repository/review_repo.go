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

type reviewRepo struct {
	db *database.DB
}

// NewReviewRepository returns a new ReviewRepository backed by MySQL.
func NewReviewRepository(db *database.DB) ReviewRepository {
	return &reviewRepo{db: db}
}

func (r *reviewRepo) Create(ctx context.Context, review *model.Review) error {
	const q = `
		INSERT INTO reviews (uuid, item_id, user_id, rating, body, fraud_status, idempotency_key, created_at, updated_at)
		VALUES (:uuid, :item_id, :user_id, :rating, :body, :fraud_status, :idempotency_key, :created_at, :updated_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, review)
	if err != nil {
		return fmt.Errorf("review repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("review repo create last insert id: %w", err)
	}
	review.ID = uint64(id)

	return nil
}

func (r *reviewRepo) GetByID(ctx context.Context, id uint64) (*model.Review, error) {
	const q = `SELECT id, uuid, item_id, user_id, rating, body, fraud_status, idempotency_key, created_at, updated_at
		FROM reviews WHERE id = ?`

	var review model.Review
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &review, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("review repo get by id: %w", err)
	}
	return &review, nil
}

func (r *reviewRepo) GetByUUID(ctx context.Context, uuid string) (*model.Review, error) {
	const q = `SELECT id, uuid, item_id, user_id, rating, body, fraud_status, idempotency_key, created_at, updated_at
		FROM reviews WHERE uuid = ?`

	var review model.Review
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &review, q, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("review repo get by uuid: %w", err)
	}
	return &review, nil
}

func (r *reviewRepo) ListByItem(ctx context.Context, itemID uint64, page Pagination) ([]*model.Review, int64, error) {
	const countQ = `SELECT COUNT(*) FROM reviews
		WHERE item_id = ? AND fraud_status NOT IN ('suspected_fraud', 'confirmed_fraud')`
	const listQ = `SELECT id, uuid, item_id, user_id, rating, body, fraud_status, idempotency_key, created_at, updated_at
		FROM reviews
		WHERE item_id = ? AND fraud_status NOT IN ('suspected_fraud', 'confirmed_fraud')
		ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ, itemID)
	if err != nil {
		return nil, 0, fmt.Errorf("review repo list by item count: %w", err)
	}

	var reviews []*model.Review
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &reviews, listQ, itemID, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("review repo list by item select: %w", err)
	}

	return reviews, total, nil
}

func (r *reviewRepo) Update(ctx context.Context, review *model.Review) error {
	const q = `UPDATE reviews SET rating = :rating, body = :body, fraud_status = :fraud_status,
		updated_at = :updated_at WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, review)
	if err != nil {
		return fmt.Errorf("review repo update: %w", err)
	}
	return nil
}

func (r *reviewRepo) Delete(ctx context.Context, id uint64) error {
	const q = `DELETE FROM reviews WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("review repo delete: %w", err)
	}
	return nil
}

func (r *reviewRepo) UpdateFraudStatus(ctx context.Context, id uint64, status model.FraudStatus) error {
	const q = `UPDATE reviews SET fraud_status = ? WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("review repo update fraud status: %w", err)
	}
	return nil
}

func (r *reviewRepo) GetItemsWithReviewsSince(ctx context.Context, since time.Time) ([]uint64, error) {
	const q = `SELECT DISTINCT item_id FROM reviews WHERE updated_at > ?`

	var itemIDs []uint64
	err := sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &itemIDs, q, since)
	if err != nil {
		return nil, fmt.Errorf("review repo get items with reviews since: %w", err)
	}
	return itemIDs, nil
}

func (r *reviewRepo) ComputeAggregate(ctx context.Context, itemID uint64) (*model.ItemRatingAggregate, error) {
	const q = `SELECT
		? AS item_id,
		COALESCE(AVG(rating), 0) AS avg_rating,
		COUNT(*) AS rating_count,
		SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END) AS rating_1,
		SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END) AS rating_2,
		SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END) AS rating_3,
		SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END) AS rating_4,
		SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END) AS rating_5,
		NOW() AS last_refreshed
		FROM reviews
		WHERE item_id = ? AND fraud_status NOT IN ('suspected_fraud', 'confirmed_fraud')`

	var agg model.ItemRatingAggregate
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &agg, q, itemID, itemID)
	if err != nil {
		return nil, fmt.Errorf("review repo compute aggregate: %w", err)
	}
	return &agg, nil
}
