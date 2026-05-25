# AntiGravity Prompt: Phase 3 — Policy Engine

## Context
You are implementing the Policy Engine for Acumius. Phases 1-2 are complete. Your task is real-time policy enforcement.

## Your Task

### 1. Policy Schema
Create `internal/policy/types.go` with Policy struct supporting memory permissions, delegation, PII, and audit rules.

### 2. Parser
Create `internal/policy/parser.go` to parse YAML/JSON policies.

### 3. Evaluator
Create `internal/policy/evaluator.go` with compiled rule tree, < 0.1ms p50 target.

### 4. Cache
Create `internal/policy/cache.go` using Valkey, 5m TTL.

### 5. Middleware
Update `internal/api/middleware.go` to enforce policy on every request. Fail-closed.

### 6. Audit Logger
Create `internal/audit/logger.go` — fire-and-forget to PostgreSQL.

### 7. GDPR Tools
Create `internal/gdpr/` with redactor, exporter, forget, rectify.

### 8. REST Handlers
- /v1/policies/*
- /v1/audit/*
- /v1/audit/stream (SSE)
- /v1/gdpr/*

### Acceptance Criteria
- [ ] Policy evaluation < 0.1ms p50
- [ ] Fail-closed: errors = DENY
- [ ] Every API call audited
- [ ] GDPR right-to-forget works
- [ ] > 80% coverage
