-- Modify "trip_days" table
ALTER TABLE "trip_days" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "trip_plans" table
ALTER TABLE "trip_plans" ALTER COLUMN "id" DROP DEFAULT, DROP COLUMN "travel_mode", ADD COLUMN "travel_modes" text[] NULL;
-- Modify "documents" table
ALTER TABLE "documents" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "tags" TYPE text;
-- Modify "trip_hops" table
ALTER TABLE "trip_hops" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "expense_splits" table
ALTER TABLE "expense_splits" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "expenses" table
ALTER TABLE "expenses" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "tags" TYPE text;
-- Modify "stays" table
ALTER TABLE "stays" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "travellers" table
ALTER TABLE "travellers" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "document_shares" table
ALTER TABLE "document_shares" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "activities" table
ALTER TABLE "activities" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "expense_settlements" table
ALTER TABLE "expense_settlements" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "user_preferences" table
ALTER TABLE "user_preferences" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "users" table
ALTER TABLE "users" ALTER COLUMN "id" DROP DEFAULT;
-- Create "notification_templates" table
CREATE TABLE "notification_templates" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "type" character varying(50) NOT NULL,
  "channel" character varying(50) NOT NULL,
  "subject" text NULL,
  "content" text NOT NULL,
  "content_html" text NULL,
  "default_priority" character varying(20) NULL DEFAULT 'normal',
  "default_metadata" text NULL,
  "variables" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "version" bigint NULL DEFAULT 1,
  "created_by" uuid NULL,
  "updated_by" uuid NULL,
  "locale" character varying(10) NULL DEFAULT 'en',
  PRIMARY KEY ("id")
);
-- Create index "idx_notification_templates_channel" to table: "notification_templates"
CREATE INDEX "idx_notification_templates_channel" ON "notification_templates" ("channel");
-- Create index "idx_notification_templates_is_active" to table: "notification_templates"
CREATE INDEX "idx_notification_templates_is_active" ON "notification_templates" ("is_active");
-- Create index "idx_notification_templates_locale" to table: "notification_templates"
CREATE INDEX "idx_notification_templates_locale" ON "notification_templates" ("locale");
-- Create index "idx_notification_templates_name" to table: "notification_templates"
CREATE UNIQUE INDEX "idx_notification_templates_name" ON "notification_templates" ("name");
-- Create index "idx_notification_templates_type" to table: "notification_templates"
CREATE INDEX "idx_notification_templates_type" ON "notification_templates" ("type");
-- Create "notification_batches" table
CREATE TABLE "notification_batches" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "channel" character varying(50) NOT NULL,
  "type" character varying(50) NOT NULL,
  "priority" character varying(20) NULL DEFAULT 'normal',
  "template_id" uuid NULL,
  "total_count" bigint NULL DEFAULT 0,
  "sent_count" bigint NULL DEFAULT 0,
  "delivered_count" bigint NULL DEFAULT 0,
  "failed_count" bigint NULL DEFAULT 0,
  "status" character varying(50) NULL DEFAULT 'draft',
  "scheduled_at" timestamptz NULL,
  "started_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "created_by" uuid NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notification_batches_channel" to table: "notification_batches"
CREATE INDEX "idx_notification_batches_channel" ON "notification_batches" ("channel");
-- Create index "idx_notification_batches_scheduled_at" to table: "notification_batches"
CREATE INDEX "idx_notification_batches_scheduled_at" ON "notification_batches" ("scheduled_at");
-- Create index "idx_notification_batches_status" to table: "notification_batches"
CREATE INDEX "idx_notification_batches_status" ON "notification_batches" ("status");
-- Create index "idx_notification_batches_type" to table: "notification_batches"
CREATE INDEX "idx_notification_batches_type" ON "notification_batches" ("type");
-- Create "notification_preferences" table
CREATE TABLE "notification_preferences" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  "channel" character varying(50) NOT NULL,
  "type" character varying(50) NOT NULL,
  "is_enabled" boolean NULL DEFAULT true,
  "priority" character varying(20) NULL,
  "quiet_hours_start" timestamptz NULL,
  "quiet_hours_end" timestamptz NULL,
  "timezone" character varying(50) NULL,
  "max_per_day" bigint NULL,
  "max_per_week" bigint NULL,
  "max_per_month" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notification_preferences_user_id" to table: "notification_preferences"
CREATE INDEX "idx_notification_preferences_user_id" ON "notification_preferences" ("user_id");
-- Create "notification_providers" table
CREATE TABLE "notification_providers" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "name" character varying(100) NOT NULL,
  "channel" character varying(50) NOT NULL,
  "provider" character varying(100) NOT NULL,
  "description" text NULL,
  "config" jsonb NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "is_default" boolean NULL DEFAULT false,
  "priority" bigint NULL DEFAULT 0,
  "rate_limit" bigint NULL DEFAULT 0,
  "current_usage" bigint NULL DEFAULT 0,
  "usage_reset_at" timestamptz NULL,
  "health_status" character varying(50) NULL DEFAULT 'healthy',
  "last_health_check" timestamptz NULL,
  "last_error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notification_providers_channel" to table: "notification_providers"
