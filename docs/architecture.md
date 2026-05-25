# Acumius Architecture Specification

> **Version:** 1.0  
> **Status:** Draft (Pending Phase 1 Implementation)  
> **Last Updated:** 2026-05-24

---

## 1. Overview

Acumius is a local-first, single-binary Go service that provides three integrated subsystems:

1. **Memory Engine** — Structured multi-type memory with namespace isolation
2. **Trust Layer** — Cryptographic identity, reputation, and attestation
3. **Policy Engine** — Real-time rule enforcement and governance

Agents connect via MCP (primary), REST, or AG-UI (Server-Sent Events). The service routes each request through authentication → policy evaluation → subsystem execution → audit logging.

---

## 2. System Context

```
┌─────────────────────────────────────────────────────────────┐
│                    AGENT ECOSYSTEM                           │
│     LangGraph · CrewAI · AutoGen · OpenAI · Custom           │
│                                                              │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐         │
│  │ Agent A │  │ Agent B │  │ Agent C │  │ Agent D │         │
│  │(LangG)  │  │(CrewAI) │  │(AutoGen)│  │(Custom) │         │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘         │
└───────┼────────────┼────────────┼────────────┼──────────────┘
        │            │            │            │
        └────────────┴────────────┴────────────┘
                         │
              ┌──────────▼──────────┐
              │   MCP / REST / AG-UI │  ← Protocol Layer
              └──────────┬──────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    ACUMIUS CORE                              │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   MEMORY    │  │   TRUST     │  │      POLICY         │  │
│  │   ENGINE    │  │   LAYER     │  │      ENGINE         │  │
│  │             │  │             │  │                     │  │
│  │ • Store     │  │ • Register  │  │ • Parse             │  │
│  │ • Retrieve  │  │ • Verify    │  │ • Evaluate          │  │
│  │ • Search    │  │ • Reputation│  │ • Enforce           │  │
│  │ • Distill   │  │ • Attest    │  │ • Cache             │  │
│  │ • Route     │  │ • Delegate  │  │ • Audit             │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              STORAGE ROUTING LAYER                       ││
│  │  Working → Valkey  │  Semantic → PostgreSQL + pgvector  ││
│  │  Episodic → PostgreSQL  │  Procedural → PostgreSQL      ││
│  │  Declarative → PostgreSQL  │  Feedback → PostgreSQL     ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              AUDIT & GDPR LAYER                        ││
│  │  • Append-only audit log  │  • PII registry            ││
│  │  • GDPR export/forget    │  • Auto-expiry worker      ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                         │
              ┌──────────▼──────────┐
│              GOVERNANCE UI (Next.js)                         │
│     Memory Explorer · Agent Directory · Policy Editor       │
│     Audit Log · GDPR Tools · Dashboard                      │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Component Details

### 3.1 Memory Engine

**Responsibility:** Store, retrieve, search, and distill memories across 6 types.

**Key Components:**

| Component | File | Responsibility |
|-----------|------|---------------|
| Types | `internal/memory/types.go` | Memory struct, Metadata, PIIField, Attestation |
| Store Interface | `internal/memory/store.go` | Abstract store interface |
| Postgres Store | `internal/memory/postgres_store.go` | Persistent memory CRUD |
| Valkey Store | `internal/memory/valkey_store.go` | Working memory with TTL |
| Router | `internal/memory/router.go` | Route memory type → backend |
| Search | `internal/memory/search.go` | Hybrid semantic + keyword |
| Distiller | `internal/memory/distiller.go` | Background episodic → semantic |

**Data Flow — Store:**
```
Agent → REST/MCP → Auth → Policy Check → Memory Router
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    │                         │                         │
                    ▼                         ▼                         ▼
              ┌─────────┐              ┌──────────┐              ┌─────────┐
              │ Valkey  │              │ Postgres │              │ pgvector│
              │(Working)│              │(Episodic│              │(Semantic│
              │         │              │ Procedural│             │         │
              └─────────┘              │ Declarative│            └─────────┘
                                       │ Feedback) │
                                       └──────────┘
```

**Data Flow — Retrieve:**
```
Agent → Search Query → Hybrid Search
                              │
                    ┌─────────┴─────────┐
                    │                   │
                    ▼                   ▼
              ┌──────────┐      ┌──────────┐
              │ Full-Text│      │  Vector  │
              │  (tsvector)│     │ (pgvector)│
              └────┬─────┘      └────┬─────┘
                   │                   │
                   └─────────┬─────────┘
                             ▼
                    ┌──────────────┐
                    │ Result Merge │
                    │ (RRF scoring) │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │ Policy Filter │
                    │ (agent can only│
                    │  see allowed) │
                    └──────┬───────┘
                           ▼
                         Agent
```

### 3.2 Trust Layer

**Responsibility:** Agent identity, registration, reputation, attestation, and delegation.

**Key Components:**

| Component | File | Responsibility |
|-----------|------|---------------|
| Identity | `internal/trust/identity.go` | DID, Ed25519 keypair |
| Registry | `internal/trust/registry.go` | Agent CRUD |
| Reputation | `internal/trust/reputation.go` | Score calculation, decay |
| Attestation | `internal/trust/attestation.go` | Sign and verify memory |
| Delegation | `internal/trust/delegation.go` | Chain validation |

**DID Format:**
```
did:acumius:<base58-encoded-ed25519-public-key>

Example: did:acumius:z6MkhaXg... (44 chars)
```

**Registration Flow:**
```
Agent generates Ed25519 keypair
        │
        ▼
POST /v1/agents/register
  { public_key: "...", name: "...", capabilities: [...] }
        │
        ▼
Acumius validates key format
        │
        ▼
Acumius creates DID, stores profile
        │
        ▼
Returns: { did: "did:acumius:...", api_key: "..." }
        │
        ▼
Agent uses DID + API key (or DID + signature) for all requests
```

### 3.3 Policy Engine

**Responsibility:** Parse, evaluate, and enforce policies on every request.

**Key Components:**

| Component | File | Responsibility |
|-----------|------|---------------|
| Parser | `internal/policy/parser.go` | YAML/JSON → internal struct |
| Evaluator | `internal/policy/evaluator.go` | Rule evaluation logic |
| Cache | `internal/policy/cache.go` | Compiled policy cache in Valkey |
| Middleware | `internal/api/middleware.go` | Enforce on every request |

**Evaluation Flow:**
```
Request arrives
      │
      ▼
Extract agent DID from auth
      │
      ▼
Load policy for agent (cache first)
      │
      ▼
Build evaluation context:
  { agent: {did, reputation, ...},
    action: "memory.read",
    resource: {type: "semantic", namespace: "..."},
    metadata: {...} }
      │
      ▼
Evaluate rules top-down
      │
      ▼
Result: ALLOW or DENY
      │
      ▼
Write audit event
      │
      ▼
Return result (or 403 if DENY)
```

**Fail-Closed Principle:**
- If policy engine errors → DENY
- If no policy found for agent → use default policy (DENY all)
- If rule syntax invalid → DENY

---

## 4. Data Flow — Complete Request Lifecycle

```
1. REQUEST
   Agent sends: POST /v1/memory
   Headers: Authorization: Bearer <api_key>
             X-Agent-DID: did:acumius:abc123
   Body: { type: "semantic", namespace: "project-alpha", content: {...} }

2. AUTHENTICATION
   Middleware validates API key or Ed25519 signature
   Extracts agent DID

3. POLICY EVALUATION
   Policy Engine loads agent's policy
   Evaluates: "Can did:acumius:abc123 write semantic memory to project-alpha?"
   Result: ALLOW (agent is member of namespace)

4. SUBSYSTEM EXECUTION
   Memory Engine receives request
   Router determines: semantic → PostgreSQL + pgvector
   Generates embedding (async if configured)
   Stores in PostgreSQL
   Indexes in pgvector

5. AUDIT LOGGING
   Audit Logger writes:
   { id: uuid, timestamp: now, agent_did: "did:acumius:abc123",
     action: "memory.write", resource: "<memory_id>",
     allowed: true, policy_id: "policy-001", reason: "namespace member" }

6. RESPONSE
   Returns: { id: "<memory_id>", status: "stored", namespace: "project-alpha" }
```

---

## 5. Storage Architecture

### 5.1 PostgreSQL

**Role:** Persistent storage for all memory types except Working.

**Tables:**
- `memories` — unified memory table (see schema.md)
- `agents` — identity and reputation
- `namespace_acl` — access control
- `attestations` — memory attestations
- `audit_log` — append-only audit (partitioned by month)
- `reputation_events` — reputation change history
- `policies` — policy documents
- `pii_registry` — PII tracking for GDPR

**Indexes:**
- GIN on `memories.content` (full-text)
- IVFFlat on `memories.embedding` (vector search)
- B-tree on `memories.namespace`, `memories.agent_did`, `memories.type`
- B-tree on `audit_log.timestamp` (partition pruning)

**Partitioning:**
- `audit_log` partitioned by `RANGE(timestamp)` monthly
- `memories` partitioned by `LIST(type)` (optional for v1.1)

### 5.2 Valkey

**Role:** High-speed Working Memory with TTL.

**Key Format:**
```
acumius:working:{namespace}:{agent_did}:{memory_id}
acumius:working:{namespace}:{agent_did}:latest  → sorted set
acumius:policy:{agent_did}  → cached compiled policy
acumius:session:{session_id}  → session metadata
```

**TTL Strategy:**
- Working Memory: 24h default, configurable per-namespace
- Policy cache: 5m TTL, invalidated on policy update
- Session data: session lifetime

---

## 6. Concurrency Model

**Go Routine Model:**
- One goroutine per HTTP request
- Background workers (distiller, auto-expiry) run as separate goroutines
- PostgreSQL connection pool: 25 connections default
- Valkey connection pool: 50 connections default

**Synchronization:**
- Memory writes: synchronous (agent waits for confirmation)
- Embedding generation: asynchronous (background goroutine, agent gets ID immediately)
- Distillation: background cron (every 15 minutes)
- Auto-expiry: background cron (every hour)
- Audit logging: asynchronous fire-and-forget (durable queue in PostgreSQL)

---

## 7. Error Handling

**Error Categories:**

| Code | HTTP | Description | Action |
|------|------|-------------|--------|
| `UNAUTHORIZED` | 401 | Invalid API key or signature | Return 401, log attempt |
| `FORBIDDEN` | 403 | Policy denied request | Return 403, write audit DENY |
| `NOT_FOUND` | 404 | Memory or agent not found | Return 404 |
| `VALIDATION_ERROR` | 400 | Invalid request body | Return 400 with details |
| `CONFLICT` | 409 | Duplicate registration | Return 409 |
| `INTERNAL_ERROR` | 500 | Unexpected server error | Return 500, log error, alert |
| `RATE_LIMITED` | 429 | Too many requests | Return 429, Retry-After header |

**Error Response Format:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Agent does not have write permission for namespace 'project-alpha'",
    "details": {
      "agent_did": "did:acumius:abc123",
      "namespace": "project-alpha",
      "required_permission": "write",
      "policy_id": "policy-001"
    }
  }
}
```

---

## 8. Security Model

### 8.1 Threat Model

| Threat | Mitigation |
|--------|-----------|
| Agent impersonation | Ed25519 signatures or API keys |
| Memory tampering | Attestation + audit log |
| Privilege escalation | Policy fail-closed, trust ceilings |
| PII leakage | Auto-redaction + GDPR tools |
| Replay attacks | Timestamp + nonce in signatures |
| DoS | Rate limiting + connection pooling |
| Supply chain | SBOM, signed releases, Dependabot |

### 8.2 Security Boundaries

```
┌─────────────────────────────────────────┐
│  UNTRUSTED: Agent code (any framework)  │
├─────────────────────────────────────────┤
│  TRUSTED: Acumius binary (Go)           │
│  • Policy engine                        │
│  • Authentication                       │
│  • Audit logging                        │
├─────────────────────────────────────────┤
│  TRUSTED: PostgreSQL + Valkey           │
│  (run in same VPC, TLS required)       │
└─────────────────────────────────────────┘
```

---

## 9. Deployment Architecture

### 9.1 Local Development

```bash
docker-compose up
# Starts: Acumius, PostgreSQL, Valkey, Governance UI
```

### 9.2 Single Server Production

```
┌─────────────────────────────┐
│  Acumius Binary (systemd)   │
│  Port 8080                  │
├─────────────────────────────┤
│  PostgreSQL 16 + pgvector   │
│  Port 5432 (localhost only) │
├─────────────────────────────┤
│  Valkey                     │
│  Port 6379 (localhost only) │
└─────────────────────────────┘
```

### 9.3 High Availability (v1.1+)

```
┌─────────────┐     ┌─────────────┐
│  Acumius    │◄───►│  Acumius    │
│  Instance 1 │     │  Instance 2 │
└──────┬──────┘     └──────┬──────┘
       │                   │
       └─────────┬─────────┘
                 ▼
        ┌───────────────┐
        │  PostgreSQL   │
        │  Primary-Rep  │
        └───────────────┘
                 │
        ┌───────────────┐
        │    Valkey     │
        │   Sentinel    │
        └───────────────┘
```

---

## 10. Monitoring & Observability

**Metrics (Prometheus):**
- `acumius_requests_total` — counter by endpoint, status
- `acumius_request_duration_seconds` — histogram
- `acumius_policy_evaluations_total` — counter by result
- `acumius_memory_operations_total` — counter by type, operation
- `acumius_active_agents` — gauge
- `acumius_reputation_distribution` — histogram

**Health Checks:**
- `GET /health` — liveness (always 200 if process running)
- `GET /ready` — readiness (200 only if PostgreSQL + Valkey connected)
- `GET /metrics` — Prometheus metrics

**Logging:**
- Structured JSON logs
- Levels: DEBUG, INFO, WARN, ERROR
- Sensitive fields redacted automatically

---

## 11. Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Policy evaluation p50 | < 0.1ms | Benchmark test |
| Policy evaluation p99 | < 1ms | Benchmark test |
| Memory store p99 | < 5ms | Integration test |
| Memory retrieve p99 | < 2ms | Integration test |
| Semantic search p99 | < 50ms | Benchmark test |
| Concurrent agents | 10,000+ | Load test |
| Cold start | < 2s | Binary startup |
| Memory usage | < 256MB idle | Runtime metric |

---

*Document version controlled. Changes require ADR and PR approval.*
