FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/ingest ./cmd/ingest

FROM gcr.io/distroless/static-debian12:nonroot AS api
COPY --from=builder /out/api /usr/local/bin/api
EXPOSE 8080
ENTRYPOINT ["api"]

FROM gcr.io/distroless/static-debian12:nonroot AS ingest
COPY --from=builder /out/ingest /usr/local/bin/ingest
ENTRYPOINT ["ingest"]
