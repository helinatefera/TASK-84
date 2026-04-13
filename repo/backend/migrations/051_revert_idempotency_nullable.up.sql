UPDATE idempotency_keys SET user_id = 0 WHERE user_id IS NULL;
ALTER TABLE idempotency_keys MODIFY COLUMN user_id BIGINT UNSIGNED NOT NULL;