CREATE INDEX "idx_notification_providers_channel" ON "notification_providers" ("channel");
-- Create index "idx_notification_providers_is_active" to table: "notification_providers"
CREATE INDEX "idx_notification_providers_is_active" ON "notification_providers" ("is_active");
-- Create index "idx_notification_providers_name" to table: "notification_providers"
CREATE UNIQUE INDEX "idx_notification_providers_name" ON "notification_providers" ("name");
-- Create "notifications" table
CREATE TABLE "notifications" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "type" character varying(50) NOT NULL,
  "channel" character varying(50) NOT NULL,
  "priority" character varying(20) NULL DEFAULT 'normal',
  "status" character varying(50) NULL DEFAULT 'pending',
  "sender_id" uuid NULL,
  "recipient_id" uuid NULL,
  "subject" text NULL,
  "content" text NOT NULL,
  "content_html" text NULL,
  "template_id" uuid NULL,
  "template_data" text NULL,
  "channel_provider" character varying(100) NULL,
  "channel_data" text NULL,
  "recipient_email" character varying(255) NULL,
  "recipient_phone" character varying(50) NULL,
  "recipient_device_id" character varying(255) NULL,
  "recipient_webhook" character varying(500) NULL,
  "scheduled_at" timestamptz NULL,
  "sent_at" timestamptz NULL,
  "delivered_at" timestamptz NULL,
  "read_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "retry_count" bigint NULL DEFAULT 0,
  "max_retries" bigint NULL DEFAULT 3,
  "next_retry_at" timestamptz NULL,
  "last_error" text NULL,
  "metadata" text NULL,
  "entity_type" character varying(100) NULL,
  "entity_id" uuid NULL,
  "tags" text NULL,
  "external_id" character varying(255) NULL,
  "external_response" text NULL,
  "expires_at" timestamptz NULL,
  "archived_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notification_templates_notifications" FOREIGN KEY ("template_id") REFERENCES "notification_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notifications_channel" to table: "notifications"
CREATE INDEX "idx_notifications_channel" ON "notifications" ("channel");
-- Create index "idx_notifications_delivered_at" to table: "notifications"
CREATE INDEX "idx_notifications_delivered_at" ON "notifications" ("delivered_at");
-- Create index "idx_notifications_entity_id" to table: "notifications"
CREATE INDEX "idx_notifications_entity_id" ON "notifications" ("entity_id");
-- Create index "idx_notifications_entity_type" to table: "notifications"
CREATE INDEX "idx_notifications_entity_type" ON "notifications" ("entity_type");
-- Create index "idx_notifications_expires_at" to table: "notifications"
CREATE INDEX "idx_notifications_expires_at" ON "notifications" ("expires_at");
-- Create index "idx_notifications_external_id" to table: "notifications"
CREATE INDEX "idx_notifications_external_id" ON "notifications" ("external_id");
-- Create index "idx_notifications_next_retry_at" to table: "notifications"
CREATE INDEX "idx_notifications_next_retry_at" ON "notifications" ("next_retry_at");
-- Create index "idx_notifications_priority" to table: "notifications"
CREATE INDEX "idx_notifications_priority" ON "notifications" ("priority");
-- Create index "idx_notifications_recipient_device_id" to table: "notifications"
CREATE INDEX "idx_notifications_recipient_device_id" ON "notifications" ("recipient_device_id");
-- Create index "idx_notifications_recipient_email" to table: "notifications"
CREATE INDEX "idx_notifications_recipient_email" ON "notifications" ("recipient_email");
-- Create index "idx_notifications_recipient_id" to table: "notifications"
CREATE INDEX "idx_notifications_recipient_id" ON "notifications" ("recipient_id");
-- Create index "idx_notifications_recipient_phone" to table: "notifications"
CREATE INDEX "idx_notifications_recipient_phone" ON "notifications" ("recipient_phone");
-- Create index "idx_notifications_scheduled_at" to table: "notifications"
CREATE INDEX "idx_notifications_scheduled_at" ON "notifications" ("scheduled_at");
-- Create index "idx_notifications_sender_id" to table: "notifications"
CREATE INDEX "idx_notifications_sender_id" ON "notifications" ("sender_id");
-- Create index "idx_notifications_sent_at" to table: "notifications"
CREATE INDEX "idx_notifications_sent_at" ON "notifications" ("sent_at");
-- Create index "idx_notifications_status" to table: "notifications"
CREATE INDEX "idx_notifications_status" ON "notifications" ("status");
-- Create index "idx_notifications_template_id" to table: "notifications"
CREATE INDEX "idx_notifications_template_id" ON "notifications" ("template_id");
-- Create index "idx_notifications_type" to table: "notifications"
CREATE INDEX "idx_notifications_type" ON "notifications" ("type");
-- Create "notification_audits" table
CREATE TABLE "notification_audits" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "notification_id" uuid NOT NULL,
  "status" character varying(50) NOT NULL,
  "event" character varying(100) NOT NULL,
  "message" text NULL,
  "details" text NULL,
  "timestamp" timestamptz NOT NULL,
  "request_data" text NULL,
  "response_data" text NULL,
  "provider" character varying(100) NULL,
  "provider_response" text NULL,
  "is_error" boolean NULL DEFAULT false,
  "error_code" character varying(100) NULL,
  "error_message" text NULL,
  "actor_id" uuid NULL,
  "actor_type" character varying(50) NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notifications_audits" FOREIGN KEY ("notification_id") REFERENCES "notifications" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notification_audits_is_error" to table: "notification_audits"
CREATE INDEX "idx_notification_audits_is_error" ON "notification_audits" ("is_error");
-- Create index "idx_notification_audits_notification_id" to table: "notification_audits"
CREATE INDEX "idx_notification_audits_notification_id" ON "notification_audits" ("notification_id");
-- Create index "idx_notification_audits_timestamp" to table: "notification_audits"
CREATE INDEX "idx_notification_audits_timestamp" ON "notification_audits" ("timestamp");
