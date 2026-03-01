# Multi-stage build for development and production

# Stage 1: Development with Air (hot reload)
FROM golang:1.25-alpine AS development

# Install git and air
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 9991

# Use air for hot reload
CMD ["air"]

# Stage 2: Build binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/app cmd/app/main.go

# Stage 3: Production
FROM alpine:latest AS production

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bin/app .

# Expose port
EXPOSE 9991

# Run binary
CMD ["./app"]
