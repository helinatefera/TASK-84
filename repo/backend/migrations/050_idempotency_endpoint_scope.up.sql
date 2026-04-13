ALTER TABLE idempotency_keys DROP FOREIGN KEY fk_idempotency_keys_user;
ALTER TABLE idempotency_keys MODIFY COLUMN user_id BIGINT UNSIGNED NULL;
DROP INDEX idx_idempotency_keys_hash_user ON idempotency_keys;
CREATE UNIQUE INDEX idx_idempotency_keys_hash ON idempotency_keys(key_hash);
