# Node.js build stage for Svelte UI
FROM node:20-alpine AS ui-builder

WORKDIR /app/web-ui

# Copy package files and install dependencies
COPY web-ui/package.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY web-ui/ ./

# Build the Svelte app (outputs to ../web)
RUN npm run build

# Go build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rexec ./cmd/rexec

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    docker-cli \
    openssh-client

# Create non-root user
RUN adduser -D -g '' rexec

# Create necessary directories
RUN mkdir -p /var/lib/rexec/volumes && \
    chown -R rexec:rexec /var/lib/rexec

# Copy binary from builder
COPY --from=builder /app/rexec /usr/local/bin/rexec

# Copy entrypoint script
COPY --from=builder /app/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Set working directory
WORKDIR /app

# Copy built web UI from ui-builder stage
COPY --from=ui-builder /app/web /app/web

# Expose port
EXPOSE 8080

# Switch to non-root user
USER rexec

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["entrypoint.sh"]
CMD ["rexec"]
