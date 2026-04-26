# Travel Knowledge Base with RAG Support

This module provides AI-generated travel content stored in PostgreSQL with pgvector for semantic search and Retrieval Augmented Generation (RAG) applications.

## 🎯 Overview

The travel knowledge system generates comprehensive travel guides for cities and stores them with vector embeddings for semantic search. This enables your trip planner to:

- Answer natural language questions about destinations
- Provide contextually relevant travel recommendations
- Power AI-assisted trip planning with accurate, up-to-date information
- Perform semantic search across travel content

## 📂 Project Structure

```
triplanner/
├── travelknowledge/                   # Travel knowledge module
│   ├── models.go                      # Data models and types
│   └── service.go                     # Query and search service
│
├── cmd/
│   ├── generate-travel-docs/          # Content generation script
│   │   ├── main.go                    # Main generator
│   │   ├── providers.go               # AI provider implementations
│   │   └── README.md                  # Generator documentation
│   │
│   └── query-travel-docs/             # Query utility
│       └── main.go                    # CLI query tool
│
├── migrations/
│   └── 20251231000000_create_travel_documents.sql  # Database schema
│
└── scripts/
    └── generate-travel-knowledge.sh   # Generation helper script
```

## 🚀 Quick Start

### 1. Prerequisites

```bash
# Start PostgreSQL with pgvector
docker run -d \
  --name travel-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=trip \
  -p 5432:5432 \
  pgvector/pgvector:pg16
```

### 2. Environment Setup

Add to your `.env` file:

```env
# Database (required)
DB_URL=postgres://postgres:postgres@localhost:5432/trip?sslmode=disable

# AI Provider - Ollama is now the default (free & local)!
# No API keys needed - just install Ollama and pull models:
# ollama pull llama3.1:8b
# ollama pull nomic-embed-text

# Optional: Override with cloud providers
# OPENAI_API_KEY=sk-...           # For OpenAI content + embeddings
# OR
# ANTHROPIC_API_KEY=sk-ant-...    # For Claude (still needs OpenAI for embeddings)
# OPENAI_API_KEY=sk-...
```

### 3. Generate Travel Content

```bash
# Easy way - using the helper script
./scripts/generate-travel-knowledge.sh

# Or manually
cd cmd/generate-travel-docs
go run main.go providers.go
```

This will generate content for 8 Indian cities:
- Varanasi, Ayodhya, Vrindavan, Ujjain, Dwarka, Asansol, Kolkata, New Delhi

### 4. Query the Knowledge Base

```bash
# Build the query tool
cd cmd/query-travel-docs
go build -o ../../bin/query-travel-docs

# List all cities
./bin/query-travel-docs -list

# Show statistics
./bin/query-travel-docs -stats

# Get all info for a city
./bin/query-travel-docs -city "Varanasi"

# Get specific type of info
./bin/query-travel-docs -city "Varanasi" -type food
```

## 📊 Database Schema

### travel_documents Table

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| city_name | VARCHAR(255) | City name (indexed) |
| city_country | VARCHAR(255) | Country name |
| document_type | VARCHAR(100) | Category (indexed) |
| title | VARCHAR(500) | Document title |
| content | TEXT | Document content |
| metadata | JSONB | Structured metadata |
| embedding | vector(768) | Vector embedding for semantic search (Ollama default) |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |

### Document Types

1. **historical** - Historical & cultural information
2. **logistics** - Travel planning & logistics
3. **food** - Cuisine & dining recommendations
4. **activities** - Attractions & things to do
5. **nearby_places** - Nearby destinations & day trips

## 🔍 Semantic Search

### Using the Service in Go

