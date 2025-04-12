FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Verify the directory structure
RUN ls -la && ls -la cmd/bot

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bot ./cmd/bot

# Start a new stage from scratch
FROM alpine:latest

# Add non-root user
RUN adduser -D -g '' appuser

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bot .

# Set ownership
RUN chown -R appuser:appuser /app

# Use non-root user
USER appuser

# Set environment variables
ENV TZ=Europe/Moscow

# Command to run the executable
ENTRYPOINT ["./bot"] 