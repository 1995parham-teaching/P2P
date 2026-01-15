# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /p2p ./cmd/p2p

# Runtime stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /p2p /app/p2p

# Create shared directory
RUN mkdir -p /app/shared

# Copy configs
COPY configs/ /app/configs/

ENTRYPOINT ["/app/p2p"]
