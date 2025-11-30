-- Add Google OAuth fields to users table
ALTER TABLE "users"
  ADD COLUMN IF NOT EXISTS "google_id" character varying UNIQUE,
  ADD COLUMN IF NOT EXISTS "name" character varying,
  ADD COLUMN IF NOT EXISTS "first_name" character varying,
  ADD COLUMN IF NOT EXISTS "last_name" character varying,
  ADD COLUMN IF NOT EXISTS "avatar_url" character varying,
  ADD COLUMN IF NOT EXISTS "locale" character varying,
  ADD COLUMN IF NOT EXISTS "provider" character varying,
  ADD COLUMN IF NOT EXISTS "access_token" text,
  ADD COLUMN IF NOT EXISTS "refresh_token" text,
  ADD COLUMN IF NOT EXISTS "expires_at" bigint;

-- Add unique constraint on email if it doesn't exist
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'users_email_key'
    AND conrelid = 'users'::regclass
  ) THEN
    ALTER TABLE "users" ADD CONSTRAINT "users_email_key" UNIQUE ("email");
  END IF;
END $$;

-- Create index on google_id for faster lookups
CREATE INDEX IF NOT EXISTS "idx_users_google_id" ON "users" ("google_id");

-- Create index on provider for filtering
CREATE INDEX IF NOT EXISTS "idx_users_provider" ON "users" ("provider");
