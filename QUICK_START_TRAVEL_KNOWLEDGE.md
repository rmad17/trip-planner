# Quick Start: Travel Knowledge RAG System

## 🎯 What Was Created

A complete RAG (Retrieval Augmented Generation) system for travel knowledge with:

- ✅ PostgreSQL + pgvector database schema
- ✅ Go models for travel documents with embeddings
- ✅ AI-powered content generator (OpenAI/Claude/Mock)
- ✅ Semantic search service
- ✅ Query CLI tool
- ✅ 8 Indian cities ready to generate: Varanasi, Ayodhya, Vrindavan, Ujjain, Dwarka, Asansol, Kolkata, New Delhi

## 📁 New Files Created

```
triplanner/
├── travelknowledge/                              # NEW MODULE
│   ├── models.go                                 # Data models
│   └── service.go                                # Search service
│
├── cmd/
│   ├── generate-travel-docs/                     # NEW: Content generator
│   │   ├── main.go
│   │   ├── providers.go
│   │   └── README.md
│   │
│   └── query-travel-docs/                        # NEW: Query tool
│       └── main.go
│
├── migrations/
│   └── 20251231000000_create_travel_documents.sql # NEW: pgvector schema
│
├── scripts/
│   └── generate-travel-knowledge.sh              # NEW: Helper script
│
├── TRAVEL_KNOWLEDGE_RAG.md                        # Full documentation
└── QUICK_START_TRAVEL_KNOWLEDGE.md               # This file
```

## 🚀 Usage in 3 Steps

### Step 1: Setup Database (1 minute)

```bash
# Option A: Using Docker (recommended)
docker run -d \
  --name travel-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=trip \
  -p 5432:5432 \
  pgvector/pgvector:pg16

# Option B: Use existing Postgres, just enable pgvector
psql -d trip -c "CREATE EXTENSION vector;"
```

### Step 2: Set API Key (optional)

```bash
# Add to .env file
echo "OPENAI_API_KEY=sk-..." >> .env

# OR use mock provider (no API key needed)
# Just skip this step for testing
```

### Step 3: Generate Content (5 minutes)

```bash
# Easy way - run the script
./scripts/generate-travel-knowledge.sh

# Manual way
cd cmd/generate-travel-docs
go run main.go providers.go
```

## 🎨 What Content Gets Generated

For each of the 8 cities, you get 5 document types:

1. **Historical & Cultural** - Heritage, monuments, traditions
2. **Logistics** - Transportation, accommodation, best times to visit
3. **Food & Dining** - Local cuisine, restaurants, must-try dishes
4. **Activities** - Attractions, things to do, shopping
5. **Nearby Places** - Day trips, surrounding attractions

**Total**: ~40 documents with vector embeddings ready for semantic search

## 🔍 Query Examples

```bash
# Build the query tool first
cd cmd/query-travel-docs
go build -o ../../bin/query-travel-docs
cd ../..

# List all cities
./bin/query-travel-docs -list

# Get database stats
./bin/query-travel-docs -stats

# Get all info for Varanasi
./bin/query-travel-docs -city "Varanasi"

# Get just food info for Varanasi
./bin/query-travel-docs -city "Varanasi" -type food
```

## 💻 Use in Your Code

```go
import (
    "triplanner/core"
    "triplanner/travelknowledge"
)

func main() {
    // Setup
    core.ConnectDB()
    db := core.GetDB()
    service := travelknowledge.NewService(db)

    // Get city documents
    docs, _ := service.GetCityDocuments(ctx, "Varanasi")

    // Get specific type
    foodDocs, _ := service.GetCityDocumentsByType(
        ctx,
        "Varanasi",
        travelknowledge.DocumentTypeFood,
    )

    // Semantic search (after generating embeddings)
    query := travelknowledge.SearchQuery{
        Query:    "best temples to visit",
        CityName: "Varanasi",
        Limit:    5,
    }
    queryEmbedding := generateEmbedding(query.Query)
    results, _ := service.SemanticSearch(ctx, query, queryEmbedding)

    // Use results in your RAG pipeline
    for _, result := range results {
        fmt.Printf("%.2f - %s\n", result.Similarity, result.Document.Title)
    }
}
```

