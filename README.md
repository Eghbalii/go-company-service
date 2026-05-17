# Company Service

A production-ready REST microservice for managing companies, built with **Clean Architecture** in Go.

## Architecture

The project follows [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) — dependencies point strictly inward:

```
cmd/api/            ← Entry point: wires all layers and starts the server
internal/
  domain/           ← Innermost layer: entities (Company, User) + port interfaces
  usecase/          ← Business logic; depends only on domain/
  adapter/          ← Bridges use cases ↔ infrastructure
    http/           ← Echo handlers, DTOs, JWT middleware, router
    repository/     ← GORM implementations of domain ports
    event/          ← Kafka implementation of EventPublisher
  infrastructure/   ← Framework setup (DB pool, Kafka writer, Echo config)
internal/integration/ ← Integration tests (testcontainers, build tag: integration)
pkg/                ← Shared, reusable utilities (config, logger, jwt, hash)
migrations/         ← Sequential SQL migration files (golang-migrate)
docker/             ← Multi-stage Dockerfile
```

**Key rules enforced by the architecture:**
- Use cases never import HTTP or database packages
- Repository and event adapter packages implement domain port interfaces
- DTOs live only in the HTTP adapter layer — entities are never serialised directly
- Error mapping from `pkg/apperr` sentinels to HTTP status codes happens in handlers

## Tech Stack

| Concern | Library |
|---|---|
| HTTP | [Echo v4](https://echo.labstack.com/) |
| ORM | [GORM](https://gorm.io/) + `gorm.io/driver/postgres` |
| Event Broker | [segmentio/kafka-go](https://github.com/segmentio/kafka-go) |
| JWT | [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) |
| Logging | [uber-go/zap](https://github.com/uber-go/zap) |
| Config | [spf13/viper](https://github.com/spf13/viper) |
| Validation | [go-playground/validator v10](https://github.com/go-playground/validator) |
| API Docs | [swaggo/swag](https://github.com/swaggo/swag) + echo-swagger |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Tests | [testify](https://github.com/stretchr/testify) + [testcontainers-go](https://golang.testcontainers.org/) |
| Lint | [golangci-lint](https://golangci-lint.run/) |

## Company Entity

| Field | Type | Constraints |
|---|---|---|
| `id` | UUID | Primary key, auto-generated |
| `name` | string | Required, unique, max 15 chars |
| `description` | string | Optional, max 3000 chars |
| `amount_of_employees` | int | Required, ≥ 0 |
| `registered` | bool | Required |
| `type` | enum | `Corporations` \| `NonProfit` \| `Cooperative` \| `Sole Proprietorship` |

## API Reference

| Method | Path | Auth | Description |
|--------|------|:----:|-------------|
| `POST` | `/api/v1/auth/login` | — | Obtain a JWT token |
| `GET` | `/api/v1/companies/:id` | — | Get a single company |
| `POST` | `/api/v1/companies` | ✓ | Create a company |
| `PATCH` | `/api/v1/companies/:id` | ✓ | Partially update a company |
| `DELETE` | `/api/v1/companies/:id` | ✓ | Delete a company |
| `GET` | `/swagger/*` | — | Swagger UI |

Protected routes require `Authorization: Bearer <token>`.

## Local Setup

### Prerequisites

- Go 1.21+
- Docker (for dependencies)
- `make install-tools` — installs `swag`, `golangci-lint`, `migrate`, `goimports`

### Run with Docker Compose

```bash
cp .env.example .env
# Edit .env — set JWT_SECRET at minimum
docker-compose up --build
```

The service starts on `http://localhost:8080`.  
Swagger UI: `http://localhost:8080/swagger/index.html`

### Run locally (without Docker)

1. Start Postgres and Kafka (or point to existing instances via `.env`).
2. Copy and edit the env file:
   ```bash
   cp .env.example .env
   ```
3. Run migrations:
   ```bash
   make migrate-up
   ```
4. Generate Swagger docs:
   ```bash
   make swagger
   ```
5. Start the server:
   ```bash
   make run
   ```

## Makefile Commands

```bash
make build            # Compile binary to ./bin/company-service
make run              # go run ./cmd/api
make test             # Run unit tests
make test-integration # Run integration tests (requires Docker)
make test-cover       # Full coverage report (coverage.html)
make lint             # golangci-lint
make fmt              # gofmt + goimports
make swagger          # Regenerate OpenAPI docs from annotations
make migrate-up       # Apply all pending migrations
make migrate-down     # Roll back last migration batch
make migrate-create   # Interactively create a new migration file
make docker-up        # Start all services via docker-compose
make docker-down      # Stop all services
make install-tools    # Install dev tooling
```

Override the migration database URL:
```bash
DB_URL=postgres://user:pass@host:5432/db?sslmode=disable make migrate-up
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_PORT` | `8080` | HTTP listen port |
| `APP_ENV` | `development` | Environment label |
| `DB_HOST` | `localhost` | Postgres host |
| `DB_PORT` | `5432` | Postgres port |
| `DB_USER` | `postgres` | Postgres username |
| `DB_PASSWORD` | `postgres` | Postgres password |
| `DB_NAME` | `company_db` | Database name |
| `DB_SSLMODE` | `disable` | Postgres SSL mode |
| `JWT_SECRET` | *(required)* | HMAC signing secret (min 32 chars) |
| `JWT_EXPIRY` | `24h` | Token lifetime |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated broker list |
| `KAFKA_TOPIC_COMPANIES` | `company-events` | Topic for company events |
| `LOG_LEVEL` | `info` | `debug` \| `info` \| `warn` \| `error` |
| `LOG_FORMAT` | `json` | `json` \| `console` |

## Migrations

Migrations are managed by [golang-migrate](https://github.com/golang-migrate/migrate/v4) and live in `migrations/`.

```bash
# Apply all pending migrations
make migrate-up

# Roll back the last applied migration
make migrate-down

# Create a new migration pair
make migrate-create   # prompts for migration name
```

The `docker-compose.yml` runs migrations automatically via the `migrate` service before starting the app.

## Testing

```bash
# Unit tests (no external dependencies)
make test

# Single test function
go test -v -run TestCreate_Success ./internal/usecase/company/...

# Integration tests (spins up Postgres via Docker)
make test-integration

# Coverage report
make test-cover
```

Integration tests are gated by the `integration` build tag so `make test` never requires Docker.

## Kafka Events

Every mutating operation publishes an event to `company-events` (configurable):

```json
{
  "event_id": "uuid",
  "event_type": "company.created | company.updated | company.deleted",
  "timestamp": "2024-01-15T10:30:00Z",
  "payload": { ... }
}
```

Events are published asynchronously — they never block the HTTP response.

## Linting

```bash
make lint
```

Configuration is in `.golangci.yml`. Enabled linters include `errcheck`, `staticcheck`, `gosec`, `revive`, `gocritic`, `misspell`, and `cyclop`.
