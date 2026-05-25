# Acumius REST API Specification

> **Version:** 1.0  
> **OpenAPI:** 3.1.0  
> **Base URL:** `http://localhost:8080`  
> **Last Updated:** 2026-05-24

---

## Authentication

All endpoints require authentication via one of:

1. **API Key:** `Authorization: Bearer <api_key>`
2. **Ed25519 Signature:** `Authorization: AgentSig <base58_pubkey>:<base64_signature>` + `X-Timestamp` header

---

## Memory Endpoints

### POST /v1/memory
Store a new memory.

**Request:**
```json
{
  "type": "semantic",
  "namespace": "project-alpha",
  "content": {
    "fact": "Q3 revenue grew 23% YoY",
    "source": "earnings_call",
    "confidence": 0.95
  },
  "metadata": {
    "tags": ["finance", "q3"],
    "source": "earnings_call_2026",
    "confidence": 0.95
  },
  "valid_from": "2026-05-01T00:00:00Z",
  "valid_until": "2027-05-01T00:00:00Z"
}
```

**Response (201):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "semantic",
  "namespace": "project-alpha",
  "agent_did": "did:acumius:abc123",
  "content": {...},
  "metadata": {...},
  "created_at": "2026-05-24T10:00:00Z",
  "status": "stored"
}
```

### GET /v1/memory/{id}
Retrieve memory by ID.

**Response (200):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "semantic",
  "namespace": "project-alpha",
  "agent_did": "did:acumius:abc123",
  "content": {...},
  "metadata": {...},
  "attestations": [
    {
      "agent_did": "did:acumius:verifier-001",
      "claim": "Verified against SEC filing",
      "timestamp": "2026-05-24T11:00:00Z"
    }
  ],
  "created_at": "2026-05-24T10:00:00Z",
  "updated_at": "2026-05-24T10:00:00Z"
}
```

### POST /v1/memory/search
Hybrid semantic + keyword search.

**Request:**
```json
{
  "query": "revenue growth Q3",
  "types": ["semantic", "episodic"],
  "namespaces": ["project-alpha", "shared:finance"],
  "limit": 10,
  "offset": 0,
  "filters": {
    "confidence_min": 0.8,
    "created_after": "2026-01-01T00:00:00Z"
  }
}
```

**Response (200):**
```json
{
  "results": [
    {
      "id": "...",
      "type": "semantic",
      "score": 0.92,
      "content": {...},
      "metadata": {...}
    }
  ],
  "total": 42,
  "limit": 10,
  "offset": 0
}
```

### DELETE /v1/memory/{id}
Soft delete memory (GDPR compliant).

**Response (204):** No content

### POST /v1/memory/redact
Bulk redact PII from memories.

**Request:**
```json
{
  "namespace": "project-alpha",
  "pii_types": ["email", "phone", "ssn"],
  "dry_run": false
}
```

---

## Trust Endpoints

### POST /v1/agents/register
Register a new agent.

**Request:**
```json
{
  "public_key": "base64-ed25519-public-key",
  "name": "Market Analyst",
  "capabilities": ["data-analysis", "reporting"],
  "protocols": ["mcp", "rest"]
}
```

**Response (201):**
```json
{
  "did": "did:acumius:z6MkhaXg...",
  "api_key": "acu_live_xxxxxxxx",
  "created_at": "2026-05-24T10:00:00Z"
}
```

### GET /v1/agents/{did}
Get agent profile.

**Response (200):**
```json
{
  "did": "did:acumius:abc123",
  "name": "Market Analyst",
  "capabilities": ["data-analysis", "reporting"],
  "reputation_score": 847,
  "status": "active",
  "created_at": "2026-05-24T10:00:00Z"
}
```

### GET /v1/agents/{did}/reputation
Get detailed reputation breakdown.

**Response (200):**
```json
{
  "score": 847,
  "breakdown": {
    "base": 500,
    "completion_rate": 150,
    "peer_verifications": 100,
    "attestations": 50,
    "violations": -50,
    "disputes": 0,
    "decay": -3
  },
  "history": [
    {
      "event_type": "task_complete",
      "delta": 25,
      "timestamp": "2026-05-23T15:00:00Z"
    }
  ]
}
```

### POST /v1/agents/{did}/verify
Submit peer verification report.

**Request:**
```json
{
  "target_did": "did:acumius:target-001",
  "result": "pass",
  "capabilities_tested": ["data-analysis"],
  "notes": "Accurate and fast"
}
```

### POST /v1/memory/{id}/attest
Attest a memory.

**Request:**
```json
{
  "claim": "Verified against primary source",
  "signature": "base64-ed25519-signature"
}
```

---

## Policy Endpoints

### POST /v1/policies
Create a new policy.

**Request:**
```json
{
  "agent_did": "did:acumius:abc123",
  "content": {
    "policy_version": "1.0",
    "permissions": {
      "memory": {
        "semantic": {
          "read": ["self", "shared:project-alpha"],
          "write": ["self"]
        }
      }
    }
  },
  "version": "1.0"
}
```

### POST /v1/policies/evaluate
Evaluate a hypothetical request.

**Request:**
```json
{
  "agent_did": "did:acumius:abc123",
  "action": "memory.read",
  "resource": {
    "type": "semantic",
    "namespace": "project-alpha"
  }
}
```

**Response (200):**
```json
{
  "allowed": true,
  "policy_id": "policy-001",
  "reason": "Agent is namespace member",
  "evaluated_at": "2026-05-24T10:00:00Z"
}
```

---

## Audit Endpoints

### GET /v1/audit
Query audit log.

**Query Parameters:**
- `agent_did` — filter by agent
- `action` — filter by action type
- `allowed` — filter by result
- `from` — start timestamp
- `to` — end timestamp
- `limit` — max results (default 100)
- `offset` — pagination offset

**Response (200):**
```json
{
  "events": [...],
  "total": 1000,
  "limit": 100,
  "offset": 0
}
```

### GET /v1/audit/stream
SSE stream of real-time audit events.

---

## GDPR Endpoints

### POST /v1/gdpr/right-to-forget
Redact all data for an agent.

**Request:**
```json
{
  "agent_did": "did:acumius:abc123",
  "confirm": true
}
```

### POST /v1/gdpr/export
Export all data for an agent.

**Response:** JSON file download

### POST /v1/gdpr/rectify
Correct inaccurate memory.

**Request:**
```json
{
  "memory_id": "...",
  "correction": "new corrected content"
}
```

---

## Health Endpoints

### GET /health
Liveness probe.

**Response (200):**
```json
{
  "service": "acumius",
  "status": "ok",
  "version": "0.1.0"
}
```

### GET /ready
Readiness probe.

**Response (200):**
```json
{
  "service": "acumius",
  "status": "ready",
  "dependencies": {
    "postgresql": "connected",
    "valkey": "connected"
  }
}
```

### GET /metrics
Prometheus metrics.

---

*Full OpenAPI 3.1 spec available at `/openapi.json` when service is running.*
