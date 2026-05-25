# Acumius Implementation Phases

> **Version:** 1.0  
> **Last Updated:** 2026-05-24  
> **Status:** Phase 0 — Foundation (In Progress)

This document is the master record of all implementation phases for Acumius. Each phase has defined scope, deliverables, acceptance criteria, and handoff prompts for the development team.

---

## Phase 0: Foundation (Weeks 1-2)

### Goal
Establish the project scaffold, CI/CD, documentation baseline, and development environment so all subsequent phases build on a stable foundation.

### Deliverables

| # | Deliverable | Owner | Status |
|---|-------------|-------|--------|
| 0.1 | Go module initialization and project structure | Core | ✅ |
| 0.2 | Service entrypoint with HTTP server and graceful shutdown | Core | ✅ |
| 0.3 | Config loading via environment variables | Core | ✅ |
| 0.4 | `GET /health` endpoint with test coverage | Core | ✅ |
| 0.5 | CI workflow (fmt, lint, test) on PRs and `main` pushes | Core | ✅ |
| 0.6 | Makefile with quality targets (`fmt`, `lint`, `test`, `check`) | Core | ✅ |
| 0.7 | `.gitignore`, migration naming conventions, examples folder | Core | ✅ |
| 0.8 | Docker Compose baseline (PostgreSQL + pgvector + Valkey) | Core | 🔄 |
| 0.9 | README.md, CONTRIBUTING.md, CODE_OF_CONDUCT.md | Docs | ✅ |
| 0.10 | Architecture Decision Records (ADR-0001 through ADR-0006) | Arch | ✅ |

### Acceptance Criteria
- [ ] `make check` passes locally (fmt, lint, test)
- [ ] `docker-compose up` starts all services without errors
- [ ] `curl http://localhost:8080/health` returns `{"service":"acumius","status":"ok"}`
- [ ] CI passes on PR #16 and all subsequent PRs
- [ ] New contributor can clone, `make up`, and verify health in < 5 minutes

### Handoff Prompt for Next Phase
```
Phase 0 is complete. The Go service scaffold is stable with:
- cmd/acumius/main.go entrypoint
- internal/config, internal/api/router.go, internal/api/health.go
- Makefile with fmt, lint, test, check
- GitHub Actions CI workflow
- Docker Compose with PostgreSQL + pgvector + Valkey

Your task: Implement the Memory Engine (Phase 1). Build on the existing
internal/ package structure. Do not refactor the scaffold.
```

---

## Phase 1: Memory Engine (Weeks 3-6)

### Goal
Agents can store and retrieve all 6 memory types via REST API with namespace isolation, storage routing, and hybrid search.

### Scope
- Memory type definitions and JSON schema
- PostgreSQL schema with pgvector extension
- Valkey integration for Working Memory
- Storage router (routes memory type → correct backend)
- REST API endpoints for CRUD + search
- Namespace access control (basic ACL)
- Migration system (golang-migrate)

### Deliverables

| # | Deliverable | Owner | Weeks |
|---|-------------|-------|-------|
| 1.1 | Memory type definitions (`internal/memory/types.go`) | Core | 3 |
| 1.2 | PostgreSQL schema migration (000001_init.up.sql) | Core | 3 |
| 1.3 | PostgreSQL store implementation | Core | 3-4 |
| 1.4 | Valkey store implementation for Working Memory | Core | 3-4 |
| 1.5 | Storage router (`internal/memory/router.go`) | Core | 4 |
| 1.6 | REST handlers: POST /v1/memory, GET /v1/memory/{id} | API | 4-5 |
| 1.7 | REST handler: POST /v1/memory/search (hybrid) | API | 5 |
| 1.8 | Namespace ACL (basic read/write permissions) | Core | 5-6 |
| 1.9 | Integration tests for store + router + API | QA | 6 |
| 1.10 | Benchmark tests for store operations | Bench | 6 |

### API Endpoints (Phase 1)
```
POST   /v1/memory              → Store memory
GET    /v1/memory/{id}         → Retrieve by ID
POST   /v1/memory/search       → Hybrid semantic + keyword search
GET    /v1/memory/namespace/{ns} → List memories in namespace
DELETE /v1/memory/{id}         → Soft delete
```

