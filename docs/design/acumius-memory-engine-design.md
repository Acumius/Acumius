# Acumius Phase 1 — Memory Engine Design Guide

> **Version:** 1.0  
> **Status:** Ready for Implementation  
> **Prerequisite:** Phase 2 (Trust Layer) Complete  
> **Last Updated:** 2026-05-26

This document provides complete Go struct outlines, interface designs, SQL schema, and integration patterns for implementing the Acumius Memory Engine. It is designed to be fed directly to AntiGravity (Claude Code) or used as a reference by the core team.

---

## Table of Contents

1. [Memory Type Definitions](#1-memory-type-definitions)
2. [Attestation Integration](#2-attestation-integration)
3. [Storage Router Design](#3-storage-router-design)
4. [PostgreSQL Schema](#4-postgresql-schema)
5. [Hybrid Search Design](#5-hybrid-search-design)
6. [Trust Integration](#6-trust-integration)
7. [Complete File Structure](#7-complete-file-structure)
8. [AntiGravity Prompt](#8-antigravity-prompt)

---

## 1. Memory Type Definitions

### 1.1 Core Memory Struct

```go
// internal/memory/types.go
package memory

import (
    "encoding/json"
    "time"

    "github.com/google/uuid"
)

// MemoryType represents the six cognitive memory types in Acumius.
type MemoryType string

const (
    Working     MemoryType = "working"
    Episodic    MemoryType = "episodic"
    Semantic    MemoryType = "semantic"
    Procedural  MemoryType = "procedural"
    Declarative MemoryType = "declarative"
    Feedback    MemoryType = "feedback"
)

// IsPersistent returns true for memory types stored in PostgreSQL.
func (mt MemoryType) IsPersistent() bool {
    return mt != Working
}

// DefaultTTL returns the default time-to-live for volatile memory types.
func (mt MemoryType) DefaultTTL() time.Duration {
    if mt == Working {
        return 24 * time.Hour
    }
    return 0 // persistent
}

// Memory is the unified struct for all six memory types.
type Memory struct {
    ID            uuid.UUID       `json:"id"`
    Type          MemoryType      `json:"type"`
    Namespace     string          `json:"namespace"`
    AgentDID      string          `json:"agent_did"`
    Content       json.RawMessage `json:"content"`
    Embedding     []float32       `json:"-"`              // excluded from JSON; stored in pgvector
    Metadata      Metadata        `json:"metadata"`
    ValidFrom     *time.Time      `json:"valid_from,omitempty"`
    ValidUntil    *time.Time      `json:"valid_until,omitempty"`
    DistilledFrom *uuid.UUID      `json:"distilled_from,omitempty"`
    CreatedAt     time.Time       `json:"created_at"`
    UpdatedAt     time.Time       `json:"updated_at"`
    DeletedAt     *time.Time      `json:"deleted_at,omitempty"`
}

// Metadata holds optional structured metadata and Phase 2 attestations.
type Metadata struct {
    Source      string            `json:"source,omitempty"`
    Confidence  float64           `json:"confidence,omitempty"`
    Tags        []string          `json:"tags,omitempty"`
    PII         []PIIField        `json:"pii,omitempty"`
    Attestations []AttestationRef  `json:"attestations,omitempty"` // Phase 2 integration
}

// PIIField marks personally identifiable information for GDPR redaction.
type PIIField struct {
    Type  string `json:"type"`  // e.g., "email", "phone", "ssn"
    Start int    `json:"start"` // byte offset in Content
    End   int    `json:"end"`
}

// AttestationRef is a lightweight reference to a full Attestation stored in the trust layer.
// This avoids duplicating cryptographic data inside the memory row.
type AttestationRef struct {
    AttestationID uuid.UUID `json:"attestation_id"`
    AgentDID      string    `json:"agent_did"`
    Claim         string    `json:"claim"`
    Timestamp     time.Time `json:"timestamp"`
}

// StoreRequest is the DTO for storing a new memory.
type StoreRequest struct {
    Type       MemoryType      `json:"type" validate:"required,oneof=working episodic semantic procedural declarative feedback"`
    Namespace  string          `json:"namespace" validate:"required,min=1,max=255"`
    Content    json.RawMessage `json:"content" validate:"required"`
    Metadata   Metadata        `json:"metadata"`
    ValidFrom  *time.Time      `json:"valid_from,omitempty"`
    ValidUntil *time.Time      `json:"valid_until,omitempty"`
}

// SearchQuery defines parameters for hybrid semantic + keyword search.
type SearchQuery struct {
    Query      string       `json:"query" validate:"required"`
    Types      []MemoryType `json:"types,omitempty"`
    Namespaces []string     `json:"namespaces,omitempty"`
    AgentDID   string       `json:"agent_did,omitempty"`
    Limit      int          `json:"limit" validate:"min=1,max=100"`
    Offset     int          `json:"offset" validate:"min=0"`
    Filters    FilterSet    `json:"filters,omitempty"`
}

// FilterSet provides optional metadata filters.
type FilterSet struct {
    Tags           []string   `json:"tags,omitempty"`
    ConfidenceMin  float64    `json:"confidence_min,omitempty"`
    ConfidenceMax  float64    `json:"confidence_max,omitempty"`
    CreatedAfter   *time.Time `json:"created_after,omitempty"`
    CreatedBefore  *time.Time `json:"created_before,omitempty"`
    HasAttestation bool       `json:"has_attestation,omitempty"`
}

// SearchResult is the DTO returned from a search operation.
type SearchResult struct {
    Results []Memory `json:"results"`
    Total   int      `json:"total"`
    Limit   int      `json:"limit"`
    Offset  int      `json:"offset"`
}
```

### 1.2 Design Rationale

| Decision | Rationale |
|----------|-----------|
| `json.RawMessage` for `Content` | Memory content is schema-less by design; agents store arbitrary JSON. The Memory Engine does not validate content structure — that is the agent's responsibility. |
| `[]float32` for `Embedding` | pgvector expects `float32` (or `real[]`). We exclude it from JSON serialization because embeddings are large and should not transit over the REST API unless explicitly requested. |
| `AttestationRef` not full `Attestation` | Full attestations contain Ed25519 signatures and public keys (200+ bytes). Storing them inline would bloat the memory table. Instead, we store lightweight references and join with the `attestations` table when needed. |
| `DeletedAt` soft delete | GDPR compliance requires recoverability. Hard deletes are only performed by the GDPR auto-expiry worker after the retention period. |
| `ValidFrom` / `ValidUntil` | Temporal validity windows enable "what was true in January?" queries. This is a key differentiator from Mem0 and Letta. |

---

## 2. Attestation Integration

### 2.1 How Attestations Flow

```
Agent C (Verifier)                    Acumius
     │                                   │
     │ POST /v1/memory/{id}/attest     │
     │ { claim: "Verified against SEC" } │
     │                                   │
     │                                   ▼
     │                          ┌─────────────────┐
     │                          │  Trust Layer    │
     │                          │  • Sign claim   │
     │                          │  • Store in     │
     │                          │    attestations │
     │                          │    table        │
     │                          └────────┬────────┘
     │                                   │
     │                                   ▼
     │                          ┌─────────────────┐
     │                          │  Memory Engine  │
     │                          │  • Add          │
     │                          │    AttestationRef│
     │                          │    to metadata  │
     │                          │  • Update       │
     │                          │    memories row  │
     │                          └─────────────────┘
```

### 2.2 AttestationRef vs Full Attestation

```go
// internal/memory/types.go — lightweight reference stored in memory metadata
type AttestationRef struct {
    AttestationID uuid.UUID `json:"attestation_id"`  // FK to attestations.id
    AgentDID      string    `json:"agent_did"`       // who attested
    Claim         string    `json:"claim"`           // human-readable claim
    Timestamp     time.Time `json:"timestamp"`       // when
}

// internal/trust/attestation.go — full cryptographic record (Phase 2)
type Attestation struct {
    ID        uuid.UUID `json:"id"`
    AgentDID  string    `json:"agent_did"`
    MemoryID  uuid.UUID `json:"memory_id"`
    Claim     string    `json:"claim"`
    Signature []byte    `json:"signature"`     // Ed25519 signature
    Timestamp time.Time `json:"timestamp"`
}
```

### 2.3 Retrieval with Attestations

When retrieving a memory, the Memory Engine should optionally join with the `attestations` table:

```go
// internal/memory/postgres_store.go
func (s *PostgresStore) Retrieve(ctx context.Context, id uuid.UUID, opts RetrieveOpts) (*Memory, error) {
    // ... base query ...

    if opts.IncludeAttestations {
        // Join with attestations table
        attestations, err := s.trustLayer.GetAttestationsForMemory(ctx, id)
        if err != nil {
            return nil, fmt.Errorf("fetch attestations: %w", err)
        }

        // Convert full attestations to lightweight refs for the response
        mem.Metadata.Attestations = make([]AttestationRef, len(attestations))
        for i, att := range attestations {
            mem.Metadata.Attestations[i] = AttestationRef{
                AttestationID: att.ID,
                AgentDID:      att.AgentDID,
                Claim:         att.Claim,
                Timestamp:     att.Timestamp,
            }
        }
    }

    return mem, nil
}
```

---

## 3. Storage Router Design

### 3.1 Store Interface

```go
// internal/memory/store.go
package memory

import (
    "context"
    "time"

    "github.com/google/uuid"
)

// Store is the abstract interface for all memory backends.
type Store interface {
    // Write operations
    Store(ctx context.Context, m *Memory) error
    Update(ctx context.Context, m *Memory) error
    Delete(ctx context.Context, id uuid.UUID) error

    // Read operations
    Retrieve(ctx context.Context, id uuid.UUID, opts RetrieveOpts) (*Memory, error)
    Search(ctx context.Context, query SearchQuery) (*SearchResult, error)
    ListByNamespace(ctx context.Context, namespace string, opts ListOpts) (*SearchResult, error)

    // Lifecycle
    RedactPII(ctx context.Context, namespace string, piiTypes []string) (int, error)
    Expire(ctx context.Context, before time.Time) (int, error)
    Ping(ctx context.Context) error
}

// RetrieveOpts controls optional data loading.
type RetrieveOpts struct {
    IncludeAttestations bool
    IncludeEmbedding    bool // rarely needed over API
}

// ListOpts controls pagination for list operations.
type ListOpts struct {
    Types  []MemoryType
    Limit  int
    Offset int
}
```

### 3.2 Router Implementation

```go
// internal/memory/router.go
package memory

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
)

// Router dynamically routes memory operations to the correct backend.
type Router struct {
    postgres Store // PostgreSQL + pgvector for persistent memories
    valkey   Store // Valkey for Working Memory with TTL
}

// NewRouter creates a storage router with the given backends.
func NewRouter(postgres, valkey Store) *Router {
    return &Router{
        postgres: postgres,
        valkey:   valkey,
    }
}

// backend selects the appropriate store for a memory type.
func (r *Router) backend(mt MemoryType) Store {
    if mt == Working {
        return r.valkey
    }
    return r.postgres
}

// Store routes the write to the correct backend.
func (r *Router) Store(ctx context.Context, m *Memory) error {
    if err := m.Type.Validate(); err != nil {
        return fmt.Errorf("invalid memory type: %w", err)
    }

    store := r.backend(m.Type)

    // Working memory gets TTL applied
    if m.Type == Working {
        // TTL is handled by the Valkey store implementation
        // The store sets EXPIRE on the key
    }

    return store.Store(ctx, m)
}

// Update routes updates to the correct backend.
func (r *Router) Update(ctx context.Context, m *Memory) error {
    store := r.backend(m.Type)
    return store.Update(ctx, m)
}

// Delete routes deletes to the correct backend.
func (r *Router) Delete(ctx context.Context, id uuid.UUID, memType MemoryType) error {
    store := r.backend(memType)
    return store.Delete(ctx, id)
}

// Retrieve routes retrieval. For Working memory, goes to Valkey.
// For other types, goes to PostgreSQL.
func (r *Router) Retrieve(ctx context.Context, id uuid.UUID, memType MemoryType, opts RetrieveOpts) (*Memory, error) {
    store := r.backend(memType)
    return store.Retrieve(ctx, id, opts)
}

// Search routes search. Hybrid search only applies to PostgreSQL-backed memories.
// Working memory search uses Valkey SCAN or sorted sets.
func (r *Router) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
    // If query includes Working memory, we need to search Valkey too
    hasWorking := false
    hasPersistent := false

    for _, t := range query.Types {
        if t == Working {
            hasWorking = true
        } else {
            hasPersistent = true
        }
    }

    // If no types specified, search all (both backends)
    if len(query.Types) == 0 {
        hasWorking = true
        hasPersistent = true
    }

    var results []Memory
    var total int

    if hasPersistent {
        res, err := r.postgres.Search(ctx, query)
        if err != nil {
            return nil, fmt.Errorf("postgres search: %w", err)
        }
        results = append(results, res.Results...)
        total += res.Total
    }

    if hasWorking {
        res, err := r.valkey.Search(ctx, query)
        if err != nil {
            return nil, fmt.Errorf("valkey search: %w", err)
        }
        results = append(results, res.Results...)
        total += res.Total
    }

    // Merge and deduplicate (unlikely but possible with UUID collisions)
    // Sort by recency
    // ...

    return &SearchResult{
        Results: results,
        Total:   total,
        Limit:   query.Limit,
        Offset:  query.Offset,
    }, nil
}

// ListByNamespace lists memories in a namespace across backends.
func (r *Router) ListByNamespace(ctx context.Context, namespace string, opts ListOpts) (*SearchResult, error) {
    // Similar to Search: query both backends if needed
    // ...
}

// Cross-backend operations

// RedactPII redacts PII from all backends for a namespace.
func (r *Router) RedactPII(ctx context.Context, namespace string, piiTypes []string) (int, error) {
    pgCount, err := r.postgres.RedactPII(ctx, namespace, piiTypes)
    if err != nil {
        return 0, fmt.Errorf("postgres redact: %w", err)
    }

    vkCount, err := r.valkey.RedactPII(ctx, namespace, piiTypes)
    if err != nil {
        return 0, fmt.Errorf("valkey redact: %w", err)
    }

    return pgCount + vkCount, nil
}

// Expire removes memories past their ValidUntil from all backends.
func (r *Router) Expire(ctx context.Context, before time.Time) (int, error) {
    pgCount, err := r.postgres.Expire(ctx, before)
    if err != nil {
        return 0, fmt.Errorf("postgres expire: %w", err)
    }

    vkCount, err := r.valkey.Expire(ctx, before)
    if err != nil {
        return 0, fmt.Errorf("valkey expire: %w", err)
    }

    return pgCount + vkCount, nil
}

// Ping checks connectivity to all backends.
func (r *Router) Ping(ctx context.Context) error {
    if err := r.postgres.Ping(ctx); err != nil {
        return fmt.Errorf("postgres ping: %w", err)
    }
    if err := r.valkey.Ping(ctx); err != nil {
        return fmt.Errorf("valkey ping: %w", err)
    }
    return nil
}
```

### 3.3 Valkey Store Implementation

```go
// internal/memory/valkey_store.go
package memory

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/valkey-io/valkey-go"
)

// ValkeyStore implements Store for Working Memory with TTL.
type ValkeyStore struct {
    client valkey.Client
    ttl    time.Duration // default TTL for Working Memory
}

// NewValkeyStore creates a Valkey-backed store.
func NewValkeyStore(client valkey.Client, defaultTTL time.Duration) *ValkeyStore {
    if defaultTTL == 0 {
        defaultTTL = 24 * time.Hour
    }
    return &ValkeyStore{
        client: client,
        ttl:    defaultTTL,
    }
}

// key generates a Valkey key for a memory.
func (s *ValkeyStore) key(namespace, agentDID string, id uuid.UUID) string {
    return fmt.Sprintf("acumius:working:%s:%s:%s", namespace, agentDID, id.String())
}

// latestKey generates a sorted set key for tracking latest memories.
func (s *ValkeyStore) latestKey(namespace, agentDID string) string {
    return fmt.Sprintf("acumius:working:%s:%s:latest", namespace, agentDID)
}

// Store saves Working Memory to Valkey with TTL.
func (s *ValkeyStore) Store(ctx context.Context, m *Memory) error {
    if m.Type != Working {
        return fmt.Errorf("valkey store only supports Working memory, got %s", m.Type)
    }

    data, err := json.Marshal(m)
    if err != nil {
        return fmt.Errorf("marshal memory: %w", err)
    }

    key := s.key(m.Namespace, m.AgentDID, m.ID)

    // SET with EX (expiry in seconds)
    ttlSeconds := int(s.ttl.Seconds())
    if m.ValidUntil != nil {
        // Use ValidUntil if provided
        ttlSeconds = int(time.Until(*m.ValidUntil).Seconds())
    }

    err = s.client.Do(ctx, s.client.B().Set().Key(key).Value(string(data)).ExSeconds(ttlSeconds).Build()).Error()
    if err != nil {
        return fmt.Errorf("valkey set: %w", err)
    }

    // Add to sorted set for listing (score = timestamp)
    latestKey := s.latestKey(m.Namespace, m.AgentDID)
    err = s.client.Do(ctx, s.client.B().Zadd().Key(latestKey).ScoreMember().
        Score(float64(m.CreatedAt.Unix())).Member(m.ID.String()).Build()).Error()
    if err != nil {
        return fmt.Errorf("valkey zadd: %w", err)
    }

    // Trim sorted set to last 1000 entries (prevent unbounded growth)
    err = s.client.Do(ctx, s.client.B().Zremrangebyrank().Key(latestKey).Start(0).Stop(-1001).Build()).Error()
    if err != nil {
        // Non-fatal: log but don't fail the store
        // logger.Warn("failed to trim working memory sorted set", "error", err)
    }

    return nil
}

// Retrieve fetches Working Memory by ID.
func (s *ValkeyStore) Retrieve(ctx context.Context, id uuid.UUID, opts RetrieveOpts) (*Memory, error) {
    // We need namespace and agentDID to construct the key
    // For Valkey, we can use SCAN or maintain an index
    // Simplified: use a secondary index key

    indexKey := fmt.Sprintf("acumius:working:index:%s", id.String())

    // Get the actual key from index
    actualKey, err := s.client.Do(ctx, s.client.B().Get().Key(indexKey).Build()).ToString()
    if err != nil {
        if valkey.IsValkeyNil(err) {
            return nil, fmt.Errorf("memory not found: %s", id)
        }
        return nil, fmt.Errorf("valkey get index: %w", err)
    }

    data, err := s.client.Do(ctx, s.client.B().Get().Key(actualKey).Build()).ToString()
    if err != nil {
        if valkey.IsValkeyNil(err) {
            return nil, fmt.Errorf("memory expired or deleted: %s", id)
        }
        return nil, fmt.Errorf("valkey get: %w", err)
    }

    var m Memory
    if err := json.Unmarshal([]byte(data), &m); err != nil {
        return nil, fmt.Errorf("unmarshal memory: %w", err)
    }

    return &m, nil
}

// Search performs keyword search on Working Memory.
// Note: Valkey does not support semantic search. This is keyword-only.
func (s *ValkeyStore) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
    // Use SCAN to find matching keys, then filter in-memory
    // This is acceptable for Working Memory (small, ephemeral dataset)

    pattern := "acumius:working:*:*"
    if len(query.Namespaces) == 1 {
        pattern = fmt.Sprintf("acumius:working:%s:*", query.Namespaces[0])
    }

    var results []Memory
    var cursor uint64

    for {
        scanResult, err := s.client.Do(ctx, s.client.B().Scan().Cursor(cursor).Match(pattern).Count(100).Build()).AsScanEntry()
        if err != nil {
            return nil, fmt.Errorf("valkey scan: %w", err)
        }

        for _, key := range scanResult.Elements {
            data, err := s.client.Do(ctx, s.client.B().Get().Key(key).Build()).ToString()
            if err != nil {
                continue // skip expired/deleted
            }

            var m Memory
            if err := json.Unmarshal([]byte(data), &m); err != nil {
                continue
            }

            // Filter by query string (simple substring match in content)
            if !strings.Contains(string(m.Content), query.Query) {
                continue
            }

            // Filter by namespace
            if len(query.Namespaces) > 0 && !contains(query.Namespaces, m.Namespace) {
                continue
            }

            results = append(results, m)

            if len(results) >= query.Limit {
                break
            }
        }

        cursor = scanResult.Cursor
        if cursor == 0 || len(results) >= query.Limit {
            break
        }
    }

    return &SearchResult{
        Results: results,
        Total:   len(results), // approximate; full count requires separate SCAN
        Limit:   query.Limit,
        Offset:  query.Offset,
    }, nil
}

// ListByNamespace lists Working Memory in a namespace.
func (s *ValkeyStore) ListByNamespace(ctx context.Context, namespace string, opts ListOpts) (*SearchResult, error) {
    // Use SCAN with namespace pattern
    pattern := fmt.Sprintf("acumius:working:%s:*", namespace)
    // ... similar to Search ...
}

// Delete removes Working Memory.
func (s *ValkeyStore) Delete(ctx context.Context, id uuid.UUID) error {
    indexKey := fmt.Sprintf("acumius:working:index:%s", id.String())

    // Get actual key
    actualKey, err := s.client.Do(ctx, s.client.B().Get().Key(indexKey).Build()).ToString()
    if err != nil {
        if valkey.IsValkeyNil(err) {
            return fmt.Errorf("memory not found: %s", id)
        }
        return fmt.Errorf("valkey get index: %w", err)
    }

    // Delete both keys
    err = s.client.Do(ctx, s.client.B().Del().Key(indexKey, actualKey).Build()).Error()
    if err != nil {
        return fmt.Errorf("valkey del: %w", err)
    }

    return nil
}

// Update updates Working Memory (rarely used; usually just Store with new TTL).
func (s *ValkeyStore) Update(ctx context.Context, m *Memory) error {
    // Working Memory is typically immutable after creation
    // But we support updates for metadata changes
    return s.Store(ctx, m) // overwrite
}

// RedactPII is a no-op for Working Memory in v0.1.
// Working Memory is ephemeral; PII should not be stored here.
func (s *ValkeyStore) RedactPII(ctx context.Context, namespace string, piiTypes []string) (int, error) {
    return 0, nil
}

// Expire removes Working Memory past ValidUntil.
func (s *ValkeyStore) Expire(ctx context.Context, before time.Time) (int, error) {
    // Valkey auto-expires keys with TTL, so this is mostly a no-op
    // But we can clean up orphaned index keys
    return 0, nil
}

// Ping checks Valkey connectivity.
func (s *ValkeyStore) Ping(ctx context.Context) error {
    return s.client.Do(ctx, s.client.B().Ping().Build()).Error()
}

// Helper
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

---

## 4. PostgreSQL Schema

### 4.1 Migration File

```sql
-- migrations/000001_init.up.sql

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- ============================================
-- AGENTS (populated by Phase 2 Trust Layer)
-- ============================================
CREATE TABLE IF NOT EXISTS agents (
    did TEXT PRIMARY KEY,
    public_key BYTEA NOT NULL,
    name TEXT NOT NULL,
    capabilities TEXT[] DEFAULT '{}',
    reputation_score INT DEFAULT 500,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'revoked')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agents_reputation ON agents(reputation_score);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);

-- ============================================
-- MEMORIES (unified table for 5 persistent types)
-- ============================================
CREATE TABLE IF NOT EXISTS memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type TEXT NOT NULL CHECK (type IN ('episodic', 'semantic', 'procedural', 'declarative', 'feedback')),
    namespace TEXT NOT NULL,
    agent_did TEXT NOT NULL REFERENCES agents(did),
    content JSONB NOT NULL,
    embedding VECTOR(1536),                    -- pgvector for semantic search
    metadata JSONB DEFAULT '{}',
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    distilled_from UUID REFERENCES memories(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ                    -- soft delete for GDPR
);

-- Full-text search index on content
CREATE INDEX IF NOT EXISTS idx_memories_fts 
    ON memories USING GIN(to_tsvector('english', content::text));

-- Vector index for semantic search (IVFFlat for balance of speed/recall)
CREATE INDEX IF NOT EXISTS idx_memories_embedding 
    ON memories USING ivfflat (embedding vector_cosine_ops) 
    WITH (lists = 100);                       -- tune based on dataset size

-- Common query indexes
CREATE INDEX IF NOT EXISTS idx_memories_namespace ON memories(namespace);
CREATE INDEX IF NOT EXISTS idx_memories_agent ON memories(agent_did);
CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at);

-- Composite index for filtered queries
CREATE INDEX IF NOT EXISTS idx_memories_ns_type_created 
    ON memories(namespace, type, created_at) 
    WHERE deleted_at IS NULL;

-- Partial index for active (non-deleted) memories
CREATE INDEX IF NOT EXISTS idx_memories_active 
    ON memories(namespace, type, agent_did) 
    WHERE deleted_at IS NULL;

-- Index for temporal queries
CREATE INDEX IF NOT EXISTS idx_memories_valid 
    ON memories(valid_from, valid_until) 
    WHERE deleted_at IS NULL;

-- ============================================
-- NAMESPACE ACL
-- ============================================
CREATE TABLE IF NOT EXISTS namespace_acl (
    namespace TEXT NOT NULL,
    agent_did TEXT NOT NULL REFERENCES agents(did),
    permission TEXT NOT NULL CHECK (permission IN ('read', 'write', 'admin')),
    granted_by TEXT NOT NULL REFERENCES agents(did),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (namespace, agent_did, permission)
);

CREATE INDEX IF NOT EXISTS idx_acl_namespace ON namespace_acl(namespace);
CREATE INDEX IF NOT EXISTS idx_acl_agent ON namespace_acl(agent_did);

-- ============================================
-- ATTESTATIONS (Phase 2 Trust Layer)
-- ============================================
CREATE TABLE IF NOT EXISTS attestations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_did TEXT NOT NULL REFERENCES agents(did),
    memory_id UUID NOT NULL REFERENCES memories(id),
    claim TEXT NOT NULL,
    signature BYTEA NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_attestations_memory ON attestations(memory_id);
CREATE INDEX IF NOT EXISTS idx_attestations_agent ON attestations(agent_did);

-- ============================================
-- AUDIT LOG (append-only, partitioned by month)
-- ============================================
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    agent_did TEXT NOT NULL,
    action TEXT NOT NULL,
    resource TEXT NOT NULL,
    allowed BOOLEAN NOT NULL,
    policy_id TEXT,
    reason TEXT,
    metadata JSONB DEFAULT '{}'
) PARTITION BY RANGE (timestamp);

-- Create initial monthly partitions
CREATE TABLE IF NOT EXISTS audit_log_y2026m05 PARTITION OF audit_log
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE IF NOT EXISTS audit_log_y2026m06 PARTITION OF audit_log
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE INDEX IF NOT EXISTS idx_audit_agent ON audit_log(agent_did);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp);

-- ============================================
-- REPUTATION EVENTS (Phase 2 Trust Layer)
-- ============================================
CREATE TABLE IF NOT EXISTS reputation_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_did TEXT NOT NULL REFERENCES agents(did),
    event_type TEXT NOT NULL,
    delta INT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rep_events_agent ON reputation_events(agent_did);
CREATE INDEX IF NOT EXISTS idx_rep_events_created ON reputation_events(created_at);

-- ============================================
-- POLICIES (Phase 3 Policy Engine)
-- ============================================
CREATE TABLE IF NOT EXISTS policies (
    id TEXT PRIMARY KEY,
    agent_did TEXT REFERENCES agents(did),
    content JSONB NOT NULL,
    version TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_policies_agent ON policies(agent_did);

-- ============================================
-- PII REGISTRY (GDPR compliance)
-- ============================================
CREATE TABLE IF NOT EXISTS pii_registry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    memory_id UUID NOT NULL REFERENCES memories(id),
    pii_type TEXT NOT NULL,
    pii_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pii_memory ON pii_registry(memory_id);
CREATE INDEX IF NOT EXISTS idx_pii_hash ON pii_registry(pii_hash);
```

### 4.2 Schema Design Rationale

| Decision | Rationale |
|----------|-----------|
| Single `memories` table (not 6 tables) | Simpler queries, easier indexing, less migration overhead. Types are distinguished by the `type` column with CHECK constraint. |
| `VECTOR(1536)` | Standard embedding dimension for OpenAI `text-embedding-3-small`. Adjust if using a different model. |
| `IVFFlat` with 100 lists | Good balance for < 1M vectors. For larger datasets, switch to `hnsw` index (better recall, slower build). |
| `GIN` on `to_tsvector('english', content::text)` | Full-text search on JSON content. Uses English stemming. For multi-language support, add language-specific indexes. |
| Partial indexes (`WHERE deleted_at IS NULL`) | Exclude soft-deleted rows from all normal queries, improving performance and ensuring GDPR compliance. |
| `namespace` as TEXT (not normalized) | Namespaces are agent-defined strings. No need for a separate table; just index the column. |
| `content JSONB` (not JSON) | JSONB is binary, compressed, and supports GIN indexing. JSON is text storage — slower and larger. |

---

## 5. Hybrid Search Design

### 5.1 The Hybrid Search Algorithm

Acumius uses **Reciprocal Rank Fusion (RRF)** to combine three signals:

1. **Semantic similarity** — pgvector cosine similarity on embeddings
2. **Keyword relevance** — PostgreSQL full-text search (`ts_rank_cd`)
3. **Metadata filtering** — exact matches on tags, confidence, dates

```go
// internal/memory/search.go
package memory

import (
    "context"
    "fmt"
    "math"
    "sort"
    "strings"

    "github.com/google/uuid"
)

// HybridSearcher performs RRF-based hybrid search.
type HybridSearcher struct {
    db *sql.DB
    // embeddingFunc generates embeddings for queries
    // For v0.1, this is a placeholder that returns random vectors
    // For v0.2, this calls an external embedding model
    embeddingFunc func(ctx context.Context, text string) ([]float32, error)
}

// Search executes hybrid semantic + keyword search.
func (s *HybridSearcher) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
    // 1. Generate embedding for the query (if semantic types requested)
    var queryEmbedding []float32
    if needsSemantic(query.Types) {
        emb, err := s.embeddingFunc(ctx, query.Query)
        if err != nil {
            return nil, fmt.Errorf("generate embedding: %w", err)
        }
        queryEmbedding = emb
    }

    // 2. Semantic search (pgvector)
    semanticResults, err := s.semanticSearch(ctx, query, queryEmbedding)
    if err != nil {
        return nil, fmt.Errorf("semantic search: %w", err)
    }

    // 3. Keyword search (full-text)
    keywordResults, err := s.keywordSearch(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("keyword search: %w", err)
    }

    // 4. Merge with RRF
    merged := s.reciprocalRankFusion(semanticResults, keywordResults)

    // 5. Apply metadata filters
    filtered := s.applyFilters(merged, query.Filters)

    // 6. Paginate
    total := len(filtered)
    start := query.Offset
    end := min(start+query.Limit, total)

    if start > total {
        start = total
    }

    return &SearchResult{
        Results: filtered[start:end],
        Total:   total,
        Limit:   query.Limit,
        Offset:  query.Offset,
    }, nil
}

// semanticSearch queries pgvector for cosine similarity.
func (s *HybridSearcher) semanticSearch(ctx context.Context, query SearchQuery, embedding []float32) ([]scoredMemory, error) {
    sql := `
        SELECT 
            id, type, namespace, agent_did, content, metadata,
            created_at, updated_at,
            1 - (embedding <=> $1) as similarity  -- cosine similarity
        FROM memories
        WHERE deleted_at IS NULL
          AND type = ANY($2)
          AND namespace = ANY($3)
          AND embedding IS NOT NULL
        ORDER BY embedding <=> $1
        LIMIT $4
    `

    types := memoryTypesToStrings(query.Types)
    if len(types) == 0 {
        types = []string{"episodic", "semantic", "procedural", "declarative", "feedback"}
    }

    namespaces := query.Namespaces
    if len(namespaces) == 0 {
        namespaces = []string{"*"} // will be handled by SQL
    }

    rows, err := s.db.QueryContext(ctx, sql, pgvector.NewVector(embedding), pq.Array(types), pq.Array(namespaces), query.Limit*3)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []scoredMemory
    for rows.Next() {
        var m Memory
        var score float64
        err := rows.Scan(
            &m.ID, &m.Type, &m.Namespace, &m.AgentDID, &m.Content, &m.Metadata,
            &m.CreatedAt, &m.UpdatedAt, &score,
        )
        if err != nil {
            return nil, err
        }
        results = append(results, scoredMemory{memory: m, semanticScore: score})
    }

    return results, rows.Err()
}

// keywordSearch uses PostgreSQL full-text search.
func (s *HybridSearcher) keywordSearch(ctx context.Context, query SearchQuery) ([]scoredMemory, error) {
    // Convert query to tsquery
    tsquery := strings.Join(strings.Fields(query.Query), " | ")

    sql := `
        SELECT 
            id, type, namespace, agent_did, content, metadata,
            created_at, updated_at,
            ts_rank_cd(to_tsvector('english', content::text), plainto_tsquery('english', $1)) as rank
        FROM memories
        WHERE deleted_at IS NULL
          AND type = ANY($2)
          AND namespace = ANY($3)
          AND to_tsvector('english', content::text) @@ plainto_tsquery('english', $1)
        ORDER BY rank DESC
        LIMIT $4
    `

    types := memoryTypesToStrings(query.Types)
    if len(types) == 0 {
        types = []string{"episodic", "semantic", "procedural", "declarative", "feedback"}
    }

    rows, err := s.db.QueryContext(ctx, sql, query.Query, pq.Array(types), pq.Array(query.Namespaces), query.Limit*3)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []scoredMemory
    for rows.Next() {
        var m Memory
        var score float64
        err := rows.Scan(
            &m.ID, &m.Type, &m.Namespace, &m.AgentDID, &m.Content, &m.Metadata,
            &m.CreatedAt, &m.UpdatedAt, &score,
        )
        if err != nil {
            return nil, err
        }
        results = append(results, scoredMemory{memory: m, keywordScore: score})
    }

    return results, rows.Err()
}

// reciprocalRankFusion merges semantic and keyword results.
// RRF formula: score = sum(1 / (k + rank)) for each list
// k = 60 (standard RRF constant)
func (s *HybridSearcher) reciprocalRankFusion(semantic, keyword []scoredMemory) []scoredMemory {
    const k = 60.0

    // Build maps for O(1) lookup
    semanticRanks := make(map[uuid.UUID]int)
    for i, sm := range semantic {
        semanticRanks[sm.memory.ID] = i + 1
    }

    keywordRanks := make(map[uuid.UUID]int)
    for i, sm := range keyword {
        keywordRanks[sm.memory.ID] = i + 1
    }

    // Collect all unique IDs
    allIDs := make(map[uuid.UUID]struct{})
    for _, sm := range semantic {
        allIDs[sm.memory.ID] = struct{}{}
    }
    for _, sm := range keyword {
        allIDs[sm.memory.ID] = struct{}{}
    }

    // Calculate RRF scores
    var merged []scoredMemory
    for id := range allIDs {
        score := 0.0

        if rank, ok := semanticRanks[id]; ok {
            score += 1.0 / (k + float64(rank))
        }
        if rank, ok := keywordRanks[id]; ok {
            score += 1.0 / (k + float64(rank))
        }

        // Get the memory from either list
        var mem Memory
        for _, sm := range semantic {
            if sm.memory.ID == id {
                mem = sm.memory
                break
            }
        }
        if mem.ID == uuid.Nil {
            for _, sm := range keyword {
                if sm.memory.ID == id {
                    mem = sm.memory
                    break
                }
            }
        }

        merged = append(merged, scoredMemory{
            memory:    mem,
            rrfScore:  score,
        })
    }

    // Sort by RRF score descending
    sort.Slice(merged, func(i, j int) bool {
        return merged[i].rrfScore > merged[j].rrfScore
    })

    return merged
}

// applyFilters applies metadata filters to search results.
func (s *HybridSearcher) applyFilters(results []scoredMemory, filters FilterSet) []scoredMemory {
    var filtered []scoredMemory

    for _, sm := range results {
        m := sm.memory

        // Tag filter
        if len(filters.Tags) > 0 {
            hasTag := false
            for _, tag := range filters.Tags {
                if contains(m.Metadata.Tags, tag) {
                    hasTag = true
                    break
                }
            }
            if !hasTag {
                continue
            }
        }

        // Confidence filter
        if filters.ConfidenceMin > 0 && m.Metadata.Confidence < filters.ConfidenceMin {
            continue
        }
        if filters.ConfidenceMax > 0 && m.Metadata.Confidence > filters.ConfidenceMax {
            continue
        }

        // Date filters
        if filters.CreatedAfter != nil && m.CreatedAt.Before(*filters.CreatedAfter) {
            continue
        }
        if filters.CreatedBefore != nil && m.CreatedAt.After(*filters.CreatedBefore) {
            continue
        }

        // Attestation filter
        if filters.HasAttestation && len(m.Metadata.Attestations) == 0 {
            continue
        }

        filtered = append(filtered, sm)
    }

    return filtered
}

// scoredMemory is an internal struct for ranking.
type scoredMemory struct {
    memory        Memory
    semanticScore float64
    keywordScore  float64
    rrfScore      float64
}

// Helper functions
func needsSemantic(types []MemoryType) bool {
    if len(types) == 0 {
        return true
    }
    for _, t := range types {
        if t == Semantic || t == Episodic {
            return true
        }
    }
    return false
}

func memoryTypesToStrings(types []MemoryType) []string {
    result := make([]string, len(types))
    for i, t := range types {
        result[i] = string(t)
    }
    return result
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

### 5.2 Rationale for RRF

| Approach | Pros | Cons |
|----------|------|------|
| **RRF (chosen)** | No training needed, works with any number of signals, mathematically sound | Requires fetching more results than needed (over-fetch) |
| Linear weighted sum | Simple | Requires tuning weights per dataset |
| Learned ranker | Optimal | Requires labeled data, complex |

RRF is the industry standard for combining sparse (keyword) and dense (vector) retrieval. The constant `k=60` is well-established in research.

---

## 6. Trust Integration

### 6.1 How Memory Engine Calls Trust Layer

The Memory Engine should **not** import `internal/trust` directly. Instead, use an interface to avoid circular dependencies and enable testing.

```go
// internal/memory/trust_port.go
package memory

import (
    "context"

    "github.com/google/uuid"
)

// TrustPort is the interface the Memory Engine uses from the Trust Layer.
// This is implemented by the actual Trust Layer and also by mocks in tests.
type TrustPort interface {
    // VerifySignature checks if a memory write is signed by the claimed agent.
    VerifySignature(ctx context.Context, agentDID string, memoryID uuid.UUID, content []byte, signature []byte) error

    // GetAgentReputation returns the reputation score for an agent.
    GetAgentReputation(ctx context.Context, agentDID string) (int, error)

    // IsAgentActive checks if an agent is active (not suspended/revoked).
    IsAgentActive(ctx context.Context, agentDID string) (bool, error)

    // GetAttestationsForMemory returns all attestations for a memory.
    GetAttestationsForMemory(ctx context.Context, memoryID uuid.UUID) ([]AttestationRecord, error)

    // RecordEvent records a reputation event (e.g., successful memory store).
    RecordEvent(ctx context.Context, agentDID string, eventType string, delta int, description string) error
}

// AttestationRecord is the full attestation data from the trust layer.
type AttestationRecord struct {
    ID        uuid.UUID
    AgentDID  string
    MemoryID  uuid.UUID
    Claim     string
    Signature []byte
    Timestamp interface{}
}
```

### 6.2 Wiring in the Router

```go
// internal/memory/router.go (updated)
type Router struct {
    postgres Store
    valkey   Store
    trust    TrustPort  // Phase 2 integration
    policy   PolicyPort // Phase 3 integration (placeholder)
}

func NewRouter(postgres, valkey Store, trust TrustPort) *Router {
    return &Router{
        postgres: postgres,
        valkey:   valkey,
        trust:    trust,
    }
}

// StoreWithVerification stores memory after verifying the agent's signature.
func (r *Router) StoreWithVerification(ctx context.Context, req StoreRequest, signature []byte) (*Memory, error) {
    // 1. Check agent is active
    active, err := r.trust.IsAgentActive(ctx, req.AgentDID)
    if err != nil {
        return nil, fmt.Errorf("check agent status: %w", err)
    }
    if !active {
        return nil, fmt.Errorf("agent %s is not active", req.AgentDID)
    }

    // 2. Verify signature (optional for v0.1, required for v0.2)
    if signature != nil {
        content, _ := json.Marshal(req.Content)
        if err := r.trust.VerifySignature(ctx, req.AgentDID, uuid.Nil, content, signature); err != nil {
            return nil, fmt.Errorf("signature verification failed: %w", err)
        }
    }

    // 3. Check reputation (optional gate)
    rep, err := r.trust.GetAgentReputation(ctx, req.AgentDID)
    if err != nil {
        return nil, fmt.Errorf("get reputation: %w", err)
    }
    if rep < 100 {
        return nil, fmt.Errorf("agent reputation too low: %d (minimum 100)", rep)
    }

    // 4. Create memory
    mem := &Memory{
        ID:        uuid.New(),
        Type:      req.Type,
        Namespace: req.Namespace,
        AgentDID:  req.AgentDID,
        Content:   req.Content,
        Metadata:  req.Metadata,
        ValidFrom: req.ValidFrom,
        ValidUntil: req.ValidUntil,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // 5. Route to backend
    if err := r.Store(ctx, mem); err != nil {
        return nil, fmt.Errorf("store memory: %w", err)
    }

    // 6. Record reputation event
    if err := r.trust.RecordEvent(ctx, req.AgentDID, "memory_store", 5, "Successfully stored memory"); err != nil {
        // Non-fatal: log but don't fail
        // logger.Warn("failed to record reputation event", "error", err)
    }

    return mem, nil
}
```

### 6.3 Trust Layer Implementation of TrustPort

```go
// internal/trust/port_adapter.go
package trust

import (
    "context"
    "fmt"

    "github.com/Acumius/Acumius/internal/memory"
    "github.com/google/uuid"
)

// PortAdapter implements memory.TrustPort using the actual Trust Layer.
type PortAdapter struct {
    identity   *IdentityService
    reputation *ReputationEngine
    attestation *AttestationService
}

func NewPortAdapter(identity *IdentityService, reputation *ReputationEngine, attestation *AttestationService) *PortAdapter {
    return &PortAdapter{
        identity:    identity,
        reputation:  reputation,
        attestation: attestation,
    }
}

func (a *PortAdapter) VerifySignature(ctx context.Context, agentDID string, memoryID uuid.UUID, content []byte, signature []byte) error {
    agent, err := a.identity.Get(ctx, agentDID)
    if err != nil {
        return fmt.Errorf("get agent: %w", err)
    }

    // Verify Ed25519 signature
    if !ed25519.Verify(agent.PublicKey, content, signature) {
        return fmt.Errorf("invalid signature")
    }

    return nil
}

func (a *PortAdapter) GetAgentReputation(ctx context.Context, agentDID string) (int, error) {
    return a.reputation.CalculateScore(ctx, agentDID)
}

func (a *PortAdapter) IsAgentActive(ctx context.Context, agentDID string) (bool, error) {
    agent, err := a.identity.Get(ctx, agentDID)
    if err != nil {
        return false, err
    }
    return agent.Status == "active", nil
}

func (a *PortAdapter) GetAttestationsForMemory(ctx context.Context, memoryID uuid.UUID) ([]memory.AttestationRecord, error) {
    atts, err := a.attestation.GetForMemory(ctx, memoryID)
    if err != nil {
        return nil, err
    }

    result := make([]memory.AttestationRecord, len(atts))
    for i, att := range atts {
        result[i] = memory.AttestationRecord{
            ID:        att.ID,
            AgentDID:  att.AgentDID,
            MemoryID:  att.MemoryID,
            Claim:     att.Claim,
            Signature: att.Signature,
            Timestamp: att.Timestamp,
        }
    }

    return result, nil
}

func (a *PortAdapter) RecordEvent(ctx context.Context, agentDID string, eventType string, delta int, description string) error {
    return a.reputation.RecordEvent(ctx, agentDID, ReputationEvent{
        Type:        eventType,
        Delta:       delta,
        Description: description,
    })
}
```

### 6.4 Dependency Diagram

```
┌─────────────────────────────────────────┐
│         internal/memory                 │
│  • types.go                             │
│  • store.go (Store interface)           │
│  • router.go (Router)                   │
│  • postgres_store.go (PostgresStore)    │
│  • valkey_store.go (ValkeyStore)        │
│  • search.go (HybridSearcher)           │
│  • trust_port.go (TrustPort interface)  │  ← defines the contract
└─────────────────┬───────────────────────┘
                  │ implements
                  ▼
┌─────────────────────────────────────────┐
│         internal/trust                  │
│  • identity.go                          │
│  • registry.go                          │
│  • reputation.go                        │
│  • attestation.go                       │
│  • port_adapter.go (TrustPort)          │  ← implements the contract
└─────────────────────────────────────────┘
```

**Key principle:** `internal/memory` defines the interface (`TrustPort`). `internal/trust` provides the implementation (`PortAdapter`). This prevents circular imports and allows mocking the trust layer in memory tests.

---

## 7. Complete File Structure

```
internal/
├── memory/
│   ├── types.go              # Memory, Metadata, PIIField, AttestationRef, SearchQuery, etc.
│   ├── store.go              # Store interface
│   ├── router.go             # Router with dynamic backend selection
│   ├── postgres_store.go     # PostgreSQL implementation (CRUD + hybrid search)
│   ├── valkey_store.go       # Valkey implementation (Working Memory with TTL)
│   ├── search.go             # HybridSearcher with RRF
│   ├── trust_port.go         # TrustPort interface for Phase 2 integration
│   └── doc.go                # Package documentation
├── trust/
│   ├── identity.go           # (Phase 2 — exists)
│   ├── registry.go           # (Phase 2 — exists)
│   ├── reputation.go         # (Phase 2 — exists)
│   ├── attestation.go        # (Phase 2 — exists)
│   └── port_adapter.go       # NEW: implements memory.TrustPort
├── api/
│   ├── router.go             # HTTP router
│   ├── health.go             # Health endpoints
│   ├── memory_handlers.go    # NEW: REST handlers for /v1/memory/*
│   └── middleware.go         # Auth, rate limiting (Phase 4)
└── storage/
    ├── postgres.go           # PostgreSQL connection pool
    └── valkey.go             # Valkey client

migrations/
└── 000001_init.up.sql        # Full schema (update with this design)
```

---

## 8. AntiGravity Prompt

```
You are implementing Phase 1 (Memory Engine) of Acumius.

Prerequisites complete:
- Phase 0: Go scaffold, Docker Compose, CI, storage connections
- Phase 2: Trust Layer with Ed25519 DID, agent registry, reputation, attestation

Your task:

1. Create internal/memory/types.go with the 6 MemoryType constants, Memory struct,
   Metadata with AttestationRef, StoreRequest, SearchQuery, SearchResult.

2. Create internal/memory/store.go with the Store interface.

3. Create internal/memory/router.go that routes Working memory -> Valkey and
   all other types -> PostgreSQL.

4. Create internal/memory/postgres_store.go implementing Store for PostgreSQL
   with pgvector embeddings and full-text search.

5. Create internal/memory/valkey_store.go implementing Store for Valkey
   with TTL and sorted set indexing.

6. Create internal/memory/search.go with HybridSearcher using RRF (k=60)
   to merge semantic (pgvector) and keyword (tsvector) results.

7. Create internal/memory/trust_port.go with TrustPort interface.

8. Create internal/trust/port_adapter.go implementing TrustPort.

9. Update migrations/000001_init.up.sql with the complete schema including
   memories table with pgvector, full-text indexes, and all supporting tables.

10. Create internal/api/memory_handlers.go with REST endpoints:
    POST /v1/memory, GET /v1/memory/{id}, POST /v1/memory/search,
    GET /v1/memory/namespace/{ns}, DELETE /v1/memory/{id}

11. Wire everything into cmd/acumius/main.go.

12. Write tests with > 70% coverage on internal/memory/.

Constraints:
- Use placeholder embedding generation (random vectors) for v0.1
- Do NOT implement Policy Engine yet (Phase 3)
- Do NOT implement Governance UI yet (Phase 5)
- Handle all errors, no panics
- Follow existing code style
- Use github.com/go-chi/chi/v5 for HTTP routing
```

---

## Appendix: Quick Reference

### Memory Type Routing

| Type | Backend | TTL | Embedding | Full-Text |
|------|---------|-----|-----------|-----------|
| Working | Valkey | 24h | No | No |
| Episodic | PostgreSQL | Persistent | No | Yes |
| Semantic | PostgreSQL | Persistent | Yes (pgvector) | Yes |
| Procedural | PostgreSQL | Persistent | No | Yes |
| Declarative | PostgreSQL | Persistent | No | Yes |
| Feedback | PostgreSQL | Persistent | No | Yes |

### Key Design Decisions

| Decision | Why |
|----------|-----|
| Single `memories` table | Simpler, less migration overhead, easier hybrid search |
| `json.RawMessage` for content | Schema-less by design; agents store arbitrary JSON |
| `AttestationRef` in metadata | Avoids duplicating 200+ byte signatures in memory rows |
| `TrustPort` interface | Prevents circular imports, enables mocking |
| RRF (k=60) | Industry standard for sparse + dense retrieval; no training needed |
| IVFFlat (100 lists) | Good for < 1M vectors; switch to HNSW for larger datasets |
| Partial indexes (`deleted_at IS NULL`) | Excludes soft-deleted rows automatically |

---

*Document version: 1.0*  
*For the Acumius Core Team*  
*Phase 1 Implementation Guide*
