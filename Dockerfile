# Multi-stage build for Go application

# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main app.go

# Build migration utility
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o migrate ./cmd/migrate

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies including postgresql-client for migrations
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    curl \
    postgresql-client

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .

# Copy migrations and scripts
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/scripts/run-migrations.sh ./scripts/run-migrations.sh

# Make scripts executable
USER root
RUN chmod +x ./scripts/run-migrations.sh
USER appuser

# Create necessary directories
RUN mkdir -p uploads logs && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
