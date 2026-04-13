ALTER TABLE appeals MODIFY COLUMN status ENUM('pending','accepted','rejected','needs_edit') NOT NULL DEFAULT 'pending';
