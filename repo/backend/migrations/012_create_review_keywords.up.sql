CREATE TABLE review_keywords (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    review_id BIGINT UNSIGNED NOT NULL,
    keyword VARCHAR(128) NOT NULL,
    weight DECIMAL(6,4) NOT NULL,
    INDEX idx_review_keywords_review_id (review_id),
    CONSTRAINT fk_review_keywords_review_id FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
