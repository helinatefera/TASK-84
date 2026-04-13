CREATE TABLE images (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sha256_hash CHAR(64) NOT NULL UNIQUE,
    original_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(32) NOT NULL,
    file_size INT UNSIGNED NOT NULL,
    storage_path VARCHAR(512) NOT NULL,
    width INT UNSIGNED,
    height INT UNSIGNED,
    status ENUM('processing','active','quarantined','deleted') NOT NULL DEFAULT 'processing',
    quarantine_reason VARCHAR(255) NULL,
    uploaded_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_images_status (status),
    INDEX idx_images_uploaded_by (uploaded_by),
    CONSTRAINT fk_images_uploaded_by FOREIGN KEY (uploaded_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
