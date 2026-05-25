# Acumius

<p align="center">
  <img src="https://raw.githubusercontent.com/Acumius/Acumius/main/assets/logo.svg" alt="Acumius" width="120">
</p>

<p align="center">
  <b>The Agent Collaboration Fabric</b><br>
  Persistent structured memory · Verifiable identity · Real-time governance
</p>

<p align="center">
  <a href="https://github.com/Acumius/Acumius/releases"><img src="https://img.shields.io/github/v/release/Acumius/Acumius?style=flat-square" alt="Release"></a>
  <a href="https://github.com/Acumius/Acumius/actions"><img src="https://img.shields.io/github/actions/workflow/status/Acumius/Acumius/go-quality.yml?style=flat-square" alt="CI"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square" alt="License"></a>
  <a href="https://pkg.go.dev/github.com/Acumius/Acumius"><img src="https://img.shields.io/badge/go.dev-reference-00ADD8?style=flat-square&logo=go" alt="Go Reference"></a>
  <a href="https://discord.gg/acumius"><img src="https://img.shields.io/discord/1234567890?style=flat-square&logo=discord&color=5865F2" alt="Discord"></a>
</p>

---

## What is Acumius?

Acumius is an **open-source, local-first infrastructure layer** that enables AI agents from any framework to collaborate with persistent memory, verifiable identity, and enforced governance.

Think of it as the **trust and memory bus** for the agent internet. LangChain agents, CrewAI agents, AutoGen agents, and custom Python scripts can all connect to Acumius, share structured memory across namespaces, verify each other's identity and reputation, and operate under real-time policy enforcement — without trusting each other blindly.

### The Problem

Today's agent ecosystem is fragmented and siloed:

- **Memory is ephemeral** — agents start from zero every session
- **Frameworks are walled gardens** — LangGraph memory only works in LangGraph
- **Trust is nonexistent** — agents can't verify who they're talking to
- **Governance is an afterthought** — no audit trail, no policy enforcement, no GDPR compliance
- **Collaboration is impossible** — cross-framework agents can't share memory or delegate tasks securely

### The Solution

Acunius provides three integrated pillars:

| Pillar | What It Does | Why It Matters |
|--------|-------------|----------------|
| **Memory Engine** | 6-type structured memory (Working, Episodic, Semantic, Procedural, Declarative, Feedback) with namespace isolation, temporal validity, and semantic search | Agents remember, learn, and improve over time |
| **Trust Layer** | Ed25519-based agent identity, peer reputation scoring, memory attestation, and delegation chains | Agents verify before they trust |
| **Policy Engine** | Real-time YAML/Rego policy enforcement on every memory access and agent action, with tamper-evident audit logs and GDPR tools | Enterprises can deploy agents with confidence |

---

## Key Features

### 🔮 Six-Type Memory Taxonomy

Acumius models memory the way cognitive science does — not as a blob of vectors, but as distinct types with different lifecycles:

| Type | Stores | Backend | TTL |
|------|--------|---------|-----|
| **Working** | Active task context | Valkey | 24h default |
| **Episodic** | Past sessions and events | PostgreSQL | Persistent |
| **Semantic** | Facts, entities, relationships | PostgreSQL + pgvector | Persistent |
| **Procedural** | Successful workflows and strategies | PostgreSQL | Persistent |
| **Declarative** | Policies, preferences, constraints | PostgreSQL | Persistent |
| **Feedback** | User corrections and overrides | PostgreSQL | Persistent |

### 🔐 Verifiable Identity & Reputation

Every agent gets a cryptographically verifiable DID (`did:acumius:{pubkey}`). Reputation is earned through:
- Task completion rate
- Peer verification reports
- Memory attestation quality
- Policy compliance history

Agents with low reputation can't access sensitive namespaces or delegate to high-reputation agents.

### ⚡ Real-Time Policy Enforcement

Every API call is evaluated against policy in **< 0.1ms**. Policies control:
- Which memory types an agent can read/write
- Which namespaces an agent can access
- Maximum delegation depth and cost
- PII handling and retention rules
- Audit logging requirements

Default behavior: **DENY** (fail-closed).

### 🌐 Cross-Framework Collaboration

Connect via **MCP** (primary), **REST**, or **AG-UI** (Server-Sent Events). A LangGraph agent can write to a shared namespace, and a CrewAI agent can read from it — with full audit trails and policy enforcement.

### 🛡️ Enterprise-Ready Governance

- **Audit logs**: Tamper-evident, queryable, exportable to SIEM
- **GDPR tools**: Right-to-forget, data export, rectification, auto-expiry
- **PII redaction**: Automatic detection and bulk redaction
- **Compliance mapping**: EU AI Act, SOC 2, NIST AI RMF, HIPAA

