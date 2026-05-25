# AntiGravity Prompt: Phase 4 — Protocol Layer

## Context
Expose Memory, Trust, and Policy via MCP, REST (OpenAPI), and AG-UI SSE.

## Your Task

### 1. MCP Server
Create `protocol/mcp/server.go` using `github.com/mark3labs/mcp-go`.

Expose 7 tools: memory_store, memory_retrieve, memory_search, memory_delete, agent_register, agent_reputation, policy_check.

### 2. OpenAPI 3.1 Spec
Create `protocol/rest/openapi.yaml` covering all endpoints.

### 3. AG-UI SSE
Create `protocol/agui/sse.go` streaming memory and audit events.

### 4. Authentication
Support API key and Ed25519 signature auth.

### 5. Rate Limiting
100 req/min per agent, using Valkey counters.

### Acceptance Criteria
- [ ] MCP passes inspector
- [ ] OpenAPI spec valid
- [ ] SSE streams events
- [ ] Both auth methods work
- [ ] Rate limiting enforced
