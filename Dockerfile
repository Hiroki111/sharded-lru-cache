# --- Stage 1: Build ---
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o cache-server ./cmd/cache-server/main.go

# --- Stage 2: Final Image ---
FROM alpine:latest
WORKDIR /app

# Copy the binary to /app/
COPY --from=builder /app/cache-server .

# Create a data directory for the AOF
RUN mkdir /app/data

EXPOSE 8080

# Run the server from /app/
CMD ["./cache-server"]