### 🚀 Local-First, Single Binary

One Go binary. One `docker-compose up`. Runs entirely on your machine or VPC. No cloud dependency. No vendor lock-in.

---

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.25+ (for development)
- Make

### One-Command Start

```bash
git clone https://github.com/Acumius/Acumius.git
cd Acumius
make up
```

This starts:
- Acumius core service on `localhost:8080`
- PostgreSQL + pgvector on `localhost:5432`
- Valkey on `localhost:6379`
- Governance UI on `localhost:3000`

### Verify Installation

```bash
curl http://localhost:8080/health
# {"service":"acumius","status":"ok","version":"0.1.0"}
```

### Store Your First Memory

```bash
curl -X POST http://localhost:8080/v1/memory   -H "Content-Type: application/json"   -H "X-API-Key: your-api-key"   -d '{
    "type": "semantic",
    "namespace": "demo",
    "content": {"fact": "The capital of France is Paris"},
    "metadata": {"source": "manual", "confidence": 1.0}
  }'
```

### Python SDK Quick Start

```bash
pip install acumius
```

```python
from acumius import AcumiusClient

client = AcumiusClient(base_url="http://localhost:8080", api_key="your-api-key")

# Store memory
client.memory.store(
    type="semantic",
    namespace="my-project",
    content={"fact": "Q3 revenue grew 23%"},
    metadata={"source": "earnings_call", "confidence": 0.95}
)

# Search across types
results = client.memory.search(
    query="revenue growth",
    types=["semantic", "episodic"],
    namespace="my-project"
)

# Check agent reputation before delegating
rep = client.trust.get_reputation("did:acumius:analyst-001")
if rep.score > 700:
    print("Safe to delegate analysis task")
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    AGENT ECOSYSTEM                           │
│     LangGraph · CrewAI · AutoGen · OpenAI · Custom           │
└────────────────────────┬────────────────────────────────────┘
                         │  MCP · REST · AG-UI
┌────────────────────────▼────────────────────────────────────┐
│                    ACUMIUS CORE (Go)                        │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   MEMORY    │  │   TRUST     │  │      POLICY         │  │
│  │   ENGINE    │  │   LAYER     │  │      ENGINE         │  │
│  │             │  │             │  │                     │  │
│  │ • 6 types   │  │ • Agent DID │  │ • YAML/Rego rules   │  │
│  │ • Namespace │  │ • Reputation│  │ • Real-time enforce │  │
│  │ • Temporal  │  │ • Attestation│ │ • Audit log         │  │
│  │ • Distill   │  │ • Delegation│  │ • GDPR redaction    │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              STORAGE ROUTING LAYER                       ││
│  │  Working → Valkey  │  Semantic → PostgreSQL + pgvector  ││
│  │  Episodic → PostgreSQL  │  Procedural → PostgreSQL      ││
│  │  Declarative → PostgreSQL  │  Feedback → PostgreSQL     ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│              GOVERNANCE UI (Next.js 15)                      │
│     Memory Explorer · Agent Directory · Policy Editor        │
│     Audit Log · GDPR Tools · Dashboard                       │
└─────────────────────────────────────────────────────────────┘
```

**Design Principles:**
- **Local-first** — runs on-device, nothing leaves unless configured
- **Protocol-neutral** — MCP primary, REST and AG-UI secondary
- **Modular backends** — swap Valkey for Redis, pgvector for Chroma without changing agent code
- **Fail-closed** — policy errors result in DENY, not ALLOW
- **Transparent** — agents call `store` and `retrieve`; routing is invisible

---

## Repository Structure

