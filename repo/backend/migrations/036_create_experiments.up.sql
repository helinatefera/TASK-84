CREATE TABLE experiments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid CHAR(36) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    status ENUM('draft','running','paused','completed','rolled_back') NOT NULL DEFAULT 'draft',
    hash_salt VARCHAR(64) NOT NULL,
    min_sample_size INT UNSIGNED NOT NULL DEFAULT 100,
    created_by BIGINT UNSIGNED NOT NULL,
    started_at DATETIME(3) NULL,
    ended_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX idx_experiments_status (status),
    CONSTRAINT fk_experiments_created_by FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
