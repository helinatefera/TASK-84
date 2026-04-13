ALTER TABLE users ADD COLUMN fraud_status ENUM('clean','suspected','confirmed') NOT NULL DEFAULT 'clean' AFTER is_active;
CREATE INDEX idx_users_fraud_status ON users(fraud_status);
