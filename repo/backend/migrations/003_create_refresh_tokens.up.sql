CREATE TABLE refresh_tokens (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT UNSIGNED NOT NULL,
    token_hash VARCHAR(128)    NOT NULL UNIQUE,
    expires_at DATETIME(3)     NOT NULL,
    revoked    TINYINT(1)      NOT NULL DEFAULT 0,
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    INDEX idx_refresh_tokens_user_id    (user_id),
    INDEX idx_refresh_tokens_expires_at (expires_at),

    CONSTRAINT fk_refresh_tokens_user_id
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
