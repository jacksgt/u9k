CREATE TABLE IF NOT EXISTS files (
       id TEXT PRIMARY KEY DEFAULT (translate(substr(encode(gen_random_bytes(6), 'base64'), 1, 6), '/+', '-_')),
       -- For CockroachDB:
       -- id STRING PRIMARY KEY DEFAULT (translate(substr(encode(gen_random_uuid()::bytes, 'base64'), 1, 6), '/+', '-_')),
       filename TEXT,
       filetype TEXT,
       create_ts TIMESTAMP DEFAULT NOW()::timestamp,
       counter INT8 DEFAULT (INT '0'),
       expire INTERVAL
       );
