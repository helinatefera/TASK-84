CREATE TABLE job_state (
    job_name VARCHAR(64) PRIMARY KEY,
    last_run_at DATETIME(3) NOT NULL,
    watermark VARCHAR(255) NULL,
    status ENUM('idle','running','failed') NOT NULL DEFAULT 'idle',
    last_error TEXT NULL,
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
