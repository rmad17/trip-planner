# Database Migration Deployment Guide

## Overview

This guide explains how database migrations are automatically run as part of the deployment process for the Trip Planner application.

## Migration Strategy

The application supports **two migration methods**:

1. **SQL Migrations** (Recommended for production)
   - Uses raw SQL files in `migrations/` directory
   - Provides full control over database changes
   - Idempotent (can be run multiple times safely)
   - Managed via Atlas or direct SQL execution

2. **GORM AutoMigrate** (Good for development)
   - Automatically generates migrations from Go models
   - Faster for rapid development
   - Less control over exact SQL

## How It Works

### Automated Migration Flow

When you run `./deploy.sh`, the following happens:

```
1. Stop existing containers
2. Build new containers
3. Run migration service (waits for DB to be ready)
   → Executes all SQL migrations in order
   → Idempotent - safe to run multiple times
4. Start API service (only starts if migrations succeed)
5. Start other services
```

### Migration Service

The `migration` service in `docker-compose.yml`:

- Runs **once** before the API starts
- Uses the same Docker image as the API
- Waits for database to be healthy
- Executes `/app/scripts/run-migrations.sh`
- API only starts if migrations complete successfully

## Configuration

### Environment Variables

Set in your `.env` file:

```bash
# Database connection (required)
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=trip_planner
DB_HOST=db  # For Docker, 'db' is the service name

# Migration method (optional)
MIGRATION_METHOD=sql  # Options: sql, gorm (default: sql)
```

### Migration Method Selection

**Option 1: SQL Migrations (Recommended)**

Set in `.env` or `docker-compose.yml`:
```bash
MIGRATION_METHOD=sql
```

This runs all `.sql` files in `migrations/` directory in alphabetical order.

**Option 2: GORM AutoMigrate**

Set in `.env` or `docker-compose.yml`:
```bash
MIGRATION_METHOD=gorm
```

This runs the Go migration binary that calls `database.AutoMigrateAll()`.

## Files Modified

### 1. Dockerfile

**Changes:**
- Fixed build path: Changed from `./cmd` to `app.go` (bug fix)
- Added `postgresql-client` for running SQL migrations
- Built separate `migrate` binary for GORM migrations
- Copied `scripts/run-migrations.sh` for migration execution
- Made migration scripts executable

**Key sections:**
```dockerfile
# Build both main app and migration utility
RUN go build -o main app.go
RUN go build -o migrate ./cmd/migrate

# Install postgresql-client for SQL migrations
RUN apk add postgresql-client

# Copy migration files and scripts
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/scripts/run-migrations.sh ./scripts/
```

### 2. docker-compose.yml

**Added migration service:**
```yaml
migration:
  build:
    context: .
    dockerfile: Dockerfile
  container_name: trip-planner-migration
  restart: "no"  # Only runs once
  environment:
    - DB_HOST=db
    - DB_USER=${DB_USER}
    - DB_PASSWORD=${DB_PASSWORD}
    - DB_NAME=${DB_NAME}
    - MIGRATION_METHOD=sql
  depends_on:
    db:
      condition: service_healthy
  command: ["/app/scripts/run-migrations.sh"]
```

**Updated API service:**
```yaml
api:
  depends_on:
    db:
      condition: service_healthy
    migration:
      condition: service_completed_successfully  # Wait for migrations
```

### 3. deploy.sh

**Added migration step:**
```bash
echo "🗄️  Running database migrations..."
docker-compose up migration

echo "🚀 Starting containers..."
docker-compose up -d
```

Migrations run automatically during deployment!

### 4. scripts/run-migrations.sh

**New file** - Migration runner script that:
- Waits for database to be ready (with retries)
- Executes SQL or GORM migrations based on `MIGRATION_METHOD`
- Provides clear logging and error messages
- Handles idempotent SQL execution

## Creating New Migrations

### Method 1: SQL Migrations (Recommended)

1. **Create migration file** with timestamp:
   ```bash
   # Format: YYYYMMDDHHMMSS_description.sql
   touch migrations/20251130120000_add_user_field.sql
   ```

2. **Write idempotent SQL**:
   ```sql
   -- Add column if it doesn't exist
   ALTER TABLE users
     ADD COLUMN IF NOT EXISTS new_field VARCHAR(255);

   -- Create index if it doesn't exist
   CREATE INDEX IF NOT EXISTS idx_users_new_field
     ON users(new_field);
   ```

3. **Test locally**:
   ```bash
   docker-compose up migration
   ```

4. **Deploy**:
   ```bash
   ./deploy.sh  # Migrations run automatically
   ```

### Method 2: GORM AutoMigrate

1. **Update your model** in `accounts/models.go`, `trips/models.go`, etc.

2. **Update** `database/models.go` if adding new models:
   ```go
   func AutoMigrateAll() error {
       models := []interface{}{
           &accounts.User{},
           &trips.TripPlan{},
           &yourpackage.NewModel{},  // Add new model
       }
       return core.DB.AutoMigrate(models...)
   }
   ```

3. **Set migration method** to GORM:
   ```bash
   # In .env or docker-compose.yml
   MIGRATION_METHOD=gorm
   ```

4. **Deploy**:
   ```bash
   ./deploy.sh
   ```

## Manual Migration Commands

