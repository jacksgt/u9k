-- CREATE DATABASE IF NOT EXISTS u9k;
CREATE TABLE IF NOT EXISTS links (
       id TEXT PRIMARY KEY,
       url TEXT,
       create_ts TIMESTAMP DEFAULT NOW()::timestamp
       );
