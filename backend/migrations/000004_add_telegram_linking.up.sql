ALTER TABLE users
  ADD COLUMN telegram_id BIGINT UNIQUE,
  ADD COLUMN telegram_username TEXT;

CREATE TABLE link_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  code TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_link_codes_email_code ON link_codes(email, code);
