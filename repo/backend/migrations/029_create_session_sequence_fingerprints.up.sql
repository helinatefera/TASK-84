CREATE TABLE session_sequence_fingerprints (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    session_id BIGINT UNSIGNED NOT NULL,
    sequence_hash CHAR(64) NOT NULL,
    event_count INT UNSIGNED NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_session_seq_fp_user_hash (user_id, sequence_hash),
    UNIQUE INDEX idx_session_seq_fp_session (session_id),
    CONSTRAINT fk_session_seq_fp_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_session_seq_fp_session FOREIGN KEY (session_id) REFERENCES analytics_sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
