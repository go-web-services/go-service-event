# go-service-event

HTTP API and `pkg/client` module for domain events (CRUD, query, pagination). Built like `lmk-service-website`: Gin, PostgreSQL (pgx), goose migrations, Docker, Swagger.

## Run locally

```bash
cp .env.sample .env
# ensure Postgres is reachable; then:
goose -dir ./migrations postgres "$DSN" up
go run ./cmd/app/main.go
```

## Docker

Requires external network `go-network` (same pattern as other services):

```bash
docker network create go-network
docker compose up --build
```

## Consumer module

```go
import clientapi "github.com/Lomank123/go-service-event/pkg/client/service"
import "github.com/Lomank123/go-service-event/pkg/client/dto"

svc := clientapi.NewEventAPIService("http://go-service-event:8020")
```

Swagger is generated under `docs/` (`swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal`).
