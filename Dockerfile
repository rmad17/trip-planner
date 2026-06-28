# Multi-stage build for Go application

# Stage 1: Build
FROM golang:1.23-alpine AS builder

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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s" -o main app.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    curl \
    postgresql-client

# Download Atlas binary
RUN curl -sSfL https://release.ariga.io/atlas/atlas-linux-amd64-latest -o /usr/local/bin/atlas && \
    chmod +x /usr/local/bin/atlas

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/main .

# Copy migrations, templates, and scripts
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/scripts/entrypoint.sh ./scripts/entrypoint.sh

# Make scripts executable, create directories, and set ownership
USER root
RUN chmod +x ./scripts/entrypoint.sh && \
    mkdir -p uploads logs && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://localhost:8080/health || exit 1

# Run migrations then start the application
CMD ["./scripts/entrypoint.sh"]
