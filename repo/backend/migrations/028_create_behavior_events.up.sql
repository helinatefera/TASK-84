CREATE TABLE behavior_events (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    event_type ENUM('impression','click','dwell','favorite','share','comment') NOT NULL,
    item_id BIGINT UNSIGNED NULL,
    dwell_seconds SMALLINT UNSIGNED NULL,
    event_data JSON NULL,
    client_ts DATETIME(3) NOT NULL,
    server_ts DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    dedup_hash VARCHAR(64) NULL,
    INDEX idx_behavior_events_session (session_id),
    INDEX idx_behavior_events_item_type_ts (item_id, event_type, server_ts),
    INDEX idx_behavior_events_user_ts (user_id, server_ts),
    UNIQUE INDEX idx_behavior_events_dedup (dedup_hash),
    CONSTRAINT fk_behavior_events_session FOREIGN KEY (session_id) REFERENCES analytics_sessions(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
