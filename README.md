Visit main page: [https://github.com/go-web-services](https://github.com/go-web-services)

# Go Web Services - go-service-event

`github.com/go-web-services/go-service-event`

Analytics and event collection service. Stores structured events — page views, button clicks, custom actions — in PostgreSQL. Events are scoped by `project_id`, carry rich metadata, and support optional `user_id` and `session_id` for cross-session attribution. Idempotent ingestion via `message_id` makes it safe for frontend retries and backend job replays.

---

## Responsibilities

- Ingest single events and atomic batches (up to 100 per batch request).
- Guarantee idempotency: duplicate `project_id` + `message_id` pairs return the existing row without error.
- Soft-delete events (retain row, set `deleted_at`).
- Support filtered queries with sort and pagination.
- Expose a Go SDK via `pkg/client` for other services in the ecosystem.

---

## Configuration

| Variable | Purpose | Default |
|----------|---------|---------|
| `APP_PORT` | HTTP listen port | `8020` |
| `APP_ENV` | Environment (`dev` / `prod`) | — |
| `POSTGRES_HOST` | PostgreSQL host | — |
| `POSTGRES_PORT` | PostgreSQL port | — |
| `POSTGRES_USER` | PostgreSQL user | — |
| `POSTGRES_PASSWORD` | PostgreSQL password | — |
| `POSTGRES_DB` | Database name | — |
| `POSTGRES_SSLMODE` | SSL mode | `disable` |

Full list in `config/postgres_config.go` and `.env.sample`.

---

## Run locally

```bash
git clone git@github.com:go-web-services/go-service-event.git
cd go-service-event
cp .env.sample .env
# Set Postgres connection variables
export DSN='postgres://go-service-event:go-service-event-password@127.0.0.1:5437/go-service-event-db?sslmode=disable'
goose -dir ./migrations postgres "$DSN" up
go run ./cmd/app/main.go
```

Requires Go 1.23+ and a reachable PostgreSQL instance (default port `5437` in the sample).

---

## Docker

Uses the external network `go-network` shared with sibling services:

```bash
docker network create go-network
```

- **Dev**:
  ```bash
  docker compose -f docker-compose.yml up
  ```
- **Prod**:
  ```bash
  docker compose -f docker-compose-prod.yml up --build
  ```

---

## API surface

Swagger UI is available at `/swagger` (dev environment only). Regenerate after changes:

```bash
swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal
```

### Events (`/api/v1/events`)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/create` | Ingest one event |
| `POST` | `/create-batch` | Ingest up to 100 events atomically |
| `POST` | `/update` | Update name, payload, user, or session by `id` |
| `POST` | `/delete` | Soft-delete by `id` |
| `POST` | `/detail` | Fetch one event by `id` |
| `POST` | `/query` | List and filter with pagination |

**Create request** — required fields: `project_id`, `message_id` (UUID), `distinct_id`, `name`, `payload` (object), `occurred_at` (RFC3339). Optional: `user_id`, `session_id`, `ip`, `user_agent`.

```json
{
  "project_id": "proj_live_01",
  "message_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
  "distinct_id": "device_or_anon_stable_id",
  "name": "page_viewed",
  "payload": { "path": "/pricing" },
  "occurred_at": "2026-05-01T12:00:00Z"
}
```

Batch ingest wraps the same shape: `POST /api/v1/events/create-batch` with `{ "events": [ /* ... */ ] }`.

Validation errors follow the go-web-platform format; field names match JSON tags (`project_id`, `message_id`, not Go struct names).

---

## Client module (`pkg/client`)

Other services import `github.com/go-web-services/go-service-event/pkg/client`:

```go
import (
    clientapi "github.com/go-web-services/go-service-event/pkg/client/service"
    "github.com/go-web-services/go-service-event/pkg/client/dto"
)

svc := clientapi.NewEventAPIService("http://localhost:8020")
// Available methods: CreateV1, CreateBatchV1, UpdateV1, DeleteV1, DetailV1, QueryV1
```

Errors returned by the client are `*platformError.BaseError` values from `go-web-platform/error`. In Gin handlers, forward with `_ = c.Error(err)` or use `errors.As` to branch on specific error codes.

For local development in a consuming service:

```bash
go mod edit -replace github.com/go-web-services/go-service-event=/path/to/go-service-event
```

---

## Private dependencies

```bash
export GOPRIVATE='github.com/go-web-services/*'
```

This service depends on `go-web-platform`.

---

## Author

[Lomank](https://lomank.com)
