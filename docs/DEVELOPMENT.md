# Development Guide

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.23+ | https://go.dev/dl/ |
| PostgreSQL | 15+ | https://postgresql.org/download/ |
| Atlas CLI | latest | `curl -sSf https://atlasgo.sh | sh` |
| swag CLI | 1.16.5+ | `go install github.com/swaggo/swag/cmd/swag@latest` |

pgvector extension is required for the travel knowledge RAG feature. Install it for your PostgreSQL version before running migrations.

---

## Local Setup

### 1. Clone and install dependencies

```bash
git clone <repo>
cd trip-planner
go mod download
```

### 2. Create the database

```bash
createdb triplanner
psql -d triplanner -c 'CREATE EXTENSION IF NOT EXISTS "pgcrypto";'
psql -d triplanner -c 'CREATE EXTENSION IF NOT EXISTS "vector";'
```

### 3. Configure environment

```bash
cp .env_sample .env
```

Edit `.env` — the minimum required fields for local development:

```env
PORT=8080
DB_URL="host=localhost user=<pg_user> password=<pg_password> dbname=triplanner port=5432 sslmode=disable"
SECRET=any-random-string-for-jwt

# Optional: only needed for Google OAuth login
GOOGLE_OAUTH_CLIENT_ID=
GOOGLE_OAUTH_CLIENT_SECRET=
GOOGLE_OAUTH_CALLBACK_URL=http://localhost:8080/auth/google/callback
FRONTEND_URL=http://localhost:3000

# Optional: only needed for AI trip generation
GEMINI_API_KEY=
DEFAULT_LLM_PROVIDER=gemini

# Optional: only needed for document storage uploads
STORAGE_PROVIDER=local
```

### 4. Run migrations

```bash
atlas migrate apply --env local
```

Atlas reads `atlas.hcl` for the `local` environment. It applies all `.sql` files in `migrations/` in order.

After adding a new migration file, update the checksum:

```bash
atlas migrate hash
```

### 5. Start the server

```bash
go run app.go
```

The server listens on the port set by `PORT` (default `8080`).

---

## API Documentation

Swagger UI is served at:

```
http://localhost:8080/swagger/index.html
```

The `docs/swagger.json` and `docs/swagger.yaml` are pre-generated and committed. If you change handler signatures, request/response types, or add new routes with swagger comments, regenerate:

```bash
$(go env GOPATH)/bin/swag init --generalInfo app.go --output docs
```

### Swagger annotation conventions

- Use `swaggertype:"object"` on `json.RawMessage` fields.
- Use `swaggertype:"array,string"` on `pq.StringArray` fields.
- Public routes (no auth) live under `/api/v1/public/`. Authenticated routes live under `/api/v1/`.

---

## Running Tests

```bash
# All tests (skips DB-dependent tests automatically)
go test ./...

# Specific package
go test ./trips/...
go test ./accounts/...

# With verbose output
go test -v ./trips/...

# Race detector
go test -race ./...
```

Tests that require a live database are skipped with `t.Skip(...)` when no DB is configured. To run them, set `DB_URL` in your environment and ensure migrations have been applied.

---

## Adding a Migration

1. Create a new `.sql` file in `migrations/` with a timestamp prefix:

```bash
touch migrations/$(date +%Y%m%d%H%M%S)_describe_change.sql
```

2. Write the SQL (use `IF NOT EXISTS` / `IF EXISTS` guards so the file is idempotent):

```sql
ALTER TABLE "trip_plans" ADD COLUMN IF NOT EXISTS "new_field" text NULL;
CREATE INDEX IF NOT EXISTS "idx_trip_plans_new_field" ON "trip_plans"("new_field");
```

3. Update the Atlas checksum:

```bash
atlas migrate hash
```

4. Apply locally:

```bash
atlas migrate apply --env local
```

---

## Docker

### Run everything with Docker Compose

```bash
cp .env_sample .env
# fill in DB_USER, DB_PASSWORD, DB_NAME, JWT_SECRET at minimum
docker-compose up --build
```

This starts:
- `db` — PostgreSQL 15
- `migration` — runs migrations once at startup
- `api` — the Go backend (exposed via Caddy)
- `caddy` — reverse proxy (ports 80/443)

### Build image only

```bash
docker build -t trip-planner .
```

---

## Project Structure

```
app.go                  — entry point, router wiring
accounts/               — auth, JWT, Google OAuth, password reset, email verification
core/                   — DB connection, base models, TTL cache, Amadeus client
documents/              — document upload/download (DO Spaces / local)
expenses/               — expense tracking and splitting
flights/                — flight search provider interface + Amadeus implementation
hotels/                 — hotel search provider interface + Amadeus implementation
migrations/             — SQL migration files (applied by Atlas)
notifications/          — push/email notification models and handlers
routes/                 — routing provider interface + Mapbox/Google implementations
travelknowledge/        — pgvector RAG knowledge base for AI trip enrichment
trips/                  — trip plans, hops, days, activities, sharing, lifecycle, transport, checklist
docs/                   — swagger-generated API docs + project documentation
```

---

## Common Tasks

### Check which migrations have been applied

```bash
atlas migrate status --env local
```

### Reset the local database

```bash
dropdb triplanner && createdb triplanner
psql -d triplanner -c 'CREATE EXTENSION IF NOT EXISTS "pgcrypto";'
psql -d triplanner -c 'CREATE EXTENSION IF NOT EXISTS "vector";'
atlas migrate apply --env local
```

### Add a new LLM provider

1. Implement the `LLMProvider` interface in `trips/llm_provider.go`.
2. Register it in the factory function in the same file.
3. Add the provider's API key to `.env_sample`.

### Add a new route provider (maps)

1. Implement `routes.Provider` in a new file under `routes/`.
2. Register it in `routes/factory.go`.
3. Set `ROUTES_PROVIDER=<name>` in `.env`.
