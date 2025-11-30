# Database Migration Quick Reference

## TL;DR - Just Deploy

```bash
./deploy.sh
```

**Migrations run automatically!** ✨

---

## Common Commands

### Create New Migration

```bash
# SQL migration (recommended)
touch migrations/$(date +%Y%m%d%H%M%S)_add_feature.sql

# Edit the file and write idempotent SQL
nano migrations/20251130*.sql
```

### Run Migrations Locally

```bash
# Interactive menu
./scripts/manual-migrate.sh

# Or use Docker
docker-compose up migration

# Or run Go migration directly
go run cmd/migrate/main.go
```

### Deploy with Migrations

```bash
# Upload to droplet
./upload-to-droplet.sh

# SSH and deploy
ssh root@your-droplet-ip
cd /opt/trip-planner
./deploy.sh  # Migrations run automatically
```

### View Migration Logs

```bash
# During deployment
docker-compose logs -f migration

# After deployment
docker logs trip-planner-migration
```

### Manual Migration in Production

```bash
# Run migrations manually
docker-compose up migration

# Or exec into container
docker-compose exec api /app/scripts/run-migrations.sh
```

---

## Migration Template

### SQL Migration (Idempotent)

```sql
-- migrations/20251130120000_add_field.sql

-- Add column if it doesn't exist
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS new_field VARCHAR(255);

-- Create index if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_users_new_field
  ON users(new_field);

-- Update existing rows (safe)
UPDATE users SET new_field = 'default'
  WHERE new_field IS NULL;
```

### GORM Migration

```go
// Update models in accounts/models.go, trips/models.go, etc.
type User struct {
    // ... existing fields
    NewField string `json:"new_field"`
}

// Set MIGRATION_METHOD=gorm in .env
// Run: go run cmd/migrate/main.go
```

---

## Troubleshooting

### Migration Failed?

```bash
# Check logs
docker logs trip-planner-migration

# Retry
docker-compose rm -f migration
docker-compose up migration
```

### API Won't Start?

```bash
# Check migration status
docker-compose ps migration
# Should show "Exit 0"

# View migration logs
docker logs trip-planner-migration

# Fix and retry
docker-compose up migration
docker-compose up -d api
```

### Rollback Migration?

Create a new rollback migration:
```bash
touch migrations/$(date +%Y%m%d%H%M%S)_rollback_feature.sql
# Write SQL to reverse changes
```

---

## Configuration

### .env Settings

```bash
# Required for migrations
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=trip_planner

# Optional - choose migration method
MIGRATION_METHOD=sql  # or 'gorm'
```

### docker-compose.yml

```yaml
migration:
  environment:
    - MIGRATION_METHOD=sql  # Change to 'gorm' if needed
```

---

## File Locations

- **Migrations**: `migrations/*.sql`
- **Migration Script**: `scripts/run-migrations.sh`
- **Manual Runner**: `scripts/manual-migrate.sh`
- **GORM Migrator**: `cmd/migrate/main.go`
- **Deployment**: `deploy.sh` (runs migrations automatically)

---

## Full Documentation

See `DEPLOYMENT_MIGRATIONS.md` for complete details.
