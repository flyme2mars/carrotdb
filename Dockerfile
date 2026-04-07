# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server, CLI, and bench tool
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /carrotdb-server ./cmd/carrotdb-server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /carrotdb ./cmd/carrotdb/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /carrotdb-bench ./cmd/carrotdb-bench/main.go

# Stage 2: Final Image
FROM alpine:latest

# Install basic tools and certificates
RUN apk add --no-cache ca-certificates

# Create a non-root user
RUN adduser -D -u 1000 carrot
USER carrot

# Set working directory
WORKDIR /home/carrot

# Copy binaries from builder
COPY --from=builder /carrotdb-server .
COPY --from=builder /carrotdb .
COPY --from=builder /carrotdb-bench .

# Default data directory
RUN mkdir -p /home/carrot/data
ENV CARROT_DATA_DIR=/home/carrot/data

# Expose all relevant ports
# 6379: API
# 7000: Raft
# 8000: Router
# 8080: Dashboard
# 9000: Gossip
EXPOSE 6379 7000 8000 8080 9000

# Set entrypoint to the server
ENTRYPOINT ["./carrotdb-server"]

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/status || exit 1
