CREATE TABLE recovery_drills (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    backup_file VARCHAR(512) NOT NULL,
    started_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    completed_at DATETIME(3) NULL,
    status ENUM('running','success','failed') NOT NULL DEFAULT 'running',
    verified_tables INT UNSIGNED NOT NULL DEFAULT 0,
    error_log TEXT NULL,
    triggered_by BIGINT UNSIGNED NULL,
    INDEX idx_recovery_drills_started (started_at),
    CONSTRAINT fk_recovery_drills_triggered_by FOREIGN KEY (triggered_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
