-- Add nullable JSONB columns to "trip_hops" so AI-orchestrated trips
-- can persist the user-selected hotel/flight offers and the route
-- to the next hop alongside the existing structured fields.
ALTER TABLE "trip_hops"
  ADD COLUMN IF NOT EXISTS "selected_flight" jsonb NULL,
  ADD COLUMN IF NOT EXISTS "selected_hotel"  jsonb NULL,
  ADD COLUMN IF NOT EXISTS "route_to_next"   jsonb NULL;
