DROP TABLE IF EXISTS link_codes;
ALTER TABLE users
  DROP COLUMN IF EXISTS telegram_id,
  DROP COLUMN IF EXISTS telegram_username;
