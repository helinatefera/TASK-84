CREATE TABLE notifications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    template_key VARCHAR(64) NOT NULL,
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    rendered_subject VARCHAR(255) NOT NULL,
    rendered_body TEXT NOT NULL,
    data JSON,
    is_read TINYINT(1) NOT NULL DEFAULT 0,
    read_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_notifications_user_read_created (user_id, is_read, created_at DESC),
    CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
