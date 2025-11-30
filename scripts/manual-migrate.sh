#!/bin/bash
# Manual migration runner for local development
# Run this outside of Docker

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}🗄️  Running database migrations locally...${NC}"

# Load environment variables
if [ -f .env ]; then
    source .env
else
    echo -e "${RED}❌ .env file not found!${NC}"
    exit 1
fi

# Parse DATABASE_URL or use individual DB_ variables
if [ -n "$DATABASE_URL" ]; then
    DB_URL=$DATABASE_URL
elif [ -n "$DB_URL" ]; then
    DB_URL=$DB_URL
else
    # Construct from individual variables
    DB_URL="postgres://${DB_USER:-postgres}:${DB_PASSWORD:-postgres}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-trip_planner}?sslmode=disable"
fi

echo -e "${YELLOW}Database URL: ${DB_URL}${NC}"

# Ask which migration method
echo ""
echo "Which migration method do you want to use?"
echo "1) SQL migrations (migrations/*.sql)"
echo "2) GORM AutoMigrate (Go models)"
read -p "Enter choice [1-2]: " choice

case $choice in
    1)
        echo -e "${YELLOW}📝 Running SQL migrations...${NC}"

        # Extract connection details from URL
        DB_USER=$(echo $DB_URL | sed -n 's/.*:\/\/\([^:]*\):.*/\1/p')
        DB_PASS=$(echo $DB_URL | sed -n 's/.*:\/\/[^:]*:\([^@]*\)@.*/\1/p')
        DB_HOST=$(echo $DB_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
        DB_PORT=$(echo $DB_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
        DB_NAME=$(echo $DB_URL | sed -n 's/.*\/\([^?]*\).*/\1/p')

        # Run each migration file
        for migration in migrations/*.sql; do
            if [ -f "$migration" ]; then
                echo -e "  → Applying $(basename $migration)..."
                PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration" 2>&1 | grep -v "NOTICE: \|already exists" || true
            fi
        done

        echo -e "${GREEN}✅ SQL migrations completed!${NC}"
        ;;

    2)
        echo -e "${YELLOW}📝 Running GORM AutoMigrate...${NC}"
        go run cmd/migrate/main.go
        echo -e "${GREEN}✅ GORM migrations completed!${NC}"
        ;;

    *)
        echo -e "${RED}❌ Invalid choice${NC}"
        exit 1
        ;;
esac

echo -e "${GREEN}🎉 Migration completed successfully!${NC}"
