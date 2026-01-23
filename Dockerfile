# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install swag for generating swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Install sql-migrate for database migrations
RUN go install github.com/rubenv/sql-migrate/...@latest

# Generate swagger documentation
RUN swag init -g cmd/sales/main.go -o docs

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sales cmd/sales/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o token cmd/token/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wix cmd/wix/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crawl cmd/crawl/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata su-exec

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy binaries from builder
COPY --from=builder /app/sales .
COPY --from=builder /app/token .
COPY --from=builder /app/wix .
COPY --from=builder /app/crawl .
COPY --from=builder /go/bin/sql-migrate .

# Copy migrations and dbconfig
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/dbconfig.yml .

# Set timezone
ENV TZ=Asia/Tokyo

# Copy entrypoint script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Change ownership
RUN chown -R appuser:appuser /app

# Expose port
EXPOSE 8050

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8050/health || exit 1

# Use entrypoint to setup SSH keys and switch to appuser
ENTRYPOINT ["/docker-entrypoint.sh"]

# Run the application
CMD ["./sales"]