```
acumius/
├── cmd/acumius/              # Go service entrypoint
├── internal/
│   ├── api/                  # HTTP handlers (REST + health)
│   ├── memory/               # Memory engine (types, store, router, search, distiller)
│   ├── trust/                # Trust layer (identity, registry, reputation, attestation)
│   ├── policy/               # Policy engine (parser, evaluator, cache)
│   ├── audit/                # Audit logging
│   ├── gdpr/                 # GDPR tools (redaction, export, forget)
│   ├── storage/              # Database connections (PostgreSQL, Valkey)
│   └── config/               # Configuration management
├── protocol/
│   ├── mcp/                  # MCP server implementation
│   ├── rest/                 # REST API (OpenAPI spec)
│   └── agui/                 # AG-UI SSE server
├── pkg/sdk/                  # Shared types for SDKs
├── governance-ui/            # Next.js dashboard
├── sdk/
│   ├── python/               # PyPI package
│   └── typescript/           # npm package
├── adapters/
│   ├── langgraph/            # LangGraph drop-in adapter
│   ├── crewai/               # CrewAI adapter
│   ├── autogen/              # AutoGen adapter
│   └── openai/               # OpenAI Agents SDK adapter
├── examples/
│   └── cross_framework_demo/ # The killer demo
├── migrations/               # PostgreSQL schema migrations
├── bench/                    # Benchmark CLI and certification suite
├── docs/
│   ├── adr/                  # Architecture Decision Records
│   ├── architecture.md       # Full architecture documentation
│   ├── api_spec.md           # API specification
│   ├── schema.md             # Database schema reference
│   ├── policy_spec.md        # Policy engine specification
│   └── trust_spec.md         # Trust layer specification
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

---

## Roadmap

| Milestone | Target | Status |
|-----------|--------|--------|
| **v0.1** — Foundation | Store/retrieve Working + Episodic memory, MCP server, LangGraph adapter, basic Governance UI | 🔄 In Progress |
| **v0.2** — Full Memory | All 6 memory types, CrewAI + AutoGen adapters, Python + TypeScript SDKs, authentication | 📅 Planned |
| **v0.3** — Intelligence | Distillation worker, policy editor UI, GDPR tools, full docs site | 📅 Planned |
| **v0.4** — Trust | Agent identity, reputation scoring, attestation, delegation chains | 📅 Planned |
| **v0.5** — Governance | Advanced policy engine, audit dashboard, compliance mapping, SIEM export | 📅 Planned |
| **v1.0** — Production | Performance benchmarks, security audit, community launch, certification badge | 📅 Planned |

See [ROADMAP.md](docs/ROADMAP.md) for detailed phase breakdowns and [Milestones](https://github.com/Acumius/Acumius/milestones) for live tracking.

---

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/architecture.md) | System design, component interactions, data flow |
| [API Specification](docs/api_spec.md) | Full REST API reference with OpenAPI 3.1 spec |
| [Database Schema](docs/schema.md) | PostgreSQL schema, indexes, partitioning strategy |
| [Policy Engine](docs/policy_spec.md) | Policy language, evaluation semantics, examples |
| [Trust Layer](docs/trust_spec.md) | Identity, reputation, attestation, delegation |
| [SDK Guide](docs/sdk_guide.md) | Python and TypeScript SDK usage |
| [Contributing](CONTRIBUTING.md) | Development workflow, conventions, track ownership |
| [ADRs](docs/adr/) | Architecture Decision Records |

---

## Tech Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Core engine | Go 1.25 | Performance, single binary, easy deployment |
| Primary storage | PostgreSQL 16 + pgvector | Reliability, ACID, vector search |
| Working memory | Valkey 8 | Speed, open-source, Redis-compatible |
| API protocols | MCP 2024-11-05, REST, AG-UI SSE | Framework-neutral access |
| Policy language | YAML + Rego (OPA) | Declarative + industry standard |
| Identity | Ed25519 + DID | Fast, secure, no blockchain dependency |
| Governance UI | Next.js 15 + shadcn/ui | Modern React, accessible, fast |
| SDKs | Python 3.10+, TypeScript 5.0+ | Ecosystem coverage |
| Packaging | Docker, single binary, GHCR | Multiple deployment options |

---

## Benchmarks

Acumius is designed for performance:

| Metric | Target | Notes |
|--------|--------|-------|
| Policy evaluation | < 0.1ms p50 | Cached, compiled rule tree |
| Memory store | < 5ms p99 | Async embedding generation |
| Semantic search | < 50ms p99 | pgvector IVFFlat + full-text hybrid |
| Concurrent agents | 10,000+ | Goroutine-per-connection model |
| Cold start | < 2s | Single binary, no JVM warmup |

See [bench/](bench/) for benchmark suite and [docs/benchmarks.md](docs/benchmarks.md) for methodology.

---

## Contributing

We welcome contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development environment setup
- Branch and PR conventions
- Track ownership and responsibilities
- Code of Conduct

**Good first issues** are tagged [`good first issue`](https://github.com/Acumius/Acumius/issues?q=label%3A%22good+first+issue%22).

---

## Community

- 💬 [Discord](https://discord.gg/acumius)
- 🐦 [Twitter/X](https://x.com/acumius)
- 📧 [Security reports](mailto:security@acumius.dev)
- 📋 [Discussions](https://github.com/Acumius/Acumius/discussions)

---

## Security

See [SECURITY.md](SECURITY.md) for:
- Vulnerability reporting process
- Security boundaries and threat model
- Known limitations and recommended layered defenses

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<p align="center">
  <i>Built for the agent ecosystem. Governed by the community.</i><br>
  <i>Memory that agents can share — and verify.</i>
</p>