```go
import (
    "triplanner/core"
    "triplanner/travelknowledge"
)

// Initialize
core.ConnectDB()
db := core.GetDB()
service := travelknowledge.NewService(db)

// Semantic search
query := travelknowledge.SearchQuery{
    Query:         "best spiritual experiences",
    CityName:      "Varanasi",
    DocumentTypes: []travelknowledge.DocumentType{
        travelknowledge.DocumentTypeHistorical,
        travelknowledge.DocumentTypeActivities,
    },
    Limit:         5,
    MinSimilarity: 0.7,
}

// Generate embedding for query (using your embedding provider)
queryEmbedding := generateEmbedding(query.Query)

// Search
results, err := service.SemanticSearch(ctx, query, queryEmbedding)
for _, result := range results {
    fmt.Printf("%.2f - %s\n", result.Similarity, result.Document.Title)
    fmt.Println(result.Document.Content)
}
```

### Direct SQL Queries

```sql
-- Find documents similar to a query embedding
SELECT
    city_name,
    title,
    content,
    1 - (embedding <=> '[0.1, 0.2, ...]'::vector) as similarity
FROM travel_documents
WHERE city_name = 'Varanasi'
    AND document_type = 'food'
ORDER BY embedding <=> '[0.1, 0.2, ...]'::vector
LIMIT 5;

-- Hybrid search: filter + semantic
SELECT
    city_name,
    title,
    1 - (embedding <=> $1::vector) as similarity
FROM travel_documents
WHERE
    city_country = 'India'
    AND document_type = 'activities'
    AND 1 - (embedding <=> $1::vector) >= 0.7
ORDER BY embedding <=> $1::vector
LIMIT 10;
```

## 🤖 AI Providers

### Ollama (Default - Recommended) 🦙

```bash
# Install and pull models
ollama pull llama3.1:8b
ollama pull nomic-embed-text

# No API keys needed!
go run main.go providers.go
```

- **Content**: llama3.1:8b (or any Ollama model)
- **Embeddings**: nomic-embed-text (768 dimensions)
- **Cost**: **FREE!** ✨
- **Privacy**: Runs entirely on your machine
- **Speed**: Fast with GPU, decent with CPU

### OpenAI

```bash
export OPENAI_API_KEY=sk-...
```

- **Content**: GPT-4 Turbo
- **Embeddings**: text-embedding-ada-002 (1536 dimensions)
- **Cost**: ~$1-2 for 8 cities
- **Note**: Requires changing vector dimensions to 1536

### Claude

```bash
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-...  # Still needed for embeddings
```

