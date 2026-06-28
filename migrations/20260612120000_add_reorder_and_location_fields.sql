-- Add direct location/booking URL to stays for map/booking links
ALTER TABLE "stays" ADD COLUMN IF NOT EXISTS "location_url" text NULL;

-- Add sort_order to activities for manual drag-and-drop reordering
ALTER TABLE "activities" ADD COLUMN IF NOT EXISTS "sort_order" bigint NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS "idx_activities_trip_day_sort" ON "activities"("trip_day", "sort_order");
