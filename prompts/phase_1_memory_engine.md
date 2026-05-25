# AntiGravity Prompt: Phase 1 — Memory Engine

## Context
You are implementing the Memory Engine for Acumius. Phase 0 is complete: the Go service scaffold, Docker Compose, CI, and storage connections are all working. Your task is to build the core memory subsystem that stores, retrieves, searches, and routes memories across 6 types.

## Current State
- `cmd/acumius/main.go` — entrypoint with HTTP server
- `internal/config/config.go` — env var config
- `internal/api/router.go` — HTTP router
- `internal/api/health.go` — health/ready endpoints
- `internal/storage/postgres.go` — PostgreSQL connection pool
- `internal/storage/valkey.go` — Valkey client
- `docker-compose.yml` — PostgreSQL + pgvector + Valkey + governance-ui
- `migrations/000001_init.up.sql` — empty (your task to populate)
- `migrations/000001_init.down.sql` — empty

## Your Task

### 1. Memory Type Definitions
Create `internal/memory/types.go`:

```go
package memory

import (
    "encoding/json"
    "time"
    "github.com/google/uuid"
)

type MemoryType string

const (
    Working     MemoryType = "working"
    Episodic    MemoryType = "episodic"
    Semantic    MemoryType = "semantic"
    Procedural  MemoryType = "procedural"
    Declarative MemoryType = "declarative"
    Feedback    MemoryType = "feedback"
)

type Memory struct {
    ID           uuid.UUID       `json:"id"`
    Type         MemoryType      `json:"type"`
    Namespace    string          `json:"namespace"`
    AgentDID     string          `json:"agent_did"`
    Content      json.RawMessage `json:"content"`
    Embedding    []float32       `json:"-"`
    Metadata     Metadata        `json:"metadata"`
    ValidFrom    *time.Time      `json:"valid_from,omitempty"`
    ValidUntil   *time.Time      `json:"valid_until,omitempty"`
    DistilledFrom *uuid.UUID     `json:"distilled_from,omitempty"`
    CreatedAt    time.Time       `json:"created_at"`
    UpdatedAt    time.Time       `json:"updated_at"`
    DeletedAt    *time.Time      `json:"deleted_at,omitempty"`
}

type Metadata struct {
    Source      string            `json:"source,omitempty"`
    Confidence  float64           `json:"confidence,omitempty"`
    Tags        []string          `json:"tags,omitempty"`
    PII         []PIIField        `json:"pii,omitempty"`
    Attestation *Attestation      `json:"attestation,omitempty"`
}

type PIIField struct {
    Type  string `json:"type"`
    Start int    `json:"start"`
    End   int    `json:"end"`
}

type Attestation struct {
    AgentDID  string    `json:"agent_did"`
    MemoryID  uuid.UUID `json:"memory_id"`
    Claim     string    `json:"claim"`
    Signature []byte    `json:"signature"`
    Timestamp time.Time `json:"timestamp"`
}
```

### 2. Database Migration
Populate `migrations/000001_init.up.sql` with the full schema from `docs/schema.md`.

Key tables: `agents`, `memories`, `namespace_acl`, `attestations`, `audit_log` (partitioned), `reputation_events`, `policies`, `pii_registry`.

Populate `migrations/000001_init.down.sql` with DROP statements (reverse order).

### 3. Store Interface
Create `internal/memory/store.go`:

```go
package memory

import "context"

type Store interface {
    Store(ctx context.Context, m *Memory) error
    Retrieve(ctx context.Context, id uuid.UUID) (*Memory, error)
    Search(ctx context.Context, query SearchQuery) (*SearchResult, error)
    ListByNamespace(ctx context.Context, namespace string, opts ListOptions) ([]Memory, error)
    Delete(ctx context.Context, id uuid.UUID) error
    RedactPII(ctx context.Context, namespace string, types []string) error
}

type SearchQuery struct {
    Query      string
    Types      []MemoryType
    Namespaces []string
    Limit      int
    Offset     int
    Filters    map[string]interface{}
}

type SearchResult struct {
    Results []Memory
    Total   int
    Limit   int
    Offset  int
}
```