- **Content**: Claude 3.5 Sonnet
- **Embeddings**: OpenAI (Claude doesn't have embedding API, 1536 dimensions)
- **Cost**: ~$0.50-1 for content + $0.10 for embeddings
- **Note**: Requires changing vector dimensions to 1536

### Mock Provider

No API key needed - generates placeholder content for testing.

```bash
# Just don't set any API keys and don't install Ollama
go run main.go providers.go
```

## 📈 Service Methods

```go
// Get all documents for a city
docs, err := service.GetCityDocuments(ctx, "Varanasi")

// Get documents by type
docs, err := service.GetDocumentsByType(ctx, travelknowledge.DocumentTypeFood)

// Get city-specific type
docs, err := service.GetCityDocumentsByType(ctx, "Varanasi", travelknowledge.DocumentTypeFood)

// Semantic search
results, err := service.SemanticSearch(ctx, query, embedding)

// Get aggregated city info
info, err := service.GetCityInfo(ctx, "Varanasi")

// List all cities
cities, err := service.ListCities(ctx)

// Get statistics
stats, err := service.GetDocumentStats(ctx)

// Hybrid search (text + vector)
results, err := service.HybridSearch(ctx, "temples", embedding, filters, 10)
```

## 🎨 Customization

### Add More Cities

Edit `cmd/generate-travel-docs/main.go`:

```go
func GetInitialCities() []travelknowledge.City {
    return []travelknowledge.City{
        // Existing cities...
        {
            Name:    "Mumbai",
            Country: "India",
            State:   "Maharashtra",
            Coordinates: &travelknowledge.GeoCoordinates{
                Latitude:  19.0760,
                Longitude: 72.8777,
            },
            Description: "Financial capital of India",
        },
    }
}
```

### Customize Content Generation

Edit the prompt in `cmd/generate-travel-docs/providers.go`:

```go
func buildContentPrompt(city travelknowledge.City) string {
    return fmt.Sprintf(`Your custom prompt for %s...`, city.Name)
}
```

### Add Custom Document Types

1. Update the enum in `travelknowledge/models.go`:
```go
const (
    DocumentTypeTransport DocumentType = "transport"
    // ...
)
```

2. Update the content generation to include the new type

### Change Embedding Model

For different dimensions (e.g., 768 for sentence-transformers):

1. Update migration:
```sql
embedding vector(768)
```

2. Update model:
```go
Embedding pgvector.Vector `gorm:"type:vector(768)"`
```

## 🔧 Maintenance

### Update Content for a City

```go
// Delete old content
db.Where("city_name = ?", "Varanasi").Delete(&travelknowledge.TravelDocument{})

// Regenerate
generator.GenerateForCity(ctx, varanasiCity)
```

### Backup Data

```bash
# Export to JSON
./bin/query-travel-docs -city "Varanasi" -type food
# Then select 'y' to export

# Or use pg_dump
pg_dump -h localhost -U postgres -d trip -t travel_documents > travel_knowledge_backup.sql
```

## 💡 Integration Examples

### RAG with LLM

```go
// 1. User asks a question
userQuery := "What are the best temples to visit in Varanasi?"

// 2. Generate embedding for the question
embedding := generateEmbedding(userQuery)

// 3. Search knowledge base
results, _ := service.SemanticSearch(ctx, travelknowledge.SearchQuery{
    Query:    userQuery,
    CityName: "Varanasi",
    Limit:    3,
}, embedding)

// 4. Build context from results
context := ""
for _, result := range results {
    context += result.Document.Content + "\n\n"
}

// 5. Send to LLM with context
prompt := fmt.Sprintf(`Based on this information:

%s

Answer the question: %s`, context, userQuery)

answer := callLLM(prompt)
```

### Trip Planning Assistant

```go
// Multi-city trip planning
cities := []string{"Delhi", "Varanasi", "Kolkata"}
itinerary := []string{}

for _, city := range cities {
    // Get recommendations
    docs, _ := service.GetCityDocumentsByType(ctx, city, travelknowledge.DocumentTypeActivities)

    // Use LLM to create day plan with retrieved context
    dayPlan := createDayPlan(city, docs)
    itinerary = append(itinerary, dayPlan)
}
```

## 📝 Notes

- **Cost**: Using real AI APIs costs ~$1-2 for the initial 8 cities
- **Time**: Generation takes ~5-10 minutes depending on API speed
- **Quality**: AI-generated content should be reviewed before production use
- **Updates**: Re-run the generator periodically to refresh content
- **Scale**: The current schema supports thousands of cities efficiently

## 🐛 Troubleshooting

### "pgvector extension not found"

```bash
# In psql
CREATE EXTENSION vector;
```

### "Failed to connect to database"

Check your DB_URL in `.env` and ensure PostgreSQL is running.

### "API rate limit exceeded"

Increase the delay in main.go or use a higher tier API key.

### "Embedding dimension mismatch"

Ensure migration vector dimension matches your embedding model output.

## 📚 Resources

- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [OpenAI Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)
- [RAG Best Practices](https://www.anthropic.com/index/retrieval-augmented-generation)

## 🎯 Next Steps

1. ✅ Generate initial dataset
2. 🔄 Integrate with your trip planning API
3. 🔍 Test semantic search with real queries
4. 📊 Monitor usage and costs
5. 🌍 Expand to more cities
6. 🎨 Customize content for your use case

---

**Ready to build intelligent travel experiences with RAG!** 🚀
