CREATE TABLE IF NOT EXISTS email_list (
       address STRING PRIMARY KEY, -- UNIQUE
       subscribe_link UUID DEFAULT gen_random_uuid(),
       unsubscribed BOOL DEFAULT FALSE
       );
CREATE INDEX IF NOT EXISTS email_list__subscribe_link ON email_list (subscribe_link);
