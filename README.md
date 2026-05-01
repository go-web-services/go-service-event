# go-service-event

**Purpose:** Persist and query **analytics-style events** (user actions, system signals) for tenants in the **go-web-services** GitHub organization—the same ingestion flow works whether the caller is a **backend service** or a **browser/mobile client** hitting your gateway.

| For humans                                                   | For LLMs / automation                                                                                                                                                                   |
| ------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Read “What this service does” and “Event JSON shapes” first. | Use the JSON examples as the canonical field names (`snake_case`). All write paths are under `/api/v1/events/*` (POST, JSON body). Idempotency key is `message_id` (UUID per delivery). |

---

## What this service does

- **Ingest** single events or atomic **batches** (Postgres, soft delete).
- **Idempotent creates:** same `project_id` + `message_id` on retry returns the existing row—safe for flaky networks (frontend) and job retries (backend).
- **Query** with filters, sort, and pagination (`POST /api/v1/events/query`).
- **Expose a Go SDK** via `pkg/client` for other services.

Emitters typically send `project_id` from config or API-key resolution (backend), or from client config (frontend). They always generate a **fresh `message_id`** (UUID) per HTTP submission for idempotency, and reuse a stable **`distinct_id`** per anonymous or identified user/device.

---

## Event data model (JSON)

Field names below match the API **exactly**. Optional fields may be omitted. Times are **RFC3339** strings.

### Stored event (`EventDTO`)

What you get back from create/detail/query (and in delete output). Server sets `id`, `received_at`; `deleted_at` appears after soft delete.

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "project_id": "proj_live_01",
  "message_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "distinct_id": "anon_6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "user_id": "usr_42",
  "session_id": "sess_01j8xyz",
  "ip": "203.0.113.10",
  "user_agent": "Mozilla/5.0 (compatible; ExampleBot/1.0)",
  "name": "button_clicked",
  "payload": {
    "button_id": "checkout",
    "surface": "cart_drawer",
    "value": 99.5
  },
  "occurred_at": "2026-04-12T10:30:00.123Z",
  "received_at": "2026-04-12T10:30:00.456Z",
  "deleted_at": null
}
```

### Create request body (`POST /api/v1/events/create`)

Required: `project_id`, `message_id` (UUID), `distinct_id`, `name`, `payload` (object), `occurred_at`. Optional: `user_id`, `session_id`, `ip`, `user_agent`.

Minimal example:

```json
{
  "project_id": "proj_live_01",
  "message_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
  "distinct_id": "device_or_anon_stable_id",
  "name": "page_viewed",
  "payload": {
    "path": "/pricing",
    "referrer": "https://example.com/"
  },
  "occurred_at": "2026-05-01T12:00:00Z"
}
```

Frontend-oriented example (extra client context):

```json
{
  "project_id": "proj_live_01",
  "message_id": "bbbbbbbb-cccc-dddd-eeee-ffffffffffff",
  "distinct_id": "anon_from_local_storage",
  "user_id": "usr_42",
  "session_id": "sess_tab_01",
  "ip": "198.51.100.2",
  "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
  "name": "signup_started",
  "payload": {
    "variant": "control",
    "locale": "en-US"
  },
  "occurred_at": "2026-05-01T15:42:31.887Z"
}
```

Backend job example (`user_id` from auth context, no browser session):

```json
{
  "project_id": "proj_live_01",
  "message_id": "cccccccc-dddd-eeee-ffff-000000000000",
  "distinct_id": "service:batch-import-712",
  "user_id": "usr_import_job",
  "name": "subscription_renewed",
  "payload": {
    "plan": "pro_annual",
    "invoice_id": "inv_1001"
  },
  "occurred_at": "2026-05-01T00:00:05Z"
}
```

Batch ingest wraps the same per-event shape in an `events` array: `POST /api/v1/events/create-batch` with `{ "events": [ /* EventCreateInputDTO, ... */ ] }` (max 100, unique `project_id` + `message_id` within the request).

---

## HTTP API (summary)

| Method & path                      | Role                                           |
| ---------------------------------- | ---------------------------------------------- |
| `POST /api/v1/events/create`       | Ingest one event                               |
| `POST /api/v1/events/create-batch` | Ingest many in one transaction                 |
| `POST /api/v1/events/update`       | Update name / payload / user / session by `id` |
| `POST /api/v1/events/delete`       | Soft delete by `id`                            |
| `POST /api/v1/events/detail`       | Fetch one by `id`                              |
| `POST /api/v1/events/query`        | List/filter with pagination                    |

OpenAPI specs are generated under `docs/` (see below).

---

## Run locally

```bash
cp .env.sample .env
# Postgres must be reachable — default port in sample is 5437
export DSN='postgres://go-service-event:go-service-event-password@127.0.0.1:5437/go-service-event-db?sslmode=disable'
goose -dir ./migrations postgres "$DSN" up
go run ./cmd/app/main.go
```

Default app port **`8020`** (`APP_PORT`).

---

## Docker

Uses external network **`go-network`** (same pattern as sibling services):

```bash
docker network create go-network
docker compose up --build
```

---

## Swagger / docs

Generate or refresh bundled API docs:

```bash
swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal
```

---

## Go client (`pkg/client`)

Consuming modules (same org or downstream services):

```go
import (
    clientapi "github.com/go-web-services/go-service-event/pkg/client/service"
    "github.com/go-web-services/go-service-event/pkg/client/dto"
)

svc := clientapi.NewEventAPIService("http://localhost:8020")
// Methods mirror HTTP: CreateV1, CreateBatchV1, UpdateV1, DeleteV1, DetailV1, QueryV1
```

`SendRequest` (from **go-web-platform**) returns `*BaseError` values from `github.com/go-web-services/go-web-platform/error` on failure. In Gin middleware stacks, forward with `_ = c.Error(err)` or use `errors.As` when you need to branch on code/status.

**Validation responses:** `errors[].field` uses **JSON tag names** (e.g. `project_id`, `message_id`), not Go struct field names—match these in clients and tests.

---

## Related

- Platform bootstrap, logging, and HTTP error handling: **`github.com/go-web-services/go-web-platform`**.


---

## Author

[Lomank](https://lomank.com)
