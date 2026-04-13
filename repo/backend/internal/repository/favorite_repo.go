package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
)

type favoriteRepo struct {
	db *database.DB
}

// NewFavoriteRepository returns a new FavoriteRepository backed by MySQL.
func NewFavoriteRepository(db *database.DB) FavoriteRepository {
	return &favoriteRepo{db: db}
}

func (r *favoriteRepo) Add(ctx context.Context, userID, itemID uint64) error {
	const q = `INSERT IGNORE INTO favorites (user_id, item_id, created_at)
		VALUES (?, ?, ?)`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, userID, itemID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("favorite repo add: %w", err)
	}
	return nil
}

func (r *favoriteRepo) Remove(ctx context.Context, userID, itemID uint64) error {
	const q = `DELETE FROM favorites WHERE user_id = ? AND item_id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, userID, itemID)
	if err != nil {
		return fmt.Errorf("favorite repo remove: %w", err)
	}
	return nil
}

func (r *favoriteRepo) ListByUser(ctx context.Context, userID uint64, page Pagination) ([]*model.Favorite, int64, error) {
	const countQ = `SELECT COUNT(*) FROM favorites WHERE user_id = ?`
	const listQ = `SELECT id, user_id, item_id, created_at
		FROM favorites WHERE user_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("favorite repo list by user count: %w", err)
	}

	var favs []*model.Favorite
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &favs, listQ, userID, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("favorite repo list by user select: %w", err)
	}

	return favs, total, nil
}

func (r *favoriteRepo) Exists(ctx context.Context, userID, itemID uint64) (bool, error) {
	const q = `SELECT COUNT(*) FROM favorites WHERE user_id = ? AND item_id = ?`

	var count int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &count, q, userID, itemID)
	if err != nil {
		return false, fmt.Errorf("favorite repo exists: %w", err)
	}
	return count > 0, nil
}
