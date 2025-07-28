# Build stage with CGO support
FROM golang:1.24 AS builder

WORKDIR /app

# Enable CGO and target OS
ENV CGO_ENABLED=1 GOOS=linux

# Install necessary C libraries (Debian-based)
RUN apt-get update && apt-get install -y gcc libc6-dev && rm -rf /var/lib/apt/lists/*

# Copy dependencies and source code
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Force fresh data folder
RUN rm -rf /app/data && mkdir /app/data

# Build the binary (CGO will be used now)
RUN go build -o service

# Final image with libc support (slim but not scratch)
FROM debian:bookworm-slim

WORKDIR /app

# Install CA certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy binary and assets
COPY --from=builder /app/service .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/data ./data

# Expose internal port
EXPOSE 8080

# Run the web service
ENTRYPOINT ["/app/service"]
