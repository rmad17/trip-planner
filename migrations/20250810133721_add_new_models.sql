-- Modify "trip_plans" table
ALTER TABLE "trip_plans" ADD COLUMN "description" text NULL, ADD COLUMN "max_days" smallint NULL, ADD COLUMN "trip_type" text NULL, ADD COLUMN "budget" numeric NULL, ADD COLUMN "actual_spent" numeric NULL, ADD COLUMN "currency" character varying(10) NULL DEFAULT 'USD', ADD COLUMN "status" text NULL, ADD COLUMN "is_public" boolean NULL, ADD COLUMN "share_code" text NULL, ADD COLUMN "participants" text[] NULL;
-- Create "document_shares" table
CREATE TABLE "document_shares" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "document_id" uuid NOT NULL,
  "shared_with" uuid NOT NULL,
  "shared_by" uuid NOT NULL,
  "permission" character varying(20) NULL DEFAULT 'view',
  "expires_at" timestamptz NULL,
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create "documents" table
CREATE TABLE "documents" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "name" text NOT NULL,
  "original_name" text NOT NULL,
  "storage_provider" character varying(50) NOT NULL,
  "storage_path" text NOT NULL,
  "file_size" bigint NULL,
  "content_type" text NULL,
  "category" character varying(50) NOT NULL,
  "description" text NULL,
  "notes" text NULL,
  "tags" text[] NULL,
  "entity_type" text NULL,
  "entity_id" text NULL,
  "user_id" uuid NOT NULL,
  "uploaded_at" timestamptz NOT NULL,
  "expires_at" timestamptz NULL,
  "is_public" boolean NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create "expense_settlements" table
CREATE TABLE "expense_settlements" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "trip_plan" uuid NOT NULL,
  "from_traveller" uuid NOT NULL,
  "to_traveller" uuid NOT NULL,
  "amount" numeric NOT NULL,
  "currency" character varying(10) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "settled_at" timestamptz NULL,
  "payment_method" text NULL,
  "notes" text NULL,
  PRIMARY KEY ("id")
);
-- Modify "trip_hops" table
ALTER TABLE "trip_hops" ALTER COLUMN "previous_hop" TYPE uuid, ALTER COLUMN "next_hop" TYPE uuid, ADD COLUMN "description" text NULL, ADD COLUMN "city" text NULL, ADD COLUMN "country" text NULL, ADD COLUMN "region" text NULL, ADD COLUMN "latitude" numeric NULL, ADD COLUMN "longitude" numeric NULL, ADD COLUMN "estimated_budget" numeric NULL, ADD COLUMN "actual_spent" numeric NULL, ADD COLUMN "transportation" text NULL, ADD COLUMN "restaurants" text[] NULL, ADD COLUMN "activities" text[] NULL, ADD COLUMN "hop_order" bigint NULL, ADD CONSTRAINT "fk_trip_plans_trip_hops" FOREIGN KEY ("trip_plan") REFERENCES "trip_plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create "trip_days" table
CREATE TABLE "trip_days" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "date" timestamptz NOT NULL,
  "day_number" bigint NOT NULL,
  "title" text NULL,
  "day_type" character varying(20) NOT NULL,
  "notes" text NULL,
  "start_location" text NULL,
  "end_location" text NULL,
  "estimated_budget" numeric NULL,
  "actual_budget" numeric NULL,
  "weather" text NULL,
  "trip_plan" uuid NOT NULL,
  "from_trip_hop" uuid NULL,
  "to_trip_hop" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_trip_hops_trip_days" FOREIGN KEY ("from_trip_hop") REFERENCES "trip_hops" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_trip_plans_trip_days" FOREIGN KEY ("trip_plan") REFERENCES "trip_plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "activities" table
CREATE TABLE "activities" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "activity_type" character varying(20) NOT NULL,
  "start_time" timestamptz NULL,
  "end_time" timestamptz NULL,
  "duration" bigint NULL,
  "location" text NULL,
  "map_source" text NULL,
  "place_id" text NULL,
  "estimated_cost" numeric NULL,
  "actual_cost" numeric NULL,
  "priority" smallint NULL,
  "status" text NULL,
  "booking_ref" text NULL,
  "contact_info" text NULL,
  "notes" text NULL,
  "tags" text[] NULL,
  "trip_day" uuid NOT NULL,
  "trip_hop" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_trip_days_activities" FOREIGN KEY ("trip_day") REFERENCES "trip_days" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "expenses" table
CREATE TABLE "expenses" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "amount" numeric NOT NULL,
  "currency" character varying(10) NOT NULL,
  "category" character varying(30) NOT NULL,
  "other_category" text NULL,
  "date" timestamptz NOT NULL,
  "location" text NULL,
  "vendor" text NULL,
  "payment_method" character varying(20) NOT NULL,
  "split_method" character varying(20) NOT NULL DEFAULT 'equal',
  "receipt_url" text NULL,
  "notes" text NULL,
  "tags" text[] NULL,
  "is_recurring" boolean NULL DEFAULT false,
  "trip_plan" uuid NULL,
  "trip_hop" uuid NULL,
  "trip_day" uuid NULL,
  "activity" uuid NULL,
  "paid_by" uuid NOT NULL,
  "created_by" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "expense_splits" table
CREATE TABLE "expense_splits" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "expense" uuid NOT NULL,
  "traveller" uuid NOT NULL,
  "amount" numeric NOT NULL,
  "percentage" numeric NULL,
  "shares" bigint NULL,
  "is_paid" boolean NULL DEFAULT false,
  "paid_at" timestamptz NULL,
  "notes" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_expenses_expense_splits" FOREIGN KEY ("expense") REFERENCES "expenses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Modify "stays" table
ALTER TABLE "stays" ADD CONSTRAINT "fk_trip_hops_stays" FOREIGN KEY ("trip_hop") REFERENCES "trip_hops" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create "travellers" table
CREATE TABLE "travellers" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "first_name" text NOT NULL,
  "last_name" text NOT NULL,
  "email" text NULL,
  "phone" text NULL,
  "date_of_birth" timestamptz NULL,
  "nationality" text NULL,
  "passport_number" text NULL,
  "passport_expiry" timestamptz NULL,
  "emergency_contact" text NULL,
  "dietary_restrictions" text NULL,
  "medical_notes" text NULL,
  "role" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "joined_at" timestamptz NOT NULL,
  "notes" text NULL,
  "trip_plan" uuid NOT NULL,
  "user_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_trip_plans_travellers" FOREIGN KEY ("trip_plan") REFERENCES "trip_plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "user_preferences" table
CREATE TABLE "user_preferences" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  "map_provider" character varying(20) NULL DEFAULT 'google',
  "default_storage_prov" character varying(50) NULL DEFAULT 'digitalocean',
  "language" character varying(10) NULL DEFAULT 'en',
  "timezone" text NULL DEFAULT 'UTC',
  "currency" character varying(3) NULL DEFAULT 'USD',
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_user_preferences_user_id" UNIQUE ("user_id"),
  CONSTRAINT "fk_users_preferences" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
