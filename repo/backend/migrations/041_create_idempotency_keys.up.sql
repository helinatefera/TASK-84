CREATE TABLE idempotency_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    key_hash VARCHAR(128) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    response_code SMALLINT UNSIGNED NULL,
    response_body MEDIUMTEXT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    expires_at DATETIME(3) NOT NULL,
    UNIQUE INDEX idx_idempotency_keys_hash_user (key_hash, user_id),
    INDEX idx_idempotency_keys_expires (expires_at),
    CONSTRAINT fk_idempotency_keys_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
