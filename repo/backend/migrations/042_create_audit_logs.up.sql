CREATE TABLE audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    actor_id BIGINT UNSIGNED NULL,
    actor_role VARCHAR(32),
    action VARCHAR(128) NOT NULL,
    target_type VARCHAR(64) NULL,
    target_id BIGINT UNSIGNED NULL,
    ip_address VARCHAR(45),
    request_id VARCHAR(64),
    details JSON NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_audit_logs_actor_created (actor_id, created_at DESC),
    INDEX idx_audit_logs_action_created (action, created_at DESC),
    INDEX idx_audit_logs_target (target_type, target_id),
    INDEX idx_audit_logs_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
