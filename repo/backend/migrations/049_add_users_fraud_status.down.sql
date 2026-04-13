DROP INDEX idx_users_fraud_status ON users;
ALTER TABLE users DROP COLUMN fraud_status;
