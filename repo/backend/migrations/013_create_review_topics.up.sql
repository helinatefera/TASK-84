CREATE TABLE review_topics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    review_id BIGINT UNSIGNED NOT NULL,
    topic VARCHAR(128) NOT NULL,
    confidence DECIMAL(4,3) NOT NULL,
    INDEX idx_review_topics_review_id (review_id),
    CONSTRAINT fk_review_topics_review_id FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
