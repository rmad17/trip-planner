#!/bin/sh
# Migration runner script
# This script runs inside the Docker container

set -e

echo "🗄️  Running database migrations..."

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

until PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST:-db}" -U "${DB_USER}" -d "${DB_NAME}" -c '\q' 2>/dev/null; do
  RETRY_COUNT=$((RETRY_COUNT + 1))
  if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "❌ Database is not ready after $MAX_RETRIES attempts"
    exit 1
  fi
  echo "Database is unavailable - sleeping (attempt $RETRY_COUNT/$MAX_RETRIES)"
  sleep 2
done

echo "✅ Database is ready!"

# Choose migration method based on environment variable
MIGRATION_METHOD="${MIGRATION_METHOD:-sql}"

if [ "$MIGRATION_METHOD" = "sql" ]; then
    echo "📝 Running SQL migrations from migrations/ directory..."

    # Run all SQL migrations in order
    for migration in /app/migrations/*.sql; do
        if [ -f "$migration" ]; then
            echo "  → Applying $(basename $migration)..."
            PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST:-db}" -U "${DB_USER}" -d "${DB_NAME}" -f "$migration" 2>&1 | grep -v "NOTICE: \|already exists"
        fi
    done

    echo "✅ SQL migrations completed!"

elif [ "$MIGRATION_METHOD" = "gorm" ]; then
    echo "📝 Running GORM AutoMigrate..."

    # Run the Go migration command
    /app/main migrate

    echo "✅ GORM migrations completed!"

else
    echo "❌ Unknown migration method: $MIGRATION_METHOD"
    echo "   Valid options: sql, gorm"
    exit 1
fi

echo "🎉 All migrations completed successfully!"
