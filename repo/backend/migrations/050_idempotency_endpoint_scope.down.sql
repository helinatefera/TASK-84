DROP INDEX idx_idempotency_keys_hash ON idempotency_keys;
CREATE UNIQUE INDEX idx_idempotency_keys_hash_user ON idempotency_keys(key_hash, user_id);
ALTER TABLE idempotency_keys MODIFY COLUMN user_id BIGINT UNSIGNED NOT NULL;
ALTER TABLE idempotency_keys ADD CONSTRAINT fk_idempotency_keys_user FOREIGN KEY (user_id) REFERENCES users(id);
