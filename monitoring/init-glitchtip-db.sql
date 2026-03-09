-- Auto-executed by PostgreSQL on first container init via
-- docker-entrypoint-initdb.d/ mount in docker-compose.yml.
--
-- Creates the glitchtip database and user for GlitchTip error tracking.
-- This script runs as the gotogether superuser.

-- Create the glitchtip role
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'glitchtip') THEN
    CREATE ROLE glitchtip WITH LOGIN PASSWORD 'glitchtip';
  END IF;
END
$$;

-- Create the glitchtip database
-- (SELECT ... \gexec doesn't work in entrypoint; use CREATE IF NOT EXISTS pattern)
CREATE DATABASE glitchtip OWNER glitchtip;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE glitchtip TO glitchtip;
