CREATE TABLE analytics_aggregates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    item_id BIGINT UNSIGNED NOT NULL,
    period_start DATE NOT NULL,
    impressions INT UNSIGNED NOT NULL DEFAULT 0,
    clicks INT UNSIGNED NOT NULL DEFAULT 0,
    avg_dwell_secs DECIMAL(7,2) NOT NULL DEFAULT 0.00,
    favorites INT UNSIGNED NOT NULL DEFAULT 0,
    shares INT UNSIGNED NOT NULL DEFAULT 0,
    comments INT UNSIGNED NOT NULL DEFAULT 0,
    computed_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE INDEX idx_analytics_agg_item_period (item_id, period_start),
    CONSTRAINT fk_analytics_agg_item FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
