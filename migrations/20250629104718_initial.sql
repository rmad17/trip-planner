-- Create "trip_plans" table
CREATE TABLE "trip_plans" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "place_name" text NULL,
  "place_id" text NULL,
  "start_date" timestamptz NULL,
  "end_date" timestamptz NULL,
  "min_days" smallint NULL,
  "travel_mode" text NULL,
  "notes" text NULL,
  "hotels" text[] NULL,
  "tags" text[] NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "username" text NULL,
  "password" text NULL,
  "email" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_users_username" UNIQUE ("username")
);
