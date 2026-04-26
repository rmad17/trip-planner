# Ollama Setup Guide for Travel Document Generation

This guide shows you how to set up and use Ollama for completely free, local AI-powered travel document generation with embeddings.

## 🚀 Quick Start (5 minutes)

### Step 1: Install and Set Up Ollama

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull required models
ollama pull llama3.1:8b           # For content generation (~4.7GB)
ollama pull nomic-embed-text      # For embeddings (~274MB)

# Verify Ollama is running
curl http://localhost:11434/api/tags

# Test the models
ollama run llama3.1:8b "Write a short sentence about travel."
ollama run nomic-embed-text "Test"
```

### Step 2: Update Vector Dimensions

**Important**: nomic-embed-text uses 768 dimensions, but the default setup uses 1536 (OpenAI). You need to update:

#### 2a. Update the Migration File

Edit `migrations/20251231000000_create_travel_documents.sql`:

```sql
-- Change line 13 from:
embedding vector(1536), -- OpenAI ada-002 or similar dimension

-- To:
embedding vector(768), -- Ollama nomic-embed-text dimension
```

#### 2b. Update the Model File

Edit `travelknowledge/models.go`:

```go
// Change line 33 from:
Embedding    pgvector.Vector `gorm:"type:vector(1536)" json:"-"`

// To:
Embedding    pgvector.Vector `gorm:"type:vector(768)" json:"-"`
```

#### 2c. Apply the Migration

```bash
# If table already exists, drop it first
psql -h localhost -U postgres -d trip -c "DROP TABLE IF EXISTS travel_documents;"

# Apply the migration with updated dimensions
psql -h localhost -U postgres -d trip < migrations/20251231000000_create_travel_documents.sql
```

### Step 3: Set Environment Variables

```bash
# Add to your .env file or export directly
export USE_OLLAMA=true
export OLLAMA_MODEL=llama3.1:8b
export OLLAMA_EMBEDDING_MODEL=nomic-embed-text

# Database connection (required)
export DB_URL=postgres://postgres:postgres@localhost:5432/trip?sslmode=disable
```

### Step 4: Run the Generator

```bash
cd cmd/generate-travel-docs
go run main.go providers.go
```

## ⚙️ Configuration Options

### Recommended Ollama Models

#### Content Generation Models

| Model | Size | Speed | Quality | Best For |
|-------|------|-------|---------|----------|
| llama3.1:8b | 4.7GB | Medium | High | **Recommended** - Best balance |
| llama2:7b | 3.8GB | Fast | Good | Faster generation, good quality |
| mistral:7b | 4.1GB | Medium | High | Alternative to llama3.1 |
| llama3.1:70b | 40GB | Slow | Excellent | High-end systems, best quality |

#### Embedding Models

| Model | Dimensions | Size | Speed | Quality | Best For |
|-------|-----------|------|-------|---------|----------|
| nomic-embed-text | 768 | 274MB | Fast | Good | **Recommended** - Best balance |
| mxbai-embed-large | 1024 | 669MB | Medium | High | Better quality, more storage |
| all-minilm | 384 | 46MB | Very Fast | Decent | Quick testing, resource-limited |

### Example Configurations

#### Default (Recommended)
```bash
export OLLAMA_MODEL=llama3.1:8b
export OLLAMA_EMBEDDING_MODEL=nomic-embed-text
# Dimensions: 768
```

#### High Quality (Requires more RAM/VRAM)
```bash
export OLLAMA_MODEL=llama3.1:70b
export OLLAMA_EMBEDDING_MODEL=mxbai-embed-large
# Update dimensions to 1024 in migration and models.go
```

#### Fast/Lightweight
```bash
export OLLAMA_MODEL=llama2:7b
export OLLAMA_EMBEDDING_MODEL=all-minilm
# Update dimensions to 384 in migration and models.go
```

## 🐛 Troubleshooting

### Timeout Errors

The implementation includes generous timeouts:
- HTTP client: 10 minutes per request
- Per-city context: 15 minutes total

If you still get timeouts:

1. **Check Ollama is running**:
   ```bash
   curl http://localhost:11434/api/tags
   ```

2. **Monitor Ollama logs**:
   ```bash
   journalctl -u ollama -f
   ```

3. **Test models individually**:
   ```bash
   time ollama run llama3.1:8b "Tell me about Paris in 100 words"
   time ollama run nomic-embed-text "Test embedding generation"
   ```

4. **For very slow machines**, edit `cmd/generate-travel-docs/providers.go`:
   ```go
   // Line 175, increase timeout:
   Timeout: 20 * time.Minute, // Change from 10 to 20 minutes
   ```

### Embedding Dimension Errors

Error: `ERROR: expected 1536 dimensions, not 768`

**Solution**: You forgot to update the vector dimensions. See Step 2 above.

### Model Not Found

Error: `model 'nomic-embed-text' not found`

**Solution**:
```bash
ollama pull nomic-embed-text
ollama list  # Verify it's installed
```

### Ollama Not Responding

**Solution**:
```bash
# Restart Ollama
sudo systemctl restart ollama

# Or if installed via script:
killall ollama
ollama serve &
```

### Out of Memory

If generation crashes with OOM errors:

1. **Use smaller models**:
   ```bash
   export OLLAMA_MODEL=llama2:7b  # Instead of llama3.1:8b
   ```

2. **Reduce concurrent operations** (edit main.go):
   ```go
   // Process one city at a time (already default)
   // Increase delay between cities:
   time.Sleep(5 * time.Second)  // Line 386
   ```

3. **Close other applications** to free up RAM/VRAM

## 📊 Performance Expectations

### With GPU (Recommended)

- **Content per city**: ~30-60 seconds
- **Embedding per chunk**: ~1-2 seconds
- **Total for 8 cities**: ~20-30 minutes

### CPU Only

- **Content per city**: 2-5 minutes
- **Embedding per chunk**: ~5-10 seconds
- **Total for 8 cities**: 1-2 hours

## ✅ Verification

After setup, verify everything works:

```bash
# 1. Check models are installed
ollama list

# 2. Test content generation
export USE_OLLAMA=true
export OLLAMA_MODEL=llama3.1:8b
export OLLAMA_EMBEDDING_MODEL=nomic-embed-text

# 3. Run the generator
cd cmd/generate-travel-docs
go run main.go providers.go

# 4. Should see output like:
# 🦙 Using Ollama local LLM
#    Content model: llama3.1:8b
#    Embedding model: nomic-embed-text
```

## 🎯 Benefits of Ollama

✅ **Free**: No API costs
✅ **Private**: Data never leaves your machine
✅ **Offline**: Works without internet
✅ **Customizable**: Use any Ollama-compatible model
✅ **No Rate Limits**: Generate as much as you want

## 📝 Notes

- First run will be slower as models load into memory
- Keep Ollama running in the background for best performance
- GPU highly recommended but not required
- Models stay in memory (~6GB RAM for llama3.1:8b)

---

**Ready to generate travel content with Ollama!** 🦙✨
