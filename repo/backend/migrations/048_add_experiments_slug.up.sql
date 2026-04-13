ALTER TABLE experiments ADD COLUMN slug VARCHAR(128) NULL AFTER name;
CREATE UNIQUE INDEX idx_experiments_slug ON experiments(slug);
