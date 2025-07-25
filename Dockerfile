# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

# Enable static linking
ENV CGO_ENABLED=0 GOOS=linux

# Copy dependencies and source code
# Separate step to help layer caching 
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Force fresh data folder 
RUN rm -rf /app/data
RUN mkdir /app/data

RUN go build -o service

# Minimal final image
FROM scratch

WORKDIR /app

# Copy binary and assets
COPY --from=builder /app/service .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/data ./data

# Expose internal port
EXPOSE 8080

# Run the web service
ENTRYPOINT ["/app/service"]
