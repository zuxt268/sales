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

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sales cmd/sales/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o token cmd/token/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy binaries from builder
COPY --from=builder /app/sales .
COPY --from=builder /app/token .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Set timezone
ENV TZ=Asia/Tokyo

# Change ownership
RUN chown -R appuser:appuser /app

# Use non-root user
USER appuser

# Expose port
EXPOSE 8091

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8091/health || exit 1

# Run the application
CMD ["./sales"]