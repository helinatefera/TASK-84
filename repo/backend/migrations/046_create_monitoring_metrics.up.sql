CREATE TABLE monitoring_metrics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    metric_name VARCHAR(128) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    tags JSON NULL,
    recorded_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_monitoring_metrics_name_recorded (metric_name, recorded_at),
    INDEX idx_monitoring_metrics_recorded (recorded_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
