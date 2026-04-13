CREATE TABLE analytics_sessions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_uuid CHAR(36) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NULL,
    item_id BIGINT UNSIGNED NULL,
    experiment_variant_id BIGINT UNSIGNED NULL,
    started_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    ended_at DATETIME(3) NULL,
    last_active_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    user_agent VARCHAR(512),
    ip_address VARCHAR(45),
    INDEX idx_analytics_sessions_user_started (user_id, started_at),
    CONSTRAINT fk_analytics_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_analytics_sessions_item FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
