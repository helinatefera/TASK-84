CREATE TABLE sensitive_word_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    pattern VARCHAR(255) NOT NULL,
    action ENUM('block','flag','replace') NOT NULL DEFAULT 'flag',
    replacement VARCHAR(255) NULL,
    version INT UNSIGNED NOT NULL DEFAULT 1,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX idx_sensitive_word_rules_is_active (is_active),
    CONSTRAINT fk_sensitive_word_rules_created_by FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
