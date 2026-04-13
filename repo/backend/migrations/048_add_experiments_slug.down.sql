DROP INDEX idx_experiments_slug ON experiments;
ALTER TABLE experiments DROP COLUMN slug;
