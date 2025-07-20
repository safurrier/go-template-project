# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -s /bin/sh -u 1001 appuser

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.appVersion=$(git describe --tags --always --dirty)" \
    -a -installsuffix cgo \
    -o /out/cli ./cmd/cli

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.appVersion=$(git describe --tags --always --dirty)" \
    -a -installsuffix cgo \
    -o /out/server ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.appVersion=$(git describe --tags --always --dirty)" \
    -a -installsuffix cgo \
    -o /out/worker ./cmd/worker

# ================================
# CLI Runtime Image
# ================================
FROM gcr.io/distroless/static-debian12:nonroot AS cli

COPY --from=builder /out/cli /usr/local/bin/cli
ENTRYPOINT ["cli"]

# ================================
# Server Runtime Image
# ================================
FROM gcr.io/distroless/static-debian12:nonroot AS server

COPY --from=builder /out/server /usr/local/bin/server
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/server", "-health-check"] || exit 1

ENTRYPOINT ["server"]

# ================================
# Worker Runtime Image
# ================================
FROM gcr.io/distroless/static-debian12:nonroot AS worker

COPY --from=builder /out/worker /usr/local/bin/worker
ENTRYPOINT ["worker"]

# ================================
# Default target (CLI)
# ================================
FROM cli AS default
