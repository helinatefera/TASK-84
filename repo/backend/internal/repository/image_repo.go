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

type imageRepo struct {
	db *database.DB
}

// NewImageRepository returns a new ImageRepository backed by MySQL.
func NewImageRepository(db *database.DB) ImageRepository {
	return &imageRepo{db: db}
}

func (r *imageRepo) Create(ctx context.Context, img *model.Image) error {
	const q = `
		INSERT INTO images (sha256_hash, original_name, mime_type, file_size, storage_path,
			width, height, status, quarantine_reason, uploaded_by, created_at)
		VALUES (:sha256_hash, :original_name, :mime_type, :file_size, :storage_path,
			:width, :height, :status, :quarantine_reason, :uploaded_by, :created_at)`

	result, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, img)
	if err != nil {
		return fmt.Errorf("image repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("image repo create last insert id: %w", err)
	}
	img.ID = uint64(id)

	return nil
}

func (r *imageRepo) GetByHash(ctx context.Context, hash string) (*model.Image, error) {
	const q = `SELECT id, sha256_hash, original_name, mime_type, file_size, storage_path,
		width, height, status, quarantine_reason, uploaded_by, created_at
		FROM images WHERE sha256_hash = ?`

	var img model.Image
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &img, q, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("image repo get by hash: %w", err)
	}
	return &img, nil
}

func (r *imageRepo) GetByID(ctx context.Context, id uint64) (*model.Image, error) {
	const q = `SELECT id, sha256_hash, original_name, mime_type, file_size, storage_path,
		width, height, status, quarantine_reason, uploaded_by, created_at
		FROM images WHERE id = ?`

	var img model.Image
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &img, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("image repo get by id: %w", err)
	}
	return &img, nil
}

func (r *imageRepo) UpdateStatus(ctx context.Context, id uint64, status string, reason *string) error {
	const q = `UPDATE images SET status = ?, quarantine_reason = ? WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, status, reason, id)
	if err != nil {
		return fmt.Errorf("image repo update status: %w", err)
	}
	return nil
}

func (r *imageRepo) ListQuarantined(ctx context.Context, page Pagination) ([]*model.Image, int64, error) {
	const countQ = `SELECT COUNT(*) FROM images WHERE status = 'quarantined'`
	const listQ = `SELECT id, sha256_hash, original_name, mime_type, file_size, storage_path,
		width, height, status, quarantine_reason, uploaded_by, created_at
		FROM images WHERE status = 'quarantined'
		ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var total int64
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &total, countQ)
	if err != nil {
		return nil, 0, fmt.Errorf("image repo list quarantined count: %w", err)
	}

	var images []*model.Image
	err = sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &images, listQ, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("image repo list quarantined select: %w", err)
	}

	return images, total, nil
}
