ALTER TABLE appeals MODIFY COLUMN status ENUM('pending','approved','rejected','needs_edit') NOT NULL DEFAULT 'pending';
UPDATE appeals SET status = 'approved' WHERE status = 'accepted';
