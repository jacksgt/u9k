CREATE EXTENSION pgcrypto;
ALTER TABLE links ALTER COLUMN id
      SET DEFAULT (translate(substr(encode(gen_random_bytes(6), 'base64'), 1, 6), '/+', '-_'));
-- For CockroachDB:
-- ALTER TABLE links ALTER COLUMN id
--       SET DEFAULT (translate(substr(encode(gen_random_uuid()::bytes, 'base64'), 1, 6), '/+', '-_'));
