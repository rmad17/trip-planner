#!/bin/sh
set -e

echo "Waiting for database server..."
MAX_RETRIES=30
RETRY_COUNT=0

until PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT:-5432}" -U "${DB_USER}" -d "postgres" -c '\q' 2>/dev/null; do
  RETRY_COUNT=$((RETRY_COUNT + 1))
  if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "Database not reachable after $MAX_RETRIES attempts"
    exit 1
  fi
  echo "Waiting for database... ($RETRY_COUNT/$MAX_RETRIES)"
  sleep 2
done

echo "Creating database '${DB_NAME}' if not exists..."
PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT:-5432}" -U "${DB_USER}" -d "postgres" \
  -tc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | grep -q 1 || \
  PGPASSWORD="${DB_PASSWORD}" createdb -h "${DB_HOST}" -p "${DB_PORT:-5432}" -U "${DB_USER}" "${DB_NAME}"

echo "Applying migrations..."
atlas migrate apply --url "${DB_URL}" --dir "file:///app/migrations"

echo "Starting server..."
exec ./main
