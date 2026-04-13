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

type itemRepo struct {
	db *database.DB
}

// NewItemRepository returns a new ItemRepository backed by MySQL.
func NewItemRepository(db *database.DB) ItemRepository {
	return &itemRepo{db: db}
}

func (r *itemRepo) Create(ctx context.Context, item *model.Item) error {
	const q = `
		INSERT INTO items (uuid, title, description, category, lifecycle_state, created_by,
			published_at, archived_at, created_at, updated_at)
		VALUES (:uuid, :title, :description, :category, :lifecycle_state, :created_by,
			:published_at, :archived_at, :created_at, :updated_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, item)
	if err != nil {
		return fmt.Errorf("item repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("item repo create last insert id: %w", err)
	}
	item.ID = uint64(id)

	return nil
}

func (r *itemRepo) GetByID(ctx context.Context, id uint64) (*model.Item, error) {
	const q = `SELECT id, uuid, title, description, category, lifecycle_state, created_by,
		published_at, archived_at, created_at, updated_at
		FROM items WHERE id = ?`

	var item model.Item
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &item, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("item repo get by id: %w", err)
	}
	return &item, nil
}

func (r *itemRepo) GetByUUID(ctx context.Context, uuid string) (*model.Item, error) {
	const q = `SELECT id, uuid, title, description, category, lifecycle_state, created_by,
		published_at, archived_at, created_at, updated_at
		FROM items WHERE uuid = ?`

	var item model.Item
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &item, q, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("item repo get by uuid: %w", err)
	}
	return &item, nil
}

func (r *itemRepo) Update(ctx context.Context, item *model.Item) error {
	const q = `UPDATE items SET title = :title, description = :description, category = :category,
		lifecycle_state = :lifecycle_state, published_at = :published_at, archived_at = :archived_at,
		updated_at = :updated_at WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, item)
	if err != nil {
		return fmt.Errorf("item repo update: %w", err)
	}
	return nil
}

func (r *itemRepo) ListPublished(ctx context.Context, search string, category string, page Pagination) ([]*model.Item, int64, error) {
	countQ := `SELECT COUNT(*) FROM items WHERE lifecycle_state = 'published'`
	listQ := `SELECT id, uuid, title, description, category, lifecycle_state, created_by,
		published_at, archived_at, created_at, updated_at
		FROM items WHERE lifecycle_state = 'published'`

	var args []interface{}

	if search != "" {
		countQ += ` AND MATCH(title, description) AGAINST(? IN BOOLEAN MODE)`
		listQ += ` AND MATCH(title, description) AGAINST(? IN BOOLEAN MODE)`
		args = append(args, search)
	}

	if category != "" {
		countQ += ` AND category = ?`
		listQ += ` AND category = ?`
		args = append(args, category)
	}

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("item repo list published count: %w", err)
	}

	listQ += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, page.PerPage, page.Offset())

	var items []*model.Item
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &items, listQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("item repo list published select: %w", err)
	}

	return items, total, nil
}

func (r *itemRepo) GetRatingAggregate(ctx context.Context, itemID uint64) (*model.ItemRatingAggregate, error) {
	const q = `SELECT item_id, avg_rating, rating_count, rating_1, rating_2, rating_3, rating_4, rating_5, last_refreshed
		FROM item_rating_aggregates WHERE item_id = ?`

	var agg model.ItemRatingAggregate
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &agg, q, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("item repo get rating aggregate: %w", err)
	}
	return &agg, nil
}

func (r *itemRepo) UpsertRatingAggregate(ctx context.Context, agg *model.ItemRatingAggregate) error {
	const q = `
		INSERT INTO item_rating_aggregates (item_id, avg_rating, rating_count, rating_1, rating_2, rating_3, rating_4, rating_5, last_refreshed)
		VALUES (:item_id, :avg_rating, :rating_count, :rating_1, :rating_2, :rating_3, :rating_4, :rating_5, :last_refreshed)
		ON DUPLICATE KEY UPDATE
			avg_rating = VALUES(avg_rating),
			rating_count = VALUES(rating_count),
			rating_1 = VALUES(rating_1),
			rating_2 = VALUES(rating_2),
			rating_3 = VALUES(rating_3),
			rating_4 = VALUES(rating_4),
			rating_5 = VALUES(rating_5),
			last_refreshed = VALUES(last_refreshed)`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, agg)
	if err != nil {
		return fmt.Errorf("item repo upsert rating aggregate: %w", err)
	}
	return nil
}
