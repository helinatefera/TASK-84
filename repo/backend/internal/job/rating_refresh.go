package job

import (
	"context"
	"log/slog"
	"time"

	"github.com/localinsights/portal/internal/pkg/database"
)

type RatingRefreshJob struct {
	db *database.DB
}

func NewRatingRefreshJob(db *database.DB) *RatingRefreshJob {
	return &RatingRefreshJob{db: db}
}

func (j *RatingRefreshJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Find items with reviews updated since last refresh
	rows, err := j.db.QueryxContext(ctx, `
		SELECT DISTINCT r.item_id
		FROM reviews r
		JOIN item_rating_aggregates a ON a.item_id = r.item_id
		WHERE r.updated_at > a.last_refreshed
		UNION
		SELECT DISTINCT r.item_id
		FROM reviews r
		LEFT JOIN item_rating_aggregates a ON a.item_id = r.item_id
		WHERE a.item_id IS NULL
		LIMIT 100
	`)
	if err != nil {
		slog.Error("rating_refresh: failed to find items", "error", err)
		return
	}
	defer rows.Close()

	var itemIDs []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		itemIDs = append(itemIDs, id)
	}

	for _, itemID := range itemIDs {
		j.refreshItem(ctx, itemID)
	}

	if len(itemIDs) > 0 {
		slog.Info("rating_refresh: refreshed items", "count", len(itemIDs))
	}
}

func (j *RatingRefreshJob) refreshItem(ctx context.Context, itemID uint64) {
	row := j.db.QueryRowxContext(ctx, `
		SELECT
			COALESCE(AVG(rating), 0) as avg_rating,
			COUNT(*) as rating_count,
			COALESCE(SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END), 0) as rating_1,
			COALESCE(SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END), 0) as rating_2,
			COALESCE(SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END), 0) as rating_3,
			COALESCE(SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END), 0) as rating_4,
			COALESCE(SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END), 0) as rating_5
		FROM reviews
		WHERE item_id = ? AND fraud_status NOT IN ('suspected_fraud', 'confirmed_fraud')
	`, itemID)

	var avgRating float64
	var count, r1, r2, r3, r4, r5 int
	if err := row.Scan(&avgRating, &count, &r1, &r2, &r3, &r4, &r5); err != nil {
		slog.Error("rating_refresh: failed to compute aggregate", "item_id", itemID, "error", err)
		return
	}

	_, err := j.db.ExecContext(ctx, `
		INSERT INTO item_rating_aggregates (item_id, avg_rating, rating_count, rating_1, rating_2, rating_3, rating_4, rating_5, last_refreshed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(3))
		ON DUPLICATE KEY UPDATE
			avg_rating = VALUES(avg_rating),
			rating_count = VALUES(rating_count),
			rating_1 = VALUES(rating_1),
			rating_2 = VALUES(rating_2),
			rating_3 = VALUES(rating_3),
			rating_4 = VALUES(rating_4),
			rating_5 = VALUES(rating_5),
			last_refreshed = NOW(3)
	`, itemID, avgRating, count, r1, r2, r3, r4, r5)

	if err != nil {
		slog.Error("rating_refresh: failed to upsert aggregate", "item_id", itemID, "error", err)
	}
}
