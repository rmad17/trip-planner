# Travel Management App - Backend API

[![Test and Coverage](https://github.com/rmad17/trip-planner/actions/workflows/test-coverage.yml/badge.svg)](https://github.com/rmad17/trip-planner/actions/workflows/test-coverage.yml)
[![codecov](https://codecov.io/gh/rmad17/trip-planner/branch/main/graph/badge.svg)](https://codecov.io/gh/rmad17/trip-planner)
[![Go Report Card](https://goreportcard.com/badge/github.com/rmad17/trip-planner)](https://goreportcard.com/report/github.com/rmad17/trip-planner)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A robust backend API for a travel management application built with Go, Gin, GORM, PostgreSQL, and Atlas migrations.

## 🚀 Features

- **User Authentication**: JWT-based auth with Google OAuth integration
- **Trip Planning**: Create and manage travel itineraries
- **Place Management**: Integration with mapping services for location data
- **Database Migrations**: Version-controlled schema management with Atlas
- **RESTful API**: Clean, well-structured REST endpoints
- **Modular Architecture**: Organized codebase with separate modules

## 🛠 Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/)
- **Database**: PostgreSQL 15+
- **Migrations**: [Atlas](https://atlasgo.io/)
- **Authentication**: JWT + Google OAuth
- **Environment**: Docker-ready setup

## 📁 Project Structure

```
triplanner/
├── accounts/              # User authentication & management
│   ├── models.go         # User model
│   ├── auth.go           # Auth handlers
│   ├── middlewares.go    # JWT middleware
│   └── routers.go        # Auth routes
├── trips/                # Trip planning module
│   ├── models.go         # Trip models
│   ├── controllers.go    # Trip handlers
│   └── routers.go        # Trip routes
├── places/               # Places management
├── core/                 # Shared utilities
│   ├── models.go         # Base model with UUID
│   ├── database.go       # DB connection
│   └── loadEnvs.go       # Environment setup
├── cmd/
│   ├── migrate/          # GORM migration tool
│   └── atlas-loader/     # Atlas schema loader
├── migrations/           # Atlas migration files
├── atlas.hcl            # Atlas configuration
├── app.go               # Main application
└── go.mod
```

## 🚦 Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+
- [Atlas CLI](https://atlasgo.io/getting-started#installation)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/triplanner-backend.git
   cd triplanner-backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install Atlas CLI**
   ```bash
   # macOS
   brew install ariga/tap/atlas
   
   # Linux/Windows
   curl -sSf https://atlasgo.sh | sh
   ```

### Environment Setup

1. **Create `.env` file**
   ```bash
   cp .env.example .env
   ```

2. **Configure environment variables**
   ```env
   # Database
   DB_URL=postgres://username:password@localhost:5432/triplanner_dev?sslmode=disable
   
   # Authentication
   SECRET=your-jwt-secret-key
   GOOGLE_OAUTH_CLIENT_ID=your-google-client-id
   GOOGLE_OAUTH_CLIENT_SECRET=your-google-client-secret
   
   # External APIs
   MAPBOX_TOKEN=your-mapbox-token
   ```

### Database Setup

1. **Create PostgreSQL database**
   ```bash
   createdb triplanner_dev
   ```

2. **Enable UUID extension**
   ```bash
   psql triplanner_dev -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
   ```

3. **Run initial migration**
   ```bash
   # Option 1: Using GORM AutoMigrate (for development)
   go run cmd/migrate/main.go
   
   # Option 2: Using Atlas (recommended for production)
   atlas migrate apply --env local --baseline $(ls migrations/ | head -1 | cut -d'_' -f1)
   ```

4. **Verify database setup**
   ```bash
   atlas migrate status --env local
   ```

## 🔧 Development Workflow

### Running the Application

```bash
# Start the development server
go run app.go

# Server will start on http://localhost:8080
```

### Database Migrations

#### Using Atlas (Recommended)

1. **Generate migration after model changes**
   ```bash
   atlas migrate diff describe_your_change --env local
   ```

2. **Apply migrations**
   ```bash
   atlas migrate apply --env local
   ```

3. **Check migration status**
   ```bash
   atlas migrate status --env local
   ```

4. **Rollback if needed**
   ```bash
   atlas migrate apply --env local --to-version 20240101000001
   ```

#### Using GORM (Development Only)

```bash
# Quick development migration
go run cmd/migrate/main.go
```

### Testing Atlas Setup

```bash
# Test the Atlas loader
go run cmd/atlas-loader/main.go

# Validate migrations
atlas migrate validate --env local

# Inspect current schema
atlas schema inspect --env local
```

## 🧪 Testing

This project has comprehensive unit and integration test coverage across all modules.

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
go test -race ./...
```

### Test Coverage by Module

The project includes comprehensive tests for:
- **accounts/** - User authentication, JWT middleware, models
- **expenses/** - Expense models, split calculations, settlements
- **documents/** - Document management, storage providers
- **places/** - Mapbox API integration, place data structures
- **storage/** - Storage provider interface, multi-provider management
- **trips/** - Trip planning, activities, itineraries
- **core/** - Base models, date types, helper functions
- **subscriptions/** - Subscription tiers, limits, usage tracking
- **featureflags/** - Feature flag evaluation and management

### Coverage Reports

Coverage reports are automatically generated on every pull request and can be viewed:
- In the [GitHub Actions](https://github.com/rmad17/trip-planner/actions) tab
- On [Codecov](https://codecov.io/gh/rmad17/trip-planner)
- As PR comments with detailed coverage breakdown

### Continuous Integration

Every pull request to `main` triggers automated tests:
- ✅ Unit tests across all modules
- ✅ Integration tests with PostgreSQL
- ✅ Code coverage analysis
- ✅ Race condition detection
- ✅ Code linting with golangci-lint

## 📚 API Documentation

### Authentication Endpoints

```
POST   /api/v1/auth/signup              # Create new user
POST   /api/v1/auth/login               # User login
GET    /api/v1/auth/google/login        # Google OAuth login page
GET    /api/v1/auth/google/begin        # Start Google OAuth flow
GET    /api/v1/auth/google/callback     # Google OAuth callback
```

### Protected Endpoints (Require JWT)

```
GET    /api/v1/user/profile             # Get user profile
POST   /api/v1/trips/create             # Create new trip
GET    /api/v1/places/*                 # Places API endpoints
```

### Example API Usage

**User Registration:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}'
```

**Create Trip:**
```bash
curl -X POST http://localhost:8080/api/v1/trips/create \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "place_name": "Paris",
    "place_id": "paris_123",
    "start_date": "2024-07-01T00:00:00Z",
    "end_date": "2024-07-07T00:00:00Z",
    "min_days": 5
  }'
```

## 🏗 Database Schema

### Users Table
- `id` (UUID, Primary Key)
- `username` (String, Unique)
- `password` (String, Hashed)
- `email` (String, Unique, Optional)
- `created_at`, `updated_at` (Timestamps)

### Trip Plans Table
- `id` (UUID, Primary Key)
- `place_name`, `place_id` (String)
- `start_date`, `end_date` (Timestamp, Optional)
- `min_days` (Integer, Optional)
- `travel_mode`, `notes` (String, Optional)
- `hotels`, `tags` (String Array)
- `user_id` (UUID, Foreign Key)
- `created_at`, `updated_at` (Timestamps)

## 🚀 Deployment

### Environment Configuration

**Production `.env`:**
```env
DB_URL=postgresql://user:pass@prod-host:5432/triplanner_prod?sslmode=require
DATABASE_URL=${DB_URL}  # For Atlas production env
SECRET=your-production-secret
```

### Atlas Production Migrations

```bash
# Apply migrations to production
atlas migrate apply --env production

# Validate production schema
atlas migrate validate --env production
```

### Docker Support

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main app.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create feature branch**: `git checkout -b feature/amazing-feature`
3. **Make changes and test**
4. **Run migrations**: `atlas migrate diff feature_name --env local`
5. **Commit changes**: `git commit -m 'Add amazing feature'`
6. **Push to branch**: `git push origin feature/amazing-feature`
7. **Open Pull Request**

### Development Guidelines

- Follow Go conventions and use `go fmt`
- Add tests for new features
- Update API documentation
- Include migration files for schema changes
- Test migrations before submitting PR

## 🐛 Troubleshooting

### Common Issues

**Atlas migration errors:**
```bash
# If you get "database not clean" error
atlas migrate apply --env local --allow-dirty

# Or establish baseline
atlas migrate apply --env local --baseline 20240101000001
```

**UUID extension missing:**
```bash
psql $DB_URL -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
```

**Import circular dependency:**
```bash
# Check for circular imports
go build ./...
```

### Database Reset (Development)

```bash
# Drop all tables and restart
psql $DB_URL -c "DROP TABLE IF EXISTS trip_plans, users CASCADE;"
go run cmd/migrate/main.go
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [GORM](https://gorm.io/)
- [Atlas](https://atlasgo.io/)
- [PostgreSQL](https://www.postgresql.org/)

---

## 📞 Support

If you have any questions or run into issues, please [open an issue](https://github.com/yourusername/triplanner-backend/issues) on GitHub.

**Happy coding! ✈️**
