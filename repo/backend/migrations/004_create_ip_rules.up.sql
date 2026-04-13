CREATE TABLE ip_rules (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    cidr        VARCHAR(45)  NOT NULL,
    rule_type   ENUM('allow','deny') NOT NULL,
    description VARCHAR(255),
    created_by  BIGINT UNSIGNED NOT NULL,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    UNIQUE INDEX uq_ip_rules_cidr_rule_type (cidr, rule_type),

    CONSTRAINT fk_ip_rules_created_by
        FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
