# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o formd-stock .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 monitor && \
    adduser -D -u 1000 -G monitor monitor

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/formd-stock .

# Change ownership
RUN chown monitor:monitor /app/formd-stock

# Switch to non-root user
USER monitor

# Run the binary
ENTRYPOINT ["/app/formd-stock"]
