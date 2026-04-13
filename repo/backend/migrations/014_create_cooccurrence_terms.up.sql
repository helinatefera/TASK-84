CREATE TABLE cooccurrence_terms (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    term_a VARCHAR(128) NOT NULL,
    term_b VARCHAR(128) NOT NULL,
    item_id BIGINT UNSIGNED NOT NULL,
    frequency INT UNSIGNED NOT NULL DEFAULT 0,
    period_start DATE NOT NULL,
    UNIQUE INDEX uidx_cooccurrence_terms_term_a_term_b_item_id_period_start (term_a, term_b, item_id, period_start),
    CONSTRAINT fk_cooccurrence_terms_item_id FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
