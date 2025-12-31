# Travel Documents Generator for RAG

This script generates comprehensive travel content for cities and stores them in PostgreSQL with pgvector for semantic search and RAG (Retrieval Augmented Generation) applications.

## Features

- 🤖 **AI-Powered Content Generation**: Uses OpenAI GPT-4 or Claude to generate rich travel content
- 🎯 **Multi-Category Coverage**: Historical, Logistics, Food, Activities, and Nearby Places
- 🔍 **Vector Embeddings**: Automatically generates and stores embeddings for semantic search
- 📊 **PostgreSQL + pgvector**: Leverages pgvector for efficient similarity search
- 🌍 **Initial Dataset**: Includes 8 Indian cities (Varanasi, Ayodhya, Vrindavan, Ujjain, Dwarka, Asansol, Kolkata, New Delhi)

## Prerequisites

### 1. PostgreSQL with pgvector

```bash
# Using Docker (easiest)
docker run -d \
  --name travel-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=trip \
  -p 5432:5432 \
  pgvector/pgvector:pg16

# Or install pgvector on existing PostgreSQL
# Ubuntu/Debian
sudo apt install postgresql-16-pgvector

# macOS
brew install pgvector
```

### 2. Database Setup

```bash
# Connect to your database
psql -h localhost -U postgres -d trip

# Enable pgvector extension (if not already enabled)
CREATE EXTENSION IF NOT EXISTS vector;
```

### 3. Environment Variables

Set up your `.env` file:

```env
# Database connection
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=trip

# AI Provider API Key (choose one)
OPENAI_API_KEY=sk-...           # For OpenAI GPT-4 + embeddings
# OR
ANTHROPIC_API_KEY=sk-ant-...    # For Claude (requires OPENAI_API_KEY for embeddings)
```

## Installation

```bash
# Install dependencies
go get github.com/pgvector/pgvector-go

# Build the script
go build -o generate-travel-docs
```

## Usage

### Running the Generator

```bash
# Basic usage (uses mock provider if no API key)
go run main.go providers.go

# With OpenAI
export OPENAI_API_KEY=sk-...
go run main.go providers.go

# With Claude
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-...  # Still needed for embeddings
go run main.go providers.go
```

### Running Migrations First

```bash
# Apply the database migration
psql -h localhost -U postgres -d trip < ../../migrations/20251231000000_create_travel_documents.sql
```

## Content Categories

The script generates content in 5 categories for each city:

1. **Historical & Cultural Information**
   - Historical significance
   - Cultural heritage
   - Monuments and landmarks
   - Local customs

2. **Travel Logistics**
   - Transportation options
   - Accommodation recommendations
   - Best time to visit
   - Visa/permit requirements

3. **Food & Dining**
   - Local cuisine
   - Restaurant recommendations
   - Street food
   - Must-try dishes

4. **Activities & Attractions**
   - Tourist attractions
   - Things to do
   - Shopping areas
   - Festivals and events

5. **Nearby Places of Interest**
   - Day trip destinations
   - Surrounding attractions
   - Regional highlights

## Initial Cities

The script includes these 8 Indian cities:

- **Varanasi** (Uttar Pradesh) - Ancient spiritual city
- **Ayodhya** (Uttar Pradesh) - Historic religious city
- **Vrindavan** (Uttar Pradesh) - Sacred town of Lord Krishna
- **Ujjain** (Madhya Pradesh) - Kumbh Mela site
- **Dwarka** (Gujarat) - Coastal ancient city
- **Asansol** (West Bengal) - Industrial city
- **Kolkata** (West Bengal) - Cultural capital
- **New Delhi** (Delhi) - National capital

## Database Schema

### travel_documents Table

```sql
CREATE TABLE travel_documents (
    id UUID PRIMARY KEY,
    city_name VARCHAR(255) NOT NULL,
    city_country VARCHAR(255) NOT NULL,
    document_type VARCHAR(100) NOT NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    embedding vector(1536),
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

### Indexes

- City name index for fast filtering
- Document type index for category queries
- HNSW index on embeddings for fast similarity search
- GIN index on metadata for JSON queries

## Querying the Data

### Basic Queries

```sql
-- Get all documents for a city
SELECT * FROM travel_documents WHERE city_name = 'Varanasi';

-- Get documents by type
SELECT * FROM travel_documents WHERE document_type = 'food';

-- Count documents per city
SELECT city_name, COUNT(*) FROM travel_documents GROUP BY city_name;
```

### Semantic Search

```sql
-- Find similar documents (using a pre-computed embedding)
SELECT
    city_name,
    title,
    content,
    1 - (embedding <=> '[0.1, 0.2, ...]'::vector) as similarity
FROM travel_documents
WHERE document_type = 'food'
ORDER BY embedding <=> '[0.1, 0.2, ...]'::vector
LIMIT 5;
```

## Adding More Cities

Edit `main.go` and add cities to the `GetInitialCities()` function:

```go
{
    Name:    "Mumbai",
    Country: "India",
    State:   "Maharashtra",
    Coordinates: &travelknowledge.GeoCoordinates{
        Latitude:  19.0760,
        Longitude: 72.8777,
    },
    Description: "Financial capital of India",
}
```

## Cost Estimation

### OpenAI Pricing (as of 2024)

- **GPT-4 Turbo**: ~$0.01 per 1K tokens (input) + $0.03 per 1K tokens (output)
- **text-embedding-ada-002**: ~$0.0001 per 1K tokens

Estimated cost for 8 cities:
- Content generation: ~$1-2
- Embeddings: ~$0.10
- **Total: ~$1-2.50**

### Claude Pricing

- **Claude 3.5 Sonnet**: ~$0.003 per 1K tokens (input) + $0.015 per 1K tokens (output)
- Estimated cost: ~$0.50-1.00 (still need OpenAI for embeddings)

### Mock Provider

- **Free** - for testing and development
- Generates placeholder content
- Mock embeddings (deterministic)

## Troubleshooting

### pgvector extension not found

```bash
# Check if extension is available
psql -c "SELECT * FROM pg_available_extensions WHERE name = 'vector';"

# If not installed, install it
# Ubuntu/Debian
sudo apt install postgresql-16-pgvector

# macOS
brew install pgvector
```

### API Rate Limiting

The script includes a 2-second delay between cities to avoid rate limits. If you still hit limits:

```go
// Increase delay in main.go
time.Sleep(5 * time.Second)  // Change from 2 to 5 seconds
```

### Embedding Dimension Mismatch

If using a different embedding model, update the vector dimension:

```sql
-- In migration file
embedding vector(768)  -- For smaller models like sentence-transformers

-- In models.go
Embedding pgvector.Vector `gorm:"type:vector(768)"`
```

## Next Steps

1. **Integrate with RAG Pipeline**: Use these documents for context in your trip planning LLM
2. **Add More Cities**: Expand the dataset to cover more destinations
3. **Implement Search API**: Create endpoints to query the vector database
4. **Update Content**: Schedule periodic content updates
5. **Add Images**: Store image URLs in metadata for richer content

## License

MIT

## Support

For issues or questions, refer to the main trip-planner documentation.
