CREATE TABLE share_links (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token CHAR(64) NOT NULL UNIQUE,
    saved_view_id BIGINT UNSIGNED NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    expires_at DATETIME(3) NOT NULL,
    is_revoked TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_share_links_expires (expires_at),
    CONSTRAINT fk_share_links_view FOREIGN KEY (saved_view_id) REFERENCES saved_views(id) ON DELETE CASCADE,
    CONSTRAINT fk_share_links_created_by FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
