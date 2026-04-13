CREATE TABLE experiment_variants (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    experiment_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(64) NOT NULL,
    traffic_pct DECIMAL(5,2) NOT NULL,
    config JSON NOT NULL,
    UNIQUE INDEX idx_experiment_variants_exp_name (experiment_id, name),
    CONSTRAINT fk_experiment_variants_experiment FOREIGN KEY (experiment_id) REFERENCES experiments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
