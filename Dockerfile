FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy only go mod files first for better caching
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bot ./cmd/bot/main.go

# Start a new stage from scratch
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/bot .

# Add non-root user
RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /app

# Use non-root user
USER appuser

# Set environment variables
ENV TZ=Europe/Moscow

# Command to run the executable
ENTRYPOINT ["./bot"] 