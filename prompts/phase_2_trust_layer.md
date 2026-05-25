# AntiGravity Prompt: Phase 2 — Trust Layer

## Context
You are implementing the Trust Layer for Acumius. Phase 1 is complete: the Memory Engine supports all 6 types with PostgreSQL + pgvector + Valkey, hybrid search, namespace ACL, and REST endpoints. Your task is to add agent identity, registration, reputation scoring, and memory attestation.

## Current State
- Memory Engine: full CRUD + search for all 6 types
- PostgreSQL store with pgvector
- Valkey store for Working Memory
- Storage router
- REST handlers for memory
- Basic namespace ACL
- Migration system running

## Your Task

### 1. Identity & Crypto
Create `internal/trust/identity.go`:

Implement Ed25519 keypair generation, DID format `did:acumius:<base58_pubkey>`, and parsing.

### 2. Agent Registry
Create `internal/trust/registry.go`:

Implement agent CRUD: Register, Get, Update, List using PostgreSQL.

### 3. Reputation Engine
Create `internal/trust/reputation.go`:

Implement reputation scoring:
- base_score: 500
- + completion_rate * 200
- + peer_verifications * 50
- + memory_attestations * 25
- - policy_violations * 100
- - disputes_lost * 150
- - days_inactive * 1
- Range: 0-1000

### 4. Attestation
Create `internal/trust/attestation.go`:

Implement memory attestation with Ed25519 signatures.

### 5. REST Handlers
Create `internal/api/trust_handlers.go`:

- POST /v1/agents/register
- GET /v1/agents/{did}
- PATCH /v1/agents/{did}
- POST /v1/agents/{did}/verify
- GET /v1/agents/{did}/reputation
- POST /v1/memory/{id}/attest
- GET /v1/memory/{id}/attestations

### 6. Peer Verification
Create `internal/trust/verification.go`:

Simple assignment: new agents get 3 random active agents to verify.

### 7. Tests
Target: > 75% coverage on internal/trust/

### Acceptance Criteria
- [ ] Agent can register with Ed25519 keypair
- [ ] DID format: did:acumius:<base58_pubkey>
- [ ] Reputation score updates after events
- [ ] Memory attestation is cryptographically verifiable
- [ ] make test passes with > 75% coverage
