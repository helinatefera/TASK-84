ALTER TABLE appeals MODIFY COLUMN status ENUM('pending','accepted','rejected') NOT NULL DEFAULT 'pending';
