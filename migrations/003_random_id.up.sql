ALTER TABLE u9k.links ALTER COLUMN id SET DEFAULT (translate(substr(encode(gen_random_uuid()::bytes, 'base64'), 1, 6), '/+', '-_'));