### 4. PostgreSQL Store Implementation
Create `internal/memory/postgres_store.go`:

Implement `Store` interface using `database/sql` + `lib/pq`.

Key methods:
- `Store()` — INSERT into `memories`, generate embedding if type is semantic
- `Retrieve()` — SELECT by ID, check `deleted_at IS NULL`
- `Search()` — Hybrid search: pgvector cosine similarity + full-text tsvector + metadata filters. Use RRF (Reciprocal Rank Fusion) for merging.
- `ListByNamespace()` — SELECT with pagination
- `Delete()` — UPDATE `deleted_at = NOW()` (soft delete)
- `RedactPII()` — UPDATE content to redact PII fields

For embedding generation, use a placeholder function that returns random vectors for now. Phase 3 will integrate a real embedding model.

### 5. Valkey Store Implementation
Create `internal/memory/valkey_store.go`:

Implement `Store` interface for Working Memory only.

Key methods:
- `Store()` — SET with TTL (default 24h), key format: `acumius:working:{namespace}:{agent_did}:{id}`
- `Retrieve()` — GET by key
- `Search()` — SCAN or use sorted set for latest
- `ListByNamespace()` — SCAN pattern `acumius:working:{namespace}:*`
- `Delete()` — DEL key
- `RedactPII()` — Not applicable for Working Memory

### 6. Storage Router
Create `internal/memory/router.go`:

```go
package memory

type Router struct {
    postgres Store
    valkey   Store
}

func (r *Router) Route(m *Memory) Store {
    if m.Type == Working {
        return r.valkey
    }
    return r.postgres
}
```

### 7. REST API Handlers
Create `internal/api/memory_handlers.go`:

Implement handlers for:
- `POST /v1/memory` — parse request, route to correct store, return 201
- `GET /v1/memory/{id}` — retrieve, check namespace permissions (placeholder for now), return 200
- `POST /v1/memory/search` — parse SearchQuery, execute hybrid search, return 200
- `GET /v1/memory/namespace/{ns}` — list by namespace with pagination
- `DELETE /v1/memory/{id}` — soft delete
- `POST /v1/memory/redact` — bulk redact PII

Use `github.com/go-chi/chi/v5` for routing (add to go.mod).

### 8. Namespace ACL (Basic)
Create `internal/memory/acl.go`:

Simple in-memory map for v0.1:
```go
type ACL struct {
    mu sync.RWMutex
    rules map[string]map[string]Permission // namespace -> agent_did -> permission
}
```

Methods: `Grant()`, `Revoke()`, `Check()`

For v0.1, agents have full access to their own namespace (`"self"`) and can be granted access to shared namespaces.

### 9. Tests
Write tests for:
- Store interface (use `testcontainers-go` for PostgreSQL/Valkey in tests)
- Router logic
- API handlers (use `httptest`)
- ACL logic

Target: > 70% coverage on `internal/memory/`

### 10. Integration
Wire everything into `cmd/acumius/main.go`:
- Initialize PostgreSQL and Valkey stores
- Create router
- Register memory handlers
- Run migrations on startup

### Acceptance Criteria
- [ ] `POST /v1/memory` stores all 6 memory types
- [ ] `GET /v1/memory/{id}` retrieves by ID
- [ ] `POST /v1/memory/search` returns hybrid search results
- [ ] Working Memory auto-expires after TTL
- [ ] Semantic Memory has embedding (placeholder)
- [ ] Soft delete works (deleted_at set, not hard deleted)
- [ ] Namespace isolation prevents unauthorized reads
- [ ] `make test` passes with > 70% coverage on memory package
- [ ] `docker-compose up` runs migrations automatically

## Constraints
- Do NOT implement Trust Layer yet — use placeholder agent_did validation
- Do NOT implement Policy Engine yet — skip policy checks
- Do NOT implement Governance UI — this is backend only
- Use placeholder embedding generation (random vectors)
- Handle all errors, no panics
- Follow existing code style