### Run migrations manually inside container:

```bash
# SQL migrations
docker-compose exec api /app/scripts/run-migrations.sh

# GORM migrations
docker-compose exec api /app/migrate
```

### Run migrations on local database:

```bash
# SQL migrations
for f in migrations/*.sql; do
  psql -U postgres -d trip_planner -f "$f"
done

# GORM migrations
go run cmd/migrate/main.go
```

### Run specific migration file:

```bash
psql -U postgres -d trip_planner -f migrations/20251130100619_add_google_oauth_fields.sql
```

## Deployment Process

### Full Deployment to Digital Ocean

1. **Upload code** to droplet:
   ```bash
   ./upload-to-droplet.sh
   ```

2. **SSH into droplet**:
   ```bash
   ssh root@your-droplet-ip
   cd /opt/trip-planner
   ```

3. **Ensure .env is configured**:
   ```bash
   # Verify these are set
   cat .env | grep DB_
   cat .env | grep MIGRATION_METHOD
   ```

4. **Run deployment** (migrations run automatically):
   ```bash
   ./deploy.sh
   ```

### Watch Migration Progress

```bash
# View migration logs in real-time
docker-compose logs -f migration

# Check migration status
docker-compose ps migration

# View migration output after completion
docker logs trip-planner-migration
```

## Troubleshooting

### Migration Fails

**Check migration logs:**
```bash
docker logs trip-planner-migration
```

**Common issues:**
1. **Database not ready**: Migration script waits 30 retries, increase if needed
2. **SQL syntax error**: Check your SQL file syntax
3. **Permission denied**: Ensure database user has ALTER/CREATE permissions
4. **Constraint violation**: Review migration order or add IF NOT EXISTS

**Retry migrations:**
```bash
# Remove failed migration container
docker-compose rm -f migration

# Run migrations again
docker-compose up migration
```

### API Won't Start After Migration

The API depends on migrations completing successfully:

```bash
# Check migration exit status
docker-compose ps migration
# Should show "Exit 0"

# If migration failed (Exit 1), check logs
docker logs trip-planner-migration

# Fix migration and retry
docker-compose up migration

# Then start API
docker-compose up -d api
```

### Rollback a Migration

**For SQL migrations:**

Create a rollback migration:
```bash
# migrations/20251130130000_rollback_user_field.sql
ALTER TABLE users DROP COLUMN IF EXISTS new_field;
```

Then deploy as normal.

**For GORM migrations:**

GORM doesn't support automatic rollbacks. You'll need to:
1. Manually write SQL to reverse changes
2. Or restore database from backup

### Database Connection Issues

```bash
# Test database connection
docker-compose exec db psql -U $DB_USER -d $DB_NAME -c "SELECT 1;"

# Check database is healthy
docker-compose ps db

# View database logs
docker-compose logs db
```

## Best Practices

1. **Always use idempotent SQL**
   - Use `IF NOT EXISTS` for CREATE statements
   - Use `ADD COLUMN IF NOT EXISTS` for ALTER TABLE
   - Check existence before DROP

2. **Test migrations locally first**
   ```bash
   docker-compose up migration
   ```

3. **One migration per feature**
   - Don't mix unrelated changes
   - Use descriptive filenames

4. **Never edit old migrations**
   - Create new migration to fix issues
   - Old migrations may have already run in production

5. **Backup before major changes**
   ```bash
   docker-compose exec db pg_dump -U $DB_USER $DB_NAME > backup.sql
   ```

6. **Review SQL before deployment**
   - Check for locking issues on large tables
   - Consider adding indexes in separate migration
   - Test with production-like data volume

## Migration Tracking

### View Applied Migrations

**For SQL migrations:**
```sql
-- Check migration history (if you implement a migrations table)
SELECT * FROM schema_migrations ORDER BY applied_at DESC;
```

**For GORM migrations:**
GORM tracks migrations internally. You can check:
```bash
docker-compose exec api /app/migrate
# GORM will show "Migration completed successfully!" if no changes needed
```

### Current Migration Status

```bash
# List all migration files
ls -la migrations/

# Check last migration applied (if using Atlas)
cat migrations/atlas.sum

# Verify database schema
docker-compose exec db psql -U $DB_USER -d $DB_NAME -c "\dt"
```

## Advanced: Atlas Integration

If you want to use Atlas for more sophisticated migrations:

1. **Install Atlas** in Dockerfile:
   ```dockerfile
   RUN apk add --no-cache curl && \
       curl -sSf https://atlasgo.sh | sh
   ```

2. **Update run-migrations.sh** to use Atlas:
   ```bash
   atlas migrate apply \
     --url "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:5432/$DB_NAME?sslmode=disable" \
     --dir "file:///app/migrations"
   ```

3. **Generate migrations** from models:
   ```bash
   atlas migrate diff \
     --env local \
     --to "file://schema.hcl"
   ```

## Summary

- ✅ Migrations run **automatically** on every deployment
- ✅ API **waits** for migrations to complete
- ✅ Migrations are **idempotent** - safe to run multiple times
- ✅ **Two methods** supported: SQL and GORM
- ✅ **Easy rollback** via new migrations
- ✅ **Comprehensive logging** for debugging

Your deployment pipeline now includes automated database migrations!
