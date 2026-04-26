# Travel Documents Generator for RAG

This script generates comprehensive travel content for cities and stores them in PostgreSQL with pgvector for semantic search and RAG (Retrieval Augmented Generation) applications.

## Features

- 🤖 **AI-Powered Content Generation**: Uses OpenAI GPT-4, Claude, or Ollama (local) to generate rich travel content
- 🔍 **Vector Embeddings**: Automatically generates and stores embeddings for semantic search
- 🦙 **Ollama Support**: Fully local AI with embedding generation (no API costs!)
- 🎯 **Multi-Category Coverage**: Historical, Logistics, Food, Activities, and Nearby Places
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

### 3. AI Provider Setup

#### Option A: Ollama (Local, Free, No API Key Required) 🦙

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull the models you need
ollama pull llama3.1:8b           # For content generation
ollama pull nomic-embed-text      # For embeddings

# Verify models are installed
ollama list
```

#### Option B: OpenAI

```bash
export OPENAI_API_KEY=sk-...
```

#### Option C: Claude (requires OpenAI for embeddings)

```bash
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-...  # Still needed for embeddings
```

### 4. Environment Variables

Set up your `.env` file:

```env
# Database connection
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=trip

# Default: Ollama (local, free) - runs automatically if no API keys are set
# Optional: Set models explicitly (defaults shown below)
# OLLAMA_MODEL=llama3.1:8b               # Or llama2, mistral, etc.
# OLLAMA_EMBEDDING_MODEL=nomic-embed-text # Or mxbai-embed-large, all-minilm

# Alternative: Use Claude (requires API credits)
# ANTHROPIC_API_KEY=sk-ant-...
# OPENAI_API_KEY=sk-...  # Still needed for embeddings

# Alternative: Use OpenAI (requires API credits)
# OPENAI_API_KEY=sk-...
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
# Default: Uses Ollama automatically (Local, Free!)
# Just make sure Ollama is installed with required models
ollama pull llama3.1:8b
ollama pull nomic-embed-text
go run main.go providers.go

# Or explicitly set Ollama models (optional)
export OLLAMA_MODEL=llama3.1:8b
export OLLAMA_EMBEDDING_MODEL=nomic-embed-text
go run main.go providers.go

# Override with Claude (requires API credits)
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-...  # Still needed for embeddings
go run main.go providers.go

# Override with OpenAI (requires API credits)
export OPENAI_API_KEY=sk-...
go run main.go providers.go
```

**Note**: The script automatically uses Ollama if no API keys are set. To use cloud providers, simply set their API keys.

### Running Migrations First

```bash
# For Ollama (default, 768 dimensions) - Use the latest migration
psql -h localhost -U postgres -d trip < ../../migrations/20260104000000_update_to_ollama_embeddings.sql

# For OpenAI/Claude (1536 dimensions) - Use the original migration
# psql -h localhost -U postgres -d trip < ../../migrations/20251231000000_create_travel_documents.sql
```

**Important**:
- The new migration (20260104) is configured for **Ollama with 768 dimensions**
- The old migration (20251231) is for **OpenAI/Claude with 1536 dimensions**
- Use the migration that matches your provider choice

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
    embedding vector(768),  -- Ollama default (768), OpenAI/Claude (1536)
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

### Ollama (Local) 🦙

- **Cost**: **FREE!** ✨
- **Requirements**: Local machine with GPU (recommended) or CPU
- **Models**: llama3.1:8b (~4.7GB), nomic-embed-text (~274MB)
- **Performance**: Slower than cloud APIs but completely private and free
- **Best for**: Development, privacy-conscious use, cost-sensitive projects

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

### Ollama Timeout Issues

If you're experiencing timeout errors with Ollama:

```bash
# 1. Check if Ollama is running
curl http://localhost:11434/api/tags

# 2. Verify models are installed
ollama list

# 3. Test model manually
ollama run llama3.1:8b "Hello"
ollama run nomic-embed-text "Test embedding"

# 4. If models are missing, pull them
ollama pull llama3.1:8b
ollama pull nomic-embed-text

# 5. Increase system resources (for slower machines)
# Edit providers.go and increase timeout:
# Timeout: 15 * time.Minute  // Increase from 10 to 15 minutes
```

**Note**: The script now has improved timeout handling:
- HTTP client timeout: 10 minutes (configurable in providers.go:175)
- Per-city context timeout: 15 minutes (main.go:372)
- These timeouts are sufficient for most local LLMs

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

### API Rate Limiting (OpenAI/Claude)

The script includes a 2-second delay between cities to avoid rate limits. If you still hit limits:

```go
// Increase delay in main.go
time.Sleep(5 * time.Second)  // Change from 2 to 5 seconds
```

### Embedding Dimension Mismatch

Different embedding models have different vector dimensions:

| Model | Dimensions | Notes |
|-------|------------|-------|
| text-embedding-ada-002 (OpenAI) | 1536 | Default |
| nomic-embed-text (Ollama) | 768 | Recommended for local |
| mxbai-embed-large (Ollama) | 1024 | Good quality |
| all-minilm (Ollama) | 384 | Fastest, smallest |

If using a model with different dimensions, update:

```sql
-- In migration file (migrations/20251231000000_create_travel_documents.sql)
embedding vector(768)  -- Match your embedding model dimension

-- In models.go (travelknowledge/models.go)
Embedding pgvector.Vector `gorm:"type:vector(768)"`
```

**Important**: If you change embedding models, you'll need to:
1. Drop and recreate the table (or migrate)
2. Regenerate all embeddings with the new model

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
