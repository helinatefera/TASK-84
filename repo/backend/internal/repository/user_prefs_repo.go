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

type userPrefsRepo struct {
	db *database.DB
}

// NewUserPreferencesRepository returns a new UserPreferencesRepository backed by MySQL.
func NewUserPreferencesRepository(db *database.DB) UserPreferencesRepository {
	return &userPrefsRepo{db: db}
}

func (r *userPrefsRepo) Upsert(ctx context.Context, prefs *model.UserPreferences) error {
	const q = `
		INSERT INTO user_preferences (user_id, locale, timezone, notification_settings, created_at, updated_at)
		VALUES (:user_id, :locale, :timezone, :notification_settings, :created_at, :updated_at)
		ON DUPLICATE KEY UPDATE
			locale = VALUES(locale),
			timezone = VALUES(timezone),
			notification_settings = VALUES(notification_settings),
			updated_at = VALUES(updated_at)`

	_, err := sqlx.NamedExecContext(ctx, r.db.ExtContext(ctx), q, prefs)
	if err != nil {
		return fmt.Errorf("user prefs repo upsert: %w", err)
	}
	return nil
}

func (r *userPrefsRepo) GetByUserID(ctx context.Context, userID uint64) (*model.UserPreferences, error) {
	const q = `SELECT id, user_id, locale, timezone, notification_settings, created_at, updated_at
		FROM user_preferences WHERE user_id = ?`

	var prefs model.UserPreferences
	err := sqlx.GetContext(ctx, r.db.ExtContext(ctx), &prefs, q, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user prefs repo get by user id: %w", err)
	}
	return &prefs, nil
}
