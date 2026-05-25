# AntiGravity Prompt: Phase 0 — Foundation

## Context
You are implementing the foundational scaffold for Acumius, a Go-based agent collaboration fabric. The project already has a basic scaffold with a health endpoint. Your task is to complete Phase 0 by adding the Docker Compose baseline and ensuring everything is production-ready for Phase 1.

## Current State
- Go module initialized: `github.com/Acumius/Acumius`
- Entrypoint: `cmd/acumius/main.go` with HTTP server and graceful shutdown
- Config: `internal/config/config.go` with env var loading
- API: `internal/api/router.go`, `internal/api/health.go`, `internal/api/health_test.go`
- Makefile with: `fmt`, `fmt-check`, `lint`, `test`, `check`
- CI: `.github/workflows/go-quality.yml`
- Migrations folder with README
- Examples folder with README

## Your Task

### 1. Docker Compose Baseline
Create `docker-compose.yml` in the project root with:

```yaml
services:
  acumius:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ACUMIUS_DATABASE_URL=postgres://acumius:acumius@postgres:5432/acumius?sslmode=disable
      - ACUMIUS_VALKEY_URL=valkey:6379
    depends_on:
      postgres:
        condition: service_healthy
      valkey:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 5

  postgres:
    image: ankane/pgvector:latest
    environment:
      - POSTGRES_USER=acumius
      - POSTGRES_PASSWORD=acumius
      - POSTGRES_DB=acumius
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U acumius"]
      interval: 5s
      timeout: 5s
      retries: 5

  valkey:
    image: valkey/valkey:8
    ports:
      - "6379:6379"
    volumes:
      - valkey_data:/data
    healthcheck:
      test: ["CMD", "valkey-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  governance-ui:
    build: ./governance-ui
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080
    depends_on:
      - acumius

volumes:
  postgres_data:
  valkey_data:
```

### 2. Dockerfile
Create `Dockerfile` for the Go service:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o acumius ./cmd/acumius

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/acumius .
EXPOSE 8080
CMD ["./acumius"]
```

### 3. Environment Configuration
Update `internal/config/config.go` to support:

```go
type Config struct {
    ServerPort    string `env:"ACUMIUS_SERVER_PORT" envDefault:"8080"`
    DatabaseURL   string `env:"ACUMIUS_DATABASE_URL" envDefault:"postgres://acumius:acumius@localhost:5432/acumius?sslmode=disable"`
    ValkeyURL     string `env:"ACUMIUS_VALKEY_URL" envDefault:"localhost:6379"`
    LogLevel      string `env:"ACUMIUS_LOG_LEVEL" envDefault:"info"`
    Environment   string `env:"ACUMIUS_ENV" envDefault:"development"`
}
```

Use `github.com/caarlos0/env/v11` for env parsing.

### 4. Makefile Updates
Add to Makefile:

```makefile
.PHONY: up down migrate migrate-create

up:
	docker-compose up -d

down:
	docker-compose down

migrate:
	migrate -path migrations -database $(ACUMIUS_DATABASE_URL) up

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)
```

### 5. .env.example
Create `.env.example`:

```bash
ACUMIUS_SERVER_PORT=8080
ACUMIUS_DATABASE_URL=postgres://acumius:acumius@localhost:5432/acumius?sslmode=disable
ACUMIUS_VALKEY_URL=localhost:6379
ACUMIUS_LOG_LEVEL=info
ACUMIUS_ENV=development
```

### 6. Migration System Setup
Install golang-migrate locally. Create first migration:

```bash
make migrate-create name=init
```

This should create `migrations/000001_init.up.sql` and `migrations/000001_init.down.sql`.

Leave them empty for now — Phase 1 will populate them.

### 7. Storage Connection Scaffolding
Create placeholder files:

- `internal/storage/postgres.go` — PostgreSQL connection pool
- `internal/storage/valkey.go` — Valkey client

Both should:
- Accept connection string from config
- Provide `Ping()` method for health checks
- Return errors (don't panic)
- Use `database/sql` + `lib/pq` for PostgreSQL
- Use `github.com/valkey-io/valkey-go` for Valkey

### 8. Update Health Endpoint
Update `internal/api/health.go` to:
- Return version from build info
- Check PostgreSQL and Valkey connectivity
- Return dependency status in response

```json
{
  "service": "acumius",
  "status": "ok",
  "version": "0.1.0",
  "dependencies": {
    "postgresql": "connected",
    "valkey": "connected"
  }
}
```

### Acceptance Criteria
- [ ] `docker-compose up` starts all 4 services
- [ ] `curl http://localhost:8080/health` returns full health response
- [ ] `curl http://localhost:8080/ready` returns readiness with dependency status
- [ ] PostgreSQL is accessible on port 5432
- [ ] Valkey is accessible on port 6379
- [ ] `make check` still passes
- [ ] CI passes on PR
- [ ] New contributor can `git clone`, `cp .env.example .env`, `make up`, and verify in < 5 minutes

## Constraints
- Do NOT add business logic (memory, trust, policy) — this is infrastructure only
- Do NOT modify existing scaffold files beyond what's specified
- Keep dependencies minimal
- All errors must be handled, no panics
- Follow existing code style (see `internal/api/health.go`)
