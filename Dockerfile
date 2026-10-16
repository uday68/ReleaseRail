# syntax=docker/dockerfile:1

# Stage 1: Build binary statically
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Install git and ca-certificates for fetching dependencies if needed
RUN apk add --no-cache git ca-certificates tzdata

# Cache dependency layer
COPY go.mod ./
RUN go mod download

# Copy source tree
COPY . .

# Build arguments for build metadata injection
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown

# Compile static, unstripped binary without CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /bin/pipeforge-service \
    ./cmd/pipeforge-service

# Stage 2: Minimal runtime environment
FROM alpine:3.20

# Create unprivileged non-root user and group
RUN addgroup -g 10001 -S nonroot && \
    adduser -u 10001 -S -G nonroot -h /home/nonroot -D nonroot && \
    apk add --no-cache ca-certificates curl tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /bin/pipeforge-service /app/pipeforge-service

# Default environment configuration
ENV PIPEFORGE_HOST=0.0.0.0 \
    PIPEFORGE_PORT=8080 \
    PIPEFORGE_READ_TIMEOUT=5s \
    PIPEFORGE_WRITE_TIMEOUT=10s \
    PIPEFORGE_IDLE_TIMEOUT=60s \
    PIPEFORGE_SHUTDOWN_TIMEOUT=10s

# Run as nonroot
USER nonroot:nonroot

EXPOSE 8080

HEALTHCHECK --interval=5s --timeout=3s --start-period=3s --retries=3 \
    CMD curl -f http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/pipeforge-service"]
