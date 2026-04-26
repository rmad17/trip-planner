# Quick Start with Ollama (Default Setup)

This is the fastest way to get started with travel document generation using **Ollama** - completely free and local!

## Prerequisites

- PostgreSQL with pgvector
- Ollama installed
- Go 1.21+

## 5-Minute Setup

### Step 1: Install Ollama

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull required models (~5GB download)
ollama pull llama3.1:8b           # Content generation
ollama pull nomic-embed-text      # Embeddings

# Verify
ollama list
```

### Step 2: Setup Database

```bash
# Start PostgreSQL with pgvector (if not already running)
docker run -d \
  --name travel-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=trip \
  -p 5432:5432 \
  pgvector/pgvector:pg16

# Apply migration
psql -h localhost -U postgres -d trip < migrations/20260104000000_update_to_ollama_embeddings.sql
```

### Step 3: Configure Environment

```bash
# Database (required)
export DB_URL=postgres://postgres:postgres@localhost:5432/trip?sslmode=disable

# Ollama is now the default - no additional config needed!
# Optional: Customize models
# export OLLAMA_MODEL=llama3.1:8b
# export OLLAMA_EMBEDDING_MODEL=nomic-embed-text
```

### Step 4: Generate Content

```bash
cd cmd/generate-travel-docs
go run main.go providers.go
```

You should see:
```
🦙 Using Ollama local LLM (default)
   Content model: llama3.1:8b
   Embedding model: nomic-embed-text (768 dimensions)
```

## What Gets Generated

For each of the 8 cities, the system generates:
- Historical & Cultural Information
- Travel Logistics (transport, accommodation)
- Food & Dining recommendations
- Activities & Attractions
- Nearby Places of Interest

## Expected Performance

### With GPU (NVIDIA):
- **Time**: ~20-30 minutes for all 8 cities
- **Speed**: ~30-60 seconds per city
- **Cost**: FREE!

### With CPU Only:
- **Time**: ~1-2 hours for all 8 cities
- **Speed**: 2-5 minutes per city
- **Cost**: Still FREE!

## Verify Results

```bash
# Check generated documents
psql -h localhost -U postgres -d trip -c "SELECT city_name, COUNT(*) FROM travel_documents GROUP BY city_name;"

# Should show ~5 documents per city (one for each category)
```

## Using Other Providers (Optional)

If you want to switch to cloud providers later:

### Switch to Claude:
```bash
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-...  # For embeddings
# You'll need to update dimensions to 1536 and regenerate
```

### Switch to OpenAI:
```bash
export OPENAI_API_KEY=sk-...
# You'll need to update dimensions to 1536 and regenerate
```

## Troubleshooting

### "Failed to connect to Ollama"
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# If not, start it
ollama serve
```

### "Model not found"
```bash
# Pull the models again
ollama pull llama3.1:8b
ollama pull nomic-embed-text
```

### Slow Generation
- First run is always slower (models loading)
- Close other applications to free RAM
- GPU will significantly speed up generation

## Next Steps

1. ✅ Content generated locally
2. 🔍 Test semantic search: `cd cmd/query-travel-docs && go run main.go`
3. 🚀 Integrate with your trip planner API
4. 📈 Scale to more cities as needed

---

**You're all set! Enjoy free, private, local AI-powered travel content generation!** 🦙✨
