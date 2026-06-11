-- Phase 2 & 3: soft delete, lifecycle, sharing, transport, contacts, checklist
-- Phase 6 prerequisite: account hygiene (password reset, email verification)
-- Run `atlas migrate hash` after applying this file.

-- Soft delete: add deleted_at to core trip content tables
ALTER TABLE "trip_plans"   ADD COLUMN IF NOT EXISTS "deleted_at" timestamptz NULL;
ALTER TABLE "trip_hops"    ADD COLUMN IF NOT EXISTS "deleted_at" timestamptz NULL;
ALTER TABLE "trip_days"    ADD COLUMN IF NOT EXISTS "deleted_at" timestamptz NULL;
ALTER TABLE "activities"   ADD COLUMN IF NOT EXISTS "deleted_at" timestamptz NULL;
ALTER TABLE "travellers"   ADD COLUMN IF NOT EXISTS "deleted_at" timestamptz NULL;
ALTER TABLE "stays"        ADD COLUMN IF NOT EXISTS "deleted_at" timestamptz NULL;
ALTER TABLE "documents"    ADD COLUMN IF NOT EXISTS "deleted_at" timestamptz NULL;

CREATE INDEX IF NOT EXISTS "idx_trip_plans_deleted_at"  ON "trip_plans"("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_trip_hops_deleted_at"   ON "trip_hops"("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_trip_days_deleted_at"   ON "trip_days"("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_activities_deleted_at"  ON "activities"("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_travellers_deleted_at"  ON "travellers"("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_stays_deleted_at"       ON "stays"("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_documents_deleted_at"   ON "documents"("deleted_at");

-- Trip plan lifecycle fields
ALTER TABLE "trip_plans"
  ADD COLUMN IF NOT EXISTS "timezone"        text NULL DEFAULT 'Asia/Kolkata',
  ADD COLUMN IF NOT EXISTS "user_set_status" boolean NULL DEFAULT false;

-- Typed status: existing 'status' column is text NULL — no DDL change needed,
-- validation happens in application code (TripStatus type).

-- Unique partial index on share_code (NULLs excluded so multiple NULLs are allowed)
CREATE UNIQUE INDEX IF NOT EXISTS "idx_trip_plans_share_code"
  ON "trip_plans"("share_code") WHERE "share_code" IS NOT NULL;

-- Transport segments
CREATE TABLE IF NOT EXISTS "transport_segments" (
  "id"           uuid         NOT NULL DEFAULT gen_random_uuid(),
  "created_at"   timestamptz  NULL,
  "updated_at"   timestamptz  NULL,
  "deleted_at"   timestamptz  NULL,
  "trip_plan"    uuid         NOT NULL,
  "from_hop_id"  uuid         NULL,
  "to_hop_id"    uuid         NULL,
  "mode"         varchar(20)  NOT NULL,
  "operator"     text         NULL,
  "booking_ref"  text         NULL,
  "depart_at"    timestamptz  NULL,
  "arrive_at"    timestamptz  NULL,
  "depart_from"  text         NULL,
  "arrive_to"    text         NULL,
  "depart_tz"    text         NULL,
  "arrive_tz"    text         NULL,
  "cost"         numeric      NULL,
  "currency"     varchar(3)   NULL,
  "seats"        text         NULL,
  "deep_link_url" text        NULL,
  "notes"        text         NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_transport_trip_plan"
    FOREIGN KEY ("trip_plan") REFERENCES "trip_plans"("id") ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS "idx_transport_segments_deleted_at" ON "transport_segments"("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_transport_segments_trip_plan"  ON "transport_segments"("trip_plan");
CREATE INDEX IF NOT EXISTS "idx_transport_segments_depart_at"  ON "transport_segments"("depart_at");

-- Trip contacts
CREATE TABLE IF NOT EXISTS "trip_contacts" (
  "id"         uuid        NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "trip_plan"  uuid        NOT NULL,
  "hop_id"     uuid        NULL,
  "name"       text        NOT NULL,
  "role"       varchar(20) NOT NULL,
  "phone"      text        NULL,
  "email"      text        NULL,
  "address"    text        NULL,
  "notes"      text        NULL,
  "is_active"  boolean     NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_contact_trip_plan"
    FOREIGN KEY ("trip_plan") REFERENCES "trip_plans"("id") ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS "idx_trip_contacts_deleted_at" ON "trip_contacts"("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_trip_contacts_trip_plan"  ON "trip_contacts"("trip_plan");

-- Checklist items
CREATE TABLE IF NOT EXISTS "checklist_items" (
  "id"           uuid        NOT NULL DEFAULT gen_random_uuid(),
  "created_at"   timestamptz NULL,
  "updated_at"   timestamptz NULL,
  "trip_plan"    uuid        NOT NULL,
  "hop_id"       uuid        NULL,
  "title"        text        NOT NULL,
  "category"     varchar(20) NOT NULL DEFAULT 'other',
  "due_date"     timestamptz NULL,
  "assignee_id"  uuid        NULL,
  "is_completed" boolean     NULL DEFAULT false,
  "completed_at" timestamptz NULL,
  "notes"        text        NULL,
  "sort_order"   bigint      NULL DEFAULT 0,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_checklist_trip_plan"
    FOREIGN KEY ("trip_plan") REFERENCES "trip_plans"("id") ON DELETE NO ACTION
);
CREATE INDEX IF NOT EXISTS "idx_checklist_items_trip_plan"  ON "checklist_items"("trip_plan");
CREATE INDEX IF NOT EXISTS "idx_checklist_items_due_date"   ON "checklist_items"("due_date") WHERE "due_date" IS NOT NULL;

-- Account hygiene: password reset and email verification
ALTER TABLE "users"
  ADD COLUMN IF NOT EXISTS "email_verified"           boolean     NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS "email_verification_token" text        NULL,
  ADD COLUMN IF NOT EXISTS "password_reset_token"     text        NULL,
  ADD COLUMN IF NOT EXISTS "password_reset_expiry"    timestamptz NULL;

CREATE INDEX IF NOT EXISTS "idx_users_password_reset_token"
  ON "users"("password_reset_token") WHERE "password_reset_token" IS NOT NULL;
CREATE INDEX IF NOT EXISTS "idx_users_email_verification_token"
  ON "users"("email_verification_token") WHERE "email_verification_token" IS NOT NULL;
