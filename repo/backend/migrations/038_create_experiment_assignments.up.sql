CREATE TABLE experiment_assignments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    experiment_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    variant_id BIGINT UNSIGNED NOT NULL,
    assigned_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE INDEX idx_experiment_assignments_exp_user (experiment_id, user_id),
    CONSTRAINT fk_experiment_assignments_experiment FOREIGN KEY (experiment_id) REFERENCES experiments(id),
    CONSTRAINT fk_experiment_assignments_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_experiment_assignments_variant FOREIGN KEY (variant_id) REFERENCES experiment_variants(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
