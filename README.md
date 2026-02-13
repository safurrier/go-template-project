# CBB Team Trends Dashboard Monorepo

Go monorepo with two binaries:
- `cmd/api`: serves precomputed dashboard payload endpoints.
- `cmd/ingest`: ingests game CSV data, computes rolling windows, and stores snapshots.

Includes:
- Postgres persistence
- sqlc-style query layer (`internal/db/queries`, `internal/db/sqlc`)
- Next.js frontend in `frontend/`

## Quick start

```bash
make setup
docker compose -f docker/docker-compose.yml up --build
```

API endpoint:
- `GET /api/v1/teams/arizona-mbb/dashboard?window=5`

## Local dev without Docker

```bash
export DATABASE_URL='postgres://cbb:cbb@localhost:5432/cbb?sslmode=disable'
go run ./cmd/ingest
go run ./cmd/api
```

## Testing

```bash
make check
```
