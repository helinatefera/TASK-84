CREATE TABLE experiment_status_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    experiment_id BIGINT UNSIGNED NOT NULL,
    old_status ENUM('draft','running','paused','completed','rolled_back') NOT NULL,
    new_status ENUM('draft','running','paused','completed','rolled_back') NOT NULL,
    changed_by BIGINT UNSIGNED NOT NULL,
    changed_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    reason TEXT NULL,
    INDEX idx_experiment_status_hist_exp_changed (experiment_id, changed_at DESC),
    CONSTRAINT fk_experiment_status_hist_experiment FOREIGN KEY (experiment_id) REFERENCES experiments(id) ON DELETE CASCADE,
    CONSTRAINT fk_experiment_status_hist_changed_by FOREIGN KEY (changed_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
