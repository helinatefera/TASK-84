CREATE TABLE sensitive_word_rule_versions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_id BIGINT UNSIGNED NOT NULL,
    version INT UNSIGNED NOT NULL,
    pattern VARCHAR(255) NOT NULL,
    action ENUM('block','flag','replace') NOT NULL,
    replacement VARCHAR(255) NULL,
    changed_by BIGINT UNSIGNED NOT NULL,
    changed_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE INDEX idx_rule_versions_rule_version (rule_id, version),
    CONSTRAINT fk_rule_versions_rule FOREIGN KEY (rule_id) REFERENCES sensitive_word_rules(id) ON DELETE CASCADE,
    CONSTRAINT fk_rule_versions_changed_by FOREIGN KEY (changed_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
