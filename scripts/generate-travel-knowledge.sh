#!/bin/bash
# Travel Knowledge Generator Script
# Generates and stores travel content with vector embeddings for RAG

set -e

echo "🌍 Travel Knowledge Generator for RAG"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}❌ Error: .env file not found${NC}"
    echo "Please create a .env file with your database and API credentials"
    exit 1
fi

# Load environment variables
set -a
source .env
set +a

echo -e "${BLUE}📋 Configuration:${NC}"
echo "  Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"
echo "  User: ${DB_USER}"

# Check if pgvector extension is enabled
echo ""
echo -e "${BLUE}🔍 Checking pgvector extension...${NC}"

PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_NAME}" -t -c \
    "SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'vector');" | grep -q 't'

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ pgvector extension is enabled${NC}"
else
    echo -e "${YELLOW}⚠️  pgvector extension not found. Attempting to enable...${NC}"
    PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_NAME}" -c \
        "CREATE EXTENSION IF NOT EXISTS vector;"

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ pgvector extension enabled successfully${NC}"
    else
        echo -e "${RED}❌ Failed to enable pgvector extension${NC}"
        echo "Please run manually: CREATE EXTENSION vector;"
        exit 1
    fi
fi

# Check if migration has been run
echo ""
echo -e "${BLUE}🔍 Checking database schema...${NC}"

PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_NAME}" -t -c \
    "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'travel_documents');" | grep -q 't'

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ travel_documents table exists${NC}"
else
    echo -e "${YELLOW}⚠️  travel_documents table not found. Running migration...${NC}"

    MIGRATION_FILE="migrations/20251231000000_create_travel_documents.sql"
    if [ -f "$MIGRATION_FILE" ]; then
        PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_NAME}" < "$MIGRATION_FILE"
        echo -e "${GREEN}✅ Migration completed${NC}"
    else
        echo -e "${RED}❌ Migration file not found: $MIGRATION_FILE${NC}"
        exit 1
    fi
fi

# Check API keys
echo ""
echo -e "${BLUE}🔑 Checking API credentials...${NC}"

if [ -n "$OPENAI_API_KEY" ]; then
    echo -e "${GREEN}✅ OpenAI API key found${NC}"
    PROVIDER="OpenAI"
elif [ -n "$ANTHROPIC_API_KEY" ]; then
    echo -e "${GREEN}✅ Anthropic API key found${NC}"
    PROVIDER="Claude"
    echo -e "${YELLOW}⚠️  Note: OpenAI API key also required for embeddings${NC}"
else
    echo -e "${YELLOW}⚠️  No API keys found - using mock provider${NC}"
    PROVIDER="Mock"
fi

# Ask for confirmation
echo ""
echo -e "${BLUE}📊 Ready to generate content:${NC}"
echo "  Provider: $PROVIDER"
echo "  Cities: 8 (Varanasi, Ayodhya, Vrindavan, Ujjain, Dwarka, Asansol, Kolkata, New Delhi)"
echo "  Categories: 5 per city (Historical, Logistics, Food, Activities, Nearby Places)"
echo ""

if [ "$PROVIDER" != "Mock" ]; then
    echo -e "${YELLOW}💰 Cost estimate: ~\$1-2 for 8 cities${NC}"
    echo ""
fi

read -p "Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Build and run the generator
echo ""
echo -e "${BLUE}🔨 Building generator...${NC}"

cd cmd/generate-travel-docs
go build -o ../../bin/generate-travel-docs

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Build successful${NC}"
echo ""
echo -e "${BLUE}🚀 Running generator...${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd ../..
./bin/generate-travel-docs

EXIT_CODE=$?

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}🎉 Content generation completed successfully!${NC}"
    echo ""
    echo -e "${BLUE}📈 Database statistics:${NC}"

    # Query database for statistics
    PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_NAME}" << EOF
\x
SELECT
    COUNT(*) as total_documents,
    COUNT(DISTINCT city_name) as cities,
    COUNT(DISTINCT document_type) as document_types
FROM travel_documents;

\x off

SELECT
    document_type,
    COUNT(*) as count
FROM travel_documents
GROUP BY document_type
ORDER BY document_type;
EOF

    echo ""
    echo -e "${GREEN}✅ All done! Your RAG knowledge base is ready.${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Test semantic search with your vector embeddings"
    echo "  2. Integrate with your trip planning RAG pipeline"
    echo "  3. Add more cities as needed"
else
    echo -e "${RED}❌ Content generation failed with exit code $EXIT_CODE${NC}"
    exit $EXIT_CODE
fi
