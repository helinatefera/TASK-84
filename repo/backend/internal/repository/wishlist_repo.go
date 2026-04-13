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

type wishlistRepo struct {
	db *database.DB
}

// NewWishlistRepository returns a new WishlistRepository backed by MySQL.
func NewWishlistRepository(db *database.DB) WishlistRepository {
	return &wishlistRepo{db: db}
}

func (r *wishlistRepo) Create(ctx context.Context, w *model.Wishlist) error {
	const q = `
		INSERT INTO wishlists (uuid, user_id, name, created_at, updated_at)
		VALUES (:uuid, :user_id, :name, :created_at, :updated_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, w)
	if err != nil {
		return fmt.Errorf("wishlist repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("wishlist repo create last insert id: %w", err)
	}
	w.ID = uint64(id)

	return nil
}

func (r *wishlistRepo) GetByID(ctx context.Context, id uint64) (*model.Wishlist, error) {
	const q = `SELECT id, uuid, user_id, name, created_at, updated_at
		FROM wishlists WHERE id = ?`

	var w model.Wishlist
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &w, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("wishlist repo get by id: %w", err)
	}
	return &w, nil
}

func (r *wishlistRepo) GetByUUID(ctx context.Context, uuid string) (*model.Wishlist, error) {
	const q = `SELECT id, uuid, user_id, name, created_at, updated_at
		FROM wishlists WHERE uuid = ?`

	var w model.Wishlist
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &w, q, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("wishlist repo get by uuid: %w", err)
	}
	return &w, nil
}

func (r *wishlistRepo) ListByUser(ctx context.Context, userID uint64) ([]*model.Wishlist, error) {
	const q = `SELECT id, uuid, user_id, name, created_at, updated_at
		FROM wishlists WHERE user_id = ? ORDER BY created_at DESC`

	var wishlists []*model.Wishlist
	err := sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &wishlists, q, userID)
	if err != nil {
		return nil, fmt.Errorf("wishlist repo list by user: %w", err)
	}
	return wishlists, nil
}

func (r *wishlistRepo) Update(ctx context.Context, w *model.Wishlist) error {
	const q = `UPDATE wishlists SET name = :name, updated_at = :updated_at WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, w)
	if err != nil {
		return fmt.Errorf("wishlist repo update: %w", err)
	}
	return nil
}

func (r *wishlistRepo) Delete(ctx context.Context, id uint64) error {
	const q = `DELETE FROM wishlists WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("wishlist repo delete: %w", err)
	}
	return nil
}

func (r *wishlistRepo) AddItem(ctx context.Context, wishlistID, itemID uint64) error {
	const q = `INSERT IGNORE INTO wishlist_items (wishlist_id, item_id, added_at)
		VALUES (?, ?, ?)`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, wishlistID, itemID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("wishlist repo add item: %w", err)
	}
	return nil
}

func (r *wishlistRepo) RemoveItem(ctx context.Context, wishlistID, itemID uint64) error {
	const q = `DELETE FROM wishlist_items WHERE wishlist_id = ? AND item_id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, wishlistID, itemID)
	if err != nil {
		return fmt.Errorf("wishlist repo remove item: %w", err)
	}
	return nil
}

func (r *wishlistRepo) ListItems(ctx context.Context, wishlistID uint64) ([]*model.WishlistItem, error) {
	const q = `SELECT id, wishlist_id, item_id, added_at
		FROM wishlist_items WHERE wishlist_id = ? ORDER BY added_at DESC`

	var items []*model.WishlistItem
	err := sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &items, q, wishlistID)
	if err != nil {
		return nil, fmt.Errorf("wishlist repo list items: %w", err)
	}
	return items, nil
}