### Acceptance Criteria
- [ ] Can store and retrieve all 6 memory types via REST
- [ ] Working Memory auto-expires after TTL (default 24h)
- [ ] Semantic search returns relevant results in < 50ms p99
- [ ] Namespace isolation prevents unauthorized cross-namespace reads
- [ ] `make test` passes with > 70% coverage on memory package
- [ ] `docker-compose up` includes migration auto-run on startup

### Handoff Prompt
```
Phase 1 is complete. The Memory Engine supports all 6 types with:
- PostgreSQL + pgvector for persistent memory
- Valkey for Working Memory with TTL
- Storage router that routes by memory type
- REST CRUD + hybrid search endpoints
- Basic namespace ACL

Your task: Implement the Trust Layer (Phase 2). Add agent identity,
registration, reputation scoring, and memory attestation. Build on the
existing memory types by adding Attestation struct to Metadata.
```

---

## Phase 2: Trust Layer (Weeks 7-10)

### Goal
Every agent has a cryptographically verifiable identity, a reputation score, and can attest memory writes.

### Scope
- Ed25519 keypair generation and DID format
- Agent registration and profile management
- Reputation scoring algorithm
- Memory attestation (cryptographic signing)
- Peer verification assignment and reporting
- REST API for trust operations

### Deliverables

