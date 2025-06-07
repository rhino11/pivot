# Multi-stage build for minimal final image
FROM golang:1.22-alpine AS builder

# Install git and ca-certificates for dependency fetching
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty) -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o pivot ./cmd/main.go

# Final stage - minimal image
FROM alpine:3.18

# Install ca-certificates and sqlite
RUN apk add --no-cache ca-certificates sqlite

# Create non-root user
RUN adduser -D -s /bin/sh pivot

# Copy binary from builder stage
COPY --from=builder /app/pivot /usr/local/bin/pivot

# Set up working directory
WORKDIR /home/pivot
USER pivot

# Default command
ENTRYPOINT ["pivot"]
CMD ["--help"]
