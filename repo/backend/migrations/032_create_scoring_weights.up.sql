CREATE TABLE scoring_weights (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    impression_w DECIMAL(5,4) NOT NULL DEFAULT 0.0500,
    click_w DECIMAL(5,4) NOT NULL DEFAULT 0.2000,
    dwell_w DECIMAL(5,4) NOT NULL DEFAULT 0.3000,
    favorite_w DECIMAL(5,4) NOT NULL DEFAULT 0.2500,
    share_w DECIMAL(5,4) NOT NULL DEFAULT 0.1000,
    comment_w DECIMAL(5,4) NOT NULL DEFAULT 0.1000,
    is_active TINYINT(1) NOT NULL DEFAULT 0,
    version INT UNSIGNED NOT NULL DEFAULT 1,
    updated_by BIGINT UNSIGNED NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    CONSTRAINT fk_scoring_weights_updated_by FOREIGN KEY (updated_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO scoring_weights (name, is_active) VALUES ('default', 1);
