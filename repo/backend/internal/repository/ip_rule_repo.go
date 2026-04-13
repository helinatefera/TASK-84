package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/localinsights/portal/internal/pkg/database"
)

type ipRuleRepo struct {
	db *database.DB
}

// NewIPRuleRepository returns a new IPRuleRepository backed by MySQL.
func NewIPRuleRepository(db *database.DB) IPRuleRepository {
	return &ipRuleRepo{db: db}
}

func (r *ipRuleRepo) Create(ctx context.Context, cidr string, ruleType string, description string, createdBy uint64) error {
	const q = `INSERT INTO ip_rules (cidr, rule_type, description, created_by, created_at)
		VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, cidr, ruleType, description, createdBy, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("ip rule repo create: %w", err)
	}
	return nil
}

func (r *ipRuleRepo) Delete(ctx context.Context, id uint64) error {
	const q = `DELETE FROM ip_rules WHERE id = ?`

	_, err := r.db.ExtContext(ctx).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("ip rule repo delete: %w", err)
	}
	return nil
}

func (r *ipRuleRepo) ListAll(ctx context.Context) ([]struct {
	CIDR     string
	RuleType string
}, error) {
	const q = `SELECT cidr, rule_type FROM ip_rules ORDER BY id ASC`

	var rows []struct {
		CIDR     string `db:"cidr"`
		RuleType string `db:"rule_type"`
	}
	err := sqlx.SelectContext(ctx, r.db.ExtContext(ctx), &rows, q)
	if err != nil {
		return nil, fmt.Errorf("ip rule repo list all: %w", err)
	}

	// Convert from db-tagged struct to interface struct.
	result := make([]struct {
		CIDR     string
		RuleType string
	}, len(rows))
	for i, row := range rows {
		result[i].CIDR = row.CIDR
		result[i].RuleType = row.RuleType
	}

	return result, nil
}
