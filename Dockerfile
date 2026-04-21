# Build stage
FROM golang:1.26-alpine3.22 AS builder

# Install build dependencies (sqlite-static for static linking, upx for compression)
RUN apk add --no-cache git gcc musl-dev sqlite-dev sqlite-static upx tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled, static link SQLite
RUN CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" \
    go build -ldflags="-w -s -linkmode external -extldflags '-static'" \
    -o /gateway ./cmd/gateway

# Compress binary with UPX (reduces size by 50-70%)
RUN upx --best --lzma /gateway

# Runtime stage - scratch for minimal image
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo/

WORKDIR /app

# Copy binary from builder
COPY --from=builder /gateway /gateway

# Copy default config
COPY configs/config.yaml ./configs/

# Expose port
EXPOSE 8080

# Set environment variables
ENV GATEWAY_DATABASE_DSN=/app/data/gateway.db

# Run the binary
ENTRYPOINT ["/gateway"]