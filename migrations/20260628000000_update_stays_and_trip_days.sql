-- Add accommodation fields to stays table
ALTER TABLE "stays" ADD COLUMN IF NOT EXISTS "name" TEXT;
ALTER TABLE "stays" ADD COLUMN IF NOT EXISTS "address" TEXT;
ALTER TABLE "stays" ADD COLUMN IF NOT EXISTS "cost" NUMERIC;
ALTER TABLE "stays" ADD COLUMN IF NOT EXISTS "cost_per_night" NUMERIC;

-- Make trip_days.date nullable (date is now optional for day ordering without a specific date)
ALTER TABLE "trip_days" ALTER COLUMN "date" DROP NOT NULL;
ALTER TABLE "trip_days" ALTER COLUMN "date" SET DEFAULT NULL;
