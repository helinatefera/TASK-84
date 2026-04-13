package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/config"
)

type DB struct {
	*sqlx.DB
}

func NewMySQL(cfg config.DatabaseConfig) (*DB, error) {
	db, err := sqlx.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

type txKey struct{}

func (d *DB) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := d.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original: %w)", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}

func (d *DB) TxFromContext(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txKey{}).(*sqlx.Tx)
	return tx
}

func (d *DB) ExtContext(ctx context.Context) sqlx.ExtContext {
	if tx := d.TxFromContext(ctx); tx != nil {
		return tx
	}
	return d.DB
}

func (d *DB) HealthCheck(ctx context.Context) error {
	return d.PingContext(ctx)
}

func (d *DB) Stats() sql.DBStats {
	return d.DB.Stats()
}
