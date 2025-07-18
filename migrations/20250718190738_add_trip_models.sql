-- Rename a column from "place_name" to "name"
ALTER TABLE "trip_plans" RENAME COLUMN "place_name" TO "name";
-- Modify "trip_plans" table
ALTER TABLE "trip_plans" DROP COLUMN "place_id";
-- Create "stays" table
CREATE TABLE "stays" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "google_location" text NULL,
  "mapbox_location" text NULL,
  "stay_type" text NULL,
  "stay_notes" text NULL,
  "start_date" timestamptz NULL,
  "end_date" timestamptz NULL,
  "is_prepaid" boolean NULL,
  "payment_mode" text NULL,
  "trip_hop" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "trip_hops" table
CREATE TABLE "trip_hops" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "name" text NULL,
  "map_source" text NULL,
  "place_id" text NULL,
  "start_date" timestamptz NULL,
  "end_date" timestamptz NULL,
  "notes" text NULL,
  "po_is" text[] NULL,
  "previous_hop" text NULL,
  "next_hop" text NULL,
  "trip_plan" uuid NOT NULL,
  PRIMARY KEY ("id")
);
