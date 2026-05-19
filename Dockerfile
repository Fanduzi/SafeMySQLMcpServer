# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY config/ config/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -extldflags '-static'" -o /app/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -extldflags '-static'" -o /app/token ./pkg/token

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for MySQL
RUN apk add --no-cache mysql-client

# Copy binary from builder
COPY --from=builder /app/server /app/server
COPY --from=builder /app/token /app/token

# Copy entrypoint script
COPY scripts/docker-entrypoint.sh /app/docker-entrypoint.sh

# Copy config files
COPY --from=builder /app/config /app/config

# Create non-root user
RUN addgroup -g 1000 mysql && \
    adduser -u 1000 -G mysql -D mysql && \
    chmod +x /app/docker-entrypoint.sh && \
    chown -R mysql:mysql /app

USER mysql

# Set environment variables
ENV JWT_SECRET=""
ENV CONFIG_PATH=/app/config/config.yaml
ENV CONFIG_POLL_INTERVAL=""

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server with validation
ENTRYPOINT ["/app/docker-entrypoint.sh"]
