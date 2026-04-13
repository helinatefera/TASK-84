CREATE TABLE login_attempts (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    ip_address   VARCHAR(45)  NOT NULL,
    email        VARCHAR(255) NOT NULL,
    attempted_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    success      TINYINT(1)   NOT NULL DEFAULT 0,

    INDEX idx_login_attempts_ip_attempted   (ip_address, attempted_at),
    INDEX idx_login_attempts_email_attempted (email, attempted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
