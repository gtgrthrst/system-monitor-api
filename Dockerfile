# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies for CGO (required by go-sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files first for better cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build with CGO enabled for sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o sysinfo-api .

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/sysinfo-api .

# Create data directory for SQLite and config
RUN mkdir -p /data

# Environment variables
ENV TZ=Asia/Taipei

# Expose port
EXPOSE 8088

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8088/health || exit 1

# Run from /data directory so SQLite database is created there
WORKDIR /data
ENTRYPOINT ["/app/sysinfo-api"]
