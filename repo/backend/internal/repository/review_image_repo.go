package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
)

type reviewImageRepo struct {
	db *database.DB
}

// NewReviewImageRepository returns a new ReviewImageRepository backed by MySQL.
func NewReviewImageRepository(db *database.DB) ReviewImageRepository {
	return &reviewImageRepo{db: db}
}

func (r *reviewImageRepo) Create(ctx context.Context, ri *model.ReviewImage) error {
	const q = `
		INSERT INTO review_images (review_id, image_id, sort_order)
		VALUES (:review_id, :image_id, :sort_order)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, ri)
	if err != nil {
		return fmt.Errorf("review image repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("review image repo create last insert id: %w", err)
	}
	ri.ID = uint64(id)

	return nil
}

func (r *reviewImageRepo) ListByReview(ctx context.Context, reviewID uint64) ([]*model.ReviewImage, error) {
	const q = `SELECT id, review_id, image_id, sort_order
		FROM review_images WHERE review_id = ? ORDER BY sort_order ASC`

	var images []*model.ReviewImage
	err := sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &images, q, reviewID)
	if err != nil {
		return nil, fmt.Errorf("review image repo list by review: %w", err)
	}
	return images, nil
}

func (r *reviewImageRepo) DeleteByReview(ctx context.Context, reviewID uint64) error {
	const q = `DELETE FROM review_images WHERE review_id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, reviewID)
	if err != nil {
		return fmt.Errorf("review image repo delete by review: %w", err)
	}
	return nil
}
