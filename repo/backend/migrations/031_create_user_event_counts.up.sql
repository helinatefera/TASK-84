CREATE TABLE user_event_counts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    hour_bucket DATETIME NOT NULL,
    event_count INT UNSIGNED NOT NULL DEFAULT 0,
    UNIQUE INDEX idx_user_event_counts_user_hour (user_id, hour_bucket),
    CONSTRAINT fk_user_event_counts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