## 💰 Costs

| Provider | Cost for 8 cities | Speed | Quality |
|----------|------------------|-------|---------|
| Mock (testing) | FREE | Fast | Placeholder |
| OpenAI GPT-4 | ~$1-2 | Medium | Excellent |
| Claude 3.5 | ~$0.50-1 | Fast | Excellent |

**Note**: Embeddings require OpenAI (~$0.10 extra) regardless of content provider.

## 🎯 RAG Integration Pattern

```go
// 1. User asks a question
userQuery := "What are the best spiritual places in Varanasi?"

// 2. Generate embedding
embedding := callOpenAIEmbedding(userQuery)

// 3. Search knowledge base
results, _ := service.SemanticSearch(ctx, travelknowledge.SearchQuery{
    Query:    userQuery,
    CityName: "Varanasi",
    Limit:    3,
}, embedding)

// 4. Build context
context := ""
for _, r := range results {
    context += r.Document.Content + "\n\n"
}

// 5. Ask LLM with context
prompt := fmt.Sprintf(`Context: %s\n\nQuestion: %s`, context, userQuery)
answer := callLLM(prompt)  // GPT-4 or Claude
```

## 🔧 Common Tasks

### Add More Cities

Edit `cmd/generate-travel-docs/main.go`, function `GetInitialCities()`:

```go
{
    Name:    "Mumbai",
    Country: "India",
    State:   "Maharashtra",
    Coordinates: &travelknowledge.GeoCoordinates{
        Latitude:  19.0760,
        Longitude: 72.8777,
    },
},
```

Then re-run the generator.

### Update Content for a City

```sql
-- Delete old content
DELETE FROM travel_documents WHERE city_name = 'Varanasi';
```

Then regenerate just that city (modify the script to process only one city).

### Export Data

```bash
# Query tool has export built-in
./bin/query-travel-docs -city "Varanasi" -type food
# Select 'y' when prompted

# Or use pg_dump
pg_dump -h localhost -U postgres -d trip -t travel_documents > backup.sql
```

## 📊 Database Queries

```sql
-- Find similar content (semantic search)
SELECT
    title,
    1 - (embedding <=> '[...]'::vector) as similarity
FROM travel_documents
WHERE city_name = 'Varanasi'
ORDER BY embedding <=> '[...]'::vector
LIMIT 5;

-- Text search
SELECT * FROM travel_documents
WHERE to_tsvector('english', content) @@ plainto_tsquery('english', 'temple');

-- Count by city
SELECT city_name, COUNT(*)
FROM travel_documents
GROUP BY city_name;
```

## 🎓 Learning Resources

- Full docs: `TRAVEL_KNOWLEDGE_RAG.md`
- Generator README: `cmd/generate-travel-docs/README.md`
- pgvector guide: https://github.com/pgvector/pgvector
- RAG patterns: https://www.anthropic.com/index/retrieval-augmented-generation

## 🐛 Troubleshooting

**"pgvector extension not found"**
```bash
psql -d trip -c "CREATE EXTENSION vector;"
```

**"Cannot connect to database"**
```bash
# Check .env has DB_URL
cat .env | grep DB_URL
```

**"API rate limit"**
```bash
# Increase delay in main.go from 2s to 5s
# Or use mock provider for testing
```

## ✅ Verify Setup

```bash
# 1. Check database
psql -d trip -c "\dx vector"  # Should show pgvector

# 2. Check table exists
psql -d trip -c "\dt travel_documents"

# 3. Build both tools
cd cmd/generate-travel-docs && go build
cd ../query-travel-docs && go build

# 4. Generate sample data (mock mode)
cd ../generate-travel-docs
go run main.go providers.go

# 5. Query the data
cd ../..
./bin/query-travel-docs -stats
```

## 🎉 You're Ready!

Your travel knowledge RAG system is set up. Next steps:

1. Generate content for the 8 cities
2. Test semantic search queries
3. Integrate with your trip planning API
4. Build RAG-powered features

**Happy building!** 🚀
