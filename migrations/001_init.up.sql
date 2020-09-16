-- CREATE DATABASE IF NOT EXISTS u9k;
CREATE TABLE IF NOT EXISTS links (
       id STRING PRIMARY KEY,
       url STRING,
       create_ts TIMESTAMP DEFAULT NOW()::timestamp
       );
