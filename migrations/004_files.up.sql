CREATE TABLE IF NOT EXISTS files (
       id STRING PRIMARY KEY DEFAULT (translate(substr(encode(gen_random_uuid()::bytes, 'base64'), 1, 10), '/+', '-_')),
       filename STRING,
       filetype STRING,
       create_ts TIMESTAMP DEFAULT NOW()::timestamp,
       counter INT8 DEFAULT (INT '0')
       );
