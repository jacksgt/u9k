ALTER TABLE files ADD COLUMN filesize INT8;
ALTER TABLE files ADD COLUMN emails_sent SMALLINT DEFAULT 0;