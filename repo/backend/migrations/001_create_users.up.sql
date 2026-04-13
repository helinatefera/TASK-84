CREATE TABLE users (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid       CHAR(36)      NOT NULL UNIQUE,
    username   VARCHAR(64)   NOT NULL UNIQUE,
    email      VARCHAR(255)  NOT NULL UNIQUE,
    password_hash VARCHAR(512) NOT NULL,
    role       ENUM('admin','moderator','product_analyst','regular_user') NOT NULL DEFAULT 'regular_user',
    is_active  TINYINT(1)    NOT NULL DEFAULT 1,
    created_at DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    INDEX idx_users_email (email),
    INDEX idx_users_role  (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
