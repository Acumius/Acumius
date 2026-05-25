# Acumius Policy Engine Specification

> **Version:** 1.0  
> **Last Updated:** 2026-05-24

---

## 1. Policy Language

Policies are declarative YAML documents that control what agents can do.

### 1.1 Structure

```yaml
policy_version: "1.0"        # Required. Policy format version
agent_id: "did:acumius:..."  # Optional. If omitted, applies as default

permissions:
  memory:
    working:
      read: ["self"]
      write: ["self"]
    semantic:
      read: ["self", "shared:project-alpha"]
      write: ["self"]
    episodic:
      read: ["self"]
      write: ["self"]
    procedural:
      read: ["self", "shared:team-devops"]
      write: ["self"]
    declarative:
      read: ["self", "shared:org-policies"]
      write: ["self"]  # Only admin agents can write org policies
    feedback:
      read: ["self"]
      write: ["self", "shared:project-alpha"]

  delegation:
    max_depth: 3
    allowed_to: ["reputation > 600"]
    max_cost_per_hour: 10.00

  pii:
    auto_redact: true
    retention_days: 30
    allowed_types: ["email", "phone"]  # Empty = all redacted

  audit:
    log_level: "all"  # all, write-only, none
    retention_days: 90
```

### 1.2 Permission Values

| Value | Meaning |
|-------|---------|
| `"self"` | Agent's own namespace |
| `"shared:ns-name"` | Specific shared namespace |
| `"*"` | All namespaces (admin) |
| `[]` | No access |

### 1.3 Delegation Rules

| Field | Type | Description |
|-------|------|-------------|
| `max_depth` | int | Max delegation chain length |
| `allowed_to` | []string | Who can be delegated to. Supports: `reputation > N`, `did:...`, `*` |
| `max_cost_per_hour` | float | Max USD cost per hour of delegation |

---

## 2. Evaluation Semantics

### 2.1 Request Context

```go
type EvaluationRequest struct {
    AgentDID    string
    Action      string       // "memory.read", "memory.write", "trust.verify", etc.
    Resource    Resource
    Metadata    map[string]interface{}
    Timestamp   time.Time
}

type Resource struct {
    Type      string  // "memory", "agent", "policy"
    Subtype   string  // "semantic", "episodic", etc.
    Namespace string
    ID        string
}
```

### 2.2 Evaluation Flow

1. Load agent's policy (or default if none)
2. If policy not found → DENY
3. Build evaluation context from request
4. Evaluate rules top-down
5. First matching rule wins
6. If no rule matches → default action (DENY)
7. If evaluator errors → DENY (fail-closed)

### 2.3 Decision Format

```json
{
  "allowed": true,
  "policy_id": "policy-001",
  "rule_matched": "memory.semantic.read.shared",
  "reason": "Agent is member of namespace 'project-alpha'",
  "evaluated_at": "2026-05-24T10:00:00Z",
  "latency_ms": 0.05
}
```

---

## 3. Rego Support (OPA)

For advanced users, policies can be written in Rego:

```rego
package acumius.policy

import future.keywords.if
import future.keywords.in

default allow := false

allow if {
    input.action == "memory.read"
    input.resource.type == "semantic"
    input.agent.reputation > 600
    input.resource.namespace == input.agent.namespace
}

allow if {
    input.action == "memory.write"
    input.resource.type == "working"
    input.agent.did == input.resource.agent_did
}
```

Rego policies are compiled using OPA's WASM SDK for sandboxed evaluation.

---

## 4. Performance

| Metric | Target |
|--------|--------|
| Evaluation p50 | < 0.1ms |
| Evaluation p99 | < 1ms |
| Cache hit rate | > 95% |
| Concurrent evals | 35K ops/sec |

---

## 5. Examples

### Minimal Policy
```yaml
policy_version: "1.0"
permissions:
  memory:
    working:
      read: ["self"]
      write: ["self"]
```

### Enterprise Policy
```yaml
policy_version: "1.0"
agent_id: "did:acumius:enterprise-agent-001"
permissions:
  memory:
    semantic:
      read: ["self", "shared:finance", "shared:hr"]
      write: ["self"]
    declarative:
      read: ["shared:org-policies"]
      write: []  # Read-only
  delegation:
    max_depth: 2
    allowed_to: ["reputation > 800"]
    max_cost_per_hour: 50.00
  pii:
    auto_redact: true
    retention_days: 7
  audit:
    log_level: "all"
    retention_days: 2555  # 7 years for SOX
```

---

*Policy changes require ADR and take effect immediately (cache invalidation).*
