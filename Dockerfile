FROM golang:latest AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git || apt-get update && apt-get install -y git

# Install SQLite development libraries
RUN apt-get update && apt-get install -y sqlite3 libsqlite3-dev || apk add --no-cache sqlite sqlite-dev

# Enable automatic toolchain download
ENV GOTOOLCHAIN=auto

# Copy only go mod files first for better caching
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Get SQLite driver
RUN go get github.com/mattn/go-sqlite3

# Build the application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bot ./cmd/bot/main.go

# Start a new stage from scratch
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates tzdata sqlite3 && rm -rf /var/lib/apt/lists/*

# Create data directory for persistent storage
RUN mkdir -p /app/data

# Copy the binary from builder
COPY --from=builder /app/bot .

# Add non-root user
RUN useradd -r -u 1000 -g root appuser && \
    chown -R appuser:root /app

# Use non-root user
USER appuser

# Set environment variables
ENV TZ=Europe/Moscow

# Command to run the executable
ENTRYPOINT ["./bot"]