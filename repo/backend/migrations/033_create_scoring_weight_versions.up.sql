CREATE TABLE scoring_weight_versions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    weight_id BIGINT UNSIGNED NOT NULL,
    version INT UNSIGNED NOT NULL,
    impression_w DECIMAL(5,4) NOT NULL,
    click_w DECIMAL(5,4) NOT NULL,
    dwell_w DECIMAL(5,4) NOT NULL,
    favorite_w DECIMAL(5,4) NOT NULL,
    share_w DECIMAL(5,4) NOT NULL,
    comment_w DECIMAL(5,4) NOT NULL,
    changed_by BIGINT UNSIGNED NOT NULL,
    effective_at DATETIME(3) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE INDEX idx_scoring_weight_ver_wid_ver (weight_id, version),
    CONSTRAINT fk_scoring_weight_ver_weight FOREIGN KEY (weight_id) REFERENCES scoring_weights(id) ON DELETE CASCADE,
    CONSTRAINT fk_scoring_weight_ver_changed_by FOREIGN KEY (changed_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