| # | Deliverable | Owner | Weeks |
|---|-------------|-------|-------|
| 2.1 | Agent identity types and Ed25519 crypto (`internal/trust/identity.go`) | Core | 7 |
| 2.2 | Agent registration flow and database schema | Core | 7-8 |
| 2.3 | Reputation scoring engine (`internal/trust/reputation.go`) | Core | 8 |
| 2.4 | Memory attestation (sign MemoryID + Claim) | Core | 8-9 |
| 2.5 | Peer verification assignment and report submission | Core | 9 |
| 2.6 | REST handlers: /v1/agents/* | API | 9 |
| 2.7 | Integration tests for trust flow | QA | 10 |
| 2.8 | Update Memory Engine to include attestation in Metadata | Core | 9 |

### Reputation Formula (v1)
```
reputation = base_score (500)
  + (completion_rate * 200)
  + (peer_verifications * 50)
  + (memory_attestations * 25)
  - (policy_violations * 100)
  - (disputes_lost * 150)
  - (days_inactive * 1)  // decay

Range: 0-1000
```

### API Endpoints (Phase 2)
```
POST   /v1/agents/register     → Register new agent
GET    /v1/agents/{did}         → Get agent profile
PATCH  /v1/agents/{did}         → Update profile
POST   /v1/agents/{did}/verify  → Submit verification report
GET    /v1/agents/{did}/reputation → Get reputation breakdown
POST   /v1/memory/{id}/attest   → Attest a memory
GET    /v1/memory/{id}/attestations → List attestations
```

### Acceptance Criteria
- [ ] Agent can register with Ed25519 keypair and receive DID
- [ ] Reputation score updates correctly after events
- [ ] Memory attestation is verifiable (signature checks pass)
- [ ] Peer verification reports affect reputation
- [ ] `make test` passes with > 75% coverage on trust package

### Handoff Prompt
```
Phase 2 is complete. The Trust Layer provides:
- Ed25519-based agent identity with DID format
- Reputation scoring (0-1000) with decay
- Memory attestation with cryptographic signatures
- Peer verification system

Your task: Implement the Policy Engine (Phase 3). Build a real-time
policy evaluator that checks every memory access against YAML/Rego rules.
Integrate with the Trust Layer so policies can reference reputation scores.
```

---

## Phase 3: Policy Engine (Weeks 11-14)

### Goal
Real-time policy enforcement on every memory access and agent action, with tamper-evident audit logs and GDPR compliance tools.

### Scope
- Policy YAML schema and parser
- Policy evaluation engine (compiled rule tree, < 0.1ms target)
- Rego (Open Policy Agent) support
- Policy cache in Valkey
- Audit logging (append-only, partitioned)
- GDPR tools: right-to-forget, export, rectification, auto-expiry
- PII detection and redaction

### Deliverables

| # | Deliverable | Owner | Weeks |
|---|-------------|-------|-------|
| 3.1 | Policy YAML schema and parser (`internal/policy/parser.go`) | Core | 11 |
| 3.2 | Policy evaluation engine with compiled rule tree | Core | 11-12 |
| 3.3 | Rego (OPA) policy support | Core | 12 |
| 3.4 | Policy cache in Valkey | Core | 12 |
| 3.5 | Middleware: enforce policy on every API call | API | 12-13 |
| 3.6 | Audit logging system (`internal/audit/logger.go`) | Core | 13 |
| 3.7 | GDPR tools: redaction, export, forget, auto-expiry | Core | 13-14 |
| 3.8 | PII detection (basic regex + NER) | Core | 14 |
| 3.9 | REST handlers: /v1/policies/*, /v1/audit/*, /v1/gdpr/* | API | 13-14 |
| 3.10 | Integration tests for policy enforcement | QA | 14 |

### Policy Example
```yaml
policy_version: "1.0"
agent_id: "did:acumius:abc123"

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
      write: ["self"]
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

  audit:
    log_level: "all"
```

### API Endpoints (Phase 3)
```
POST   /v1/policies              → Create policy
GET    /v1/policies/{id}          → Get policy
PUT    /v1/policies/{id}          → Update policy
DELETE /v1/policies/{id}          → Delete policy
POST   /v1/policies/evaluate      → Evaluate a request against policy

GET    /v1/audit                  → Query audit log
GET    /v1/audit/stream           → SSE stream of audit events

POST   /v1/gdpr/right-to-forget   → Redact all data for agent/user
POST   /v1/gdpr/export            → Export all data
POST   /v1/gdpr/rectify           → Correct inaccurate data
POST   /v1/memory/redact          → Bulk redact PII
```

### Acceptance Criteria
- [ ] Policy evaluation completes in < 0.1ms p50 for single rule
- [ ] Policy errors result in DENY (fail-closed)
- [ ] Audit log is append-only and queryable by time range
- [ ] GDPR right-to-forget redacts all PII and soft-deletes memories
- [ ] Auto-expiry purges memories past `valid_until`
- [ ] `make test` passes with > 80% coverage on policy package

### Handoff Prompt
```
Phase 3 is complete. The Policy Engine provides:
- Real-time YAML/Rego policy evaluation (< 0.1ms)
- Fail-closed default behavior
- Tamper-evident audit logging
- GDPR tools (right-to-forget, export, rectification, auto-expiry)
- PII redaction

Your task: Implement the Protocol Layer (Phase 4). Build MCP server,
REST OpenAPI spec, and AG-UI SSE endpoints that expose all three pillars
(Memory, Trust, Policy) to external agents.
```

---

## Phase 4: Protocol Layer (Weeks 15-16)

### Goal
Agents connect to Acumius via MCP (primary), REST, or AG-UI (Server-Sent Events).

### Scope
- MCP server implementation (tools, resources, prompts)
- REST API OpenAPI 3.1 specification
- AG-UI SSE server for real-time events
- Authentication middleware (API key + Ed25519 signature)
- Rate limiting

### Deliverables

| # | Deliverable | Owner | Weeks |
|---|-------------|-------|-------|
| 4.1 | MCP server scaffold with health check | Protocol | 15 |
| 4.2 | MCP tools: acumius_memory_store, acumius_memory_retrieve, etc. | Protocol | 15 |
| 4.3 | REST OpenAPI 3.1 spec (`protocol/rest/openapi.yaml`) | API | 15 |
| 4.4 | REST handler validation against OpenAPI spec | API | 15-16 |
| 4.5 | AG-UI SSE server for memory updates and audit events | Protocol | 16 |
| 4.6 | Authentication middleware (API key + Ed25519) | Core | 15 |
| 4.7 | Rate limiting middleware | Core | 16 |
| 4.8 | End-to-end tests: LangGraph → MCP → Acumius → Memory | QA | 16 |

### MCP Tools
```
acumius_memory_store       → Store a memory
acumius_memory_retrieve    → Retrieve memory by ID
acumius_memory_search      → Hybrid search
acumius_memory_delete      → Soft delete memory
acumius_agent_register     → Register agent identity
acumius_agent_reputation   → Get agent reputation
acumius_policy_check       → Check if action is allowed
```

### Acceptance Criteria
- [ ] MCP server passes MCP inspector validation
- [ ] REST API conforms to OpenAPI 3.1 spec
- [ ] AG-UI SSE streams memory updates in real time
- [ ] Authentication rejects invalid API keys and signatures
- [ ] Rate limiting prevents abuse
- [ ] LangGraph agent can store/retrieve memory via MCP

### Handoff Prompt
```
Phase 4 is complete. The Protocol Layer provides:
- MCP server with 7 tools exposing Memory, Trust, and Policy
- REST API with OpenAPI 3.1 spec
- AG-UI SSE for real-time events
- API key + Ed25519 authentication
- Rate limiting

Your task: Implement the Governance UI (Phase 5). Build a Next.js
dashboard with Memory Explorer, Agent Directory, Policy Editor, Audit
Log, and GDPR Tools screens.
```

---

## Phase 5: Governance UI (Weeks 17-18)

### Goal
Web dashboard for humans to inspect memories, manage agents, edit policies, view audit logs, and execute GDPR workflows.

### Scope
- Next.js 15 + shadcn/ui scaffold
- Memory Explorer screen
- Agent Directory screen
- Policy Editor screen
- Audit Log screen
- GDPR Tools screen
- Dashboard metrics screen
- Authentication integration

### Deliverables

| # | Deliverable | Owner | Weeks |
|---|-------------|-------|-------|
| 5.1 | Next.js 15 scaffold with shadcn/ui | UI | 17 |
| 5.2 | API client layer (TanStack Query) | UI | 17 |
| 5.3 | Memory Explorer: browse, search, filter, edit | UI | 17-18 |
| 5.4 | Agent Directory: list, search, view reputation | UI | 17-18 |
| 5.5 | Policy Editor: visual YAML editor with validation | UI | 18 |
| 5.6 | Audit Log: real-time stream, filter, export | UI | 18 |
| 5.7 | GDPR Tools: right-to-forget wizard, export, rectify | UI | 18 |
| 5.8 | Dashboard: metrics, charts, agent health | UI | 18 |
| 5.9 | Auth integration (login with API key) | UI | 18 |
| 5.10 | E2E tests with Playwright | QA | 18 |

### Acceptance Criteria
- [ ] User can browse all memories by type, namespace, agent
- [ ] User can search memories with hybrid search
- [ ] User can view agent profiles and reputation history
- [ ] User can create and edit policies with live validation
- [ ] Audit log streams in real time
- [ ] GDPR right-to-forget workflow completes in < 3 clicks
- [ ] Dashboard shows key metrics (memories, agents, violations)

### Handoff Prompt
```
Phase 5 is complete. The Governance UI provides:
- Next.js 15 dashboard with 6 screens
- Real-time audit log streaming
- Visual policy editor
- GDPR workflow tools
- Metrics dashboard

Your task: Implement SDKs and Adapters (Phase 6). Build Python and
TypeScript SDKs, plus LangGraph, CrewAI, AutoGen, and OpenAI adapters.
```

---

## Phase 6: SDKs & Adapters (Weeks 19-20)

### Goal
Developers can integrate Acumius into their agents with minimal code.

### Scope
- Python SDK (PyPI package)
- TypeScript SDK (npm package)
- LangGraph drop-in memory adapter
- CrewAI adapter
- AutoGen adapter
- OpenAI Agents SDK adapter
- Documentation and examples

### Deliverables

| # | Deliverable | Owner | Weeks |
|---|-------------|-------|-------|
| 6.1 | Python SDK scaffold and core client | SDK | 19 |
| 6.2 | Python SDK: memory, trust, policy modules | SDK | 19 |
| 6.3 | TypeScript SDK scaffold and core client | SDK | 19 |
| 6.4 | TypeScript SDK: memory, trust, policy modules | SDK | 19-20 |
| 6.5 | LangGraph AcumiusMemory adapter | Adapter | 20 |
| 6.6 | CrewAI adapter | Adapter | 20 |
| 6.7 | AutoGen adapter | Adapter | 20 |
| 6.8 | OpenAI Agents SDK adapter | Adapter | 20 |
| 6.9 | SDK documentation and quickstart guides | Docs | 20 |
| 6.10 | Integration tests for each adapter | QA | 20 |

### Python SDK Example
```python
from acumius import AcumiusClient

client = AcumiusClient(base_url="http://localhost:8080", api_key="...")

# Store memory
client.memory.store(type="semantic", namespace="my-project", content={...})

# Search
results = client.memory.search(query="revenue", types=["semantic", "episodic"])

# Check reputation
rep = client.trust.get_reputation("did:acumius:analyst-001")
```

### Acceptance Criteria
- [ ] Python SDK published to PyPI with `pip install acumius`
- [ ] TypeScript SDK published to npm with `npm install acumius`
- [ ] LangGraph adapter passes LangGraph memory interface tests
- [ ] Each adapter has a working example in `examples/`
- [ ] SDK documentation covers all three pillars

### Handoff Prompt
```
Phase 6 is complete. SDKs and Adapters provide:
- Python SDK on PyPI
- TypeScript SDK on npm
- LangGraph, CrewAI, AutoGen, OpenAI adapters
- Working examples

Your task: Build the Killer Demo and Launch (Phase 7). Create the
cross-framework research team demo, write launch content, and prepare
for community release.
```

---

## Phase 7: Demo & Launch (Weeks 21-24)

### Goal
Ship v1.0 with a compelling demo, comprehensive documentation, and community launch.

### Scope
- Cross-framework research team demo
- Benchmark suite vs competitors
- Documentation site (MkDocs)
- Security audit and hardening
- Docker image publishing (GHCR)
- Package publishing (PyPI, npm)
- Launch content (blog post, HN, Twitter)
- Conference talk preparation

### Deliverables

| # | Deliverable | Owner | Weeks |
|---|-------------|-------|-------|
| 7.1 | Cross-framework demo: LangGraph + CrewAI + Custom | Demo | 21 |
| 7.2 | Benchmark suite: vs Mem0, Letta, raw PostgreSQL | Bench | 21-22 |
| 7.3 | Documentation site with MkDocs | Docs | 22 |
| 7.4 | Security audit (dependency scan, fuzzing) | Security | 22 |
| 7.5 | Docker image build and GHCR publishing | DevOps | 22 |
| 7.6 | PyPI and npm package publishing | DevOps | 22 |
| 7.7 | Launch blog post and social content | Marketing | 23 |
| 7.8 | Hacker News launch preparation | Marketing | 23 |
| 7.9 | Conference talk abstract and slides | Community | 23-24 |
| 7.10 | v1.0 release tag and changelog | Release | 24 |

### The Killer Demo

**Scenario: Cross-Framework Market Analysis Team**

1. **Agent A** (LangGraph) creates shared namespace `market-analysis-q3`
2. **Agent A** delegates web scraping to **Agent B** (CrewAI)
3. **Agent B** writes raw data to Episodic Memory in shared namespace
4. **Agent C** (custom Python) verifies facts and attests Semantic Memory
5. **Agent A** reads attested facts and writes final report
6. Audit log shows: who did what, when, with what permissions
7. Governance UI displays the full collaboration timeline

### Acceptance Criteria
- [ ] Demo runs end-to-end without manual intervention
- [ ] Benchmarks show competitive or superior performance
- [ ] Documentation site covers all features
- [ ] Security audit passes with no critical vulnerabilities
- [ ] Docker image `ghcr.io/acumius/acumius:latest` is pullable
- [ ] PyPI and npm packages installable
- [ ] Launch content is ready for HN, Twitter, Discord

---

## Post-v1.0 Roadmap

| Version | Focus | Key Features |
|---------|-------|-------------|
| **v1.1** | Scale | Connection pooling, read replicas, horizontal scaling |
| **v1.2** | Advanced Trust | Delegation chains, trust ceilings, dispute arbitration |
| **v1.3** | Marketplace | Agent discovery, capability advertising, task posting |
| **v1.4** | Economics | Payment rails, escrow, micro-transactions between agents |
| **v1.5** | Enterprise | SSO/SAML, RBAC, advanced compliance, SIEM integrations |
| **v2.0** | Agent Kernel | Full OS abstraction — process scheduling, I/O management, network stack |

---

## Phase Tracking

Track live progress via:
- [GitHub Milestones](https://github.com/Acumius/Acumius/milestones)
- [Project Board](https://github.com/Acumius/Acumius/projects)
- [Issues by label](https://github.com/Acumius/Acumius/labels)

---

*Document maintained by the Core Team. Updates require PR approval.*
