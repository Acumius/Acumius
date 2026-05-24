# Acumius

**Universal memory and infrastructure layer for AI agents.**

> Agents shouldn't start from zero every time. Acumius gives them a brain that lasts.

---

## What is Acumius?

Acumius is an open-source, local-first **agent infrastructure framework**. It starts with a universal memory system — a standalone service that any AI agent, regardless of which framework it's built on, can connect to and gain persistent, structured memory across sessions.

Think of Acumius as the infrastructure layer that sits between your agent and everything it has ever learned. It routes, stores, retrieves, and governs memories across multiple types, while exposing one clean API to the agent — no matter which database, framework, or protocol sits underneath.

Memory is the first module. As new agent infrastructure problems are identified, new modules will be added to the framework following the same design principles.

---

## The Problem

Every time an AI agent session ends, its memory is gone. The next session starts from scratch — the same questions, the same mistakes, the same context-building all over again.

Today's memory solutions make this worse, not better:

- They are framework-locked (LangGraph memory only works in LangGraph)
- They only implement one or two memory types, usually just vector search
- They are cloud-only with no self-hosting option
- They give you no visibility into what the agent has stored or learned
- They provide no governance — no way to inspect, correct, or delete memories

No existing open-source project functions as **neutral, embeddable memory infrastructure** that any agent can plug into.

---

## How Acumius Solves It

### One service, every framework

Acumius runs as a standalone service. Any agent — LangGraph, AutoGen, CrewAI, or a custom system — connects via a standard interface (MCP, REST, or AG-UI) and gains persistent memory without building it themselves.

### Six memory types, one API

Agents need different kinds of memory to behave intelligently over time. Acumius implements all six as first-class types through a single API:

| Type | What it stores |
|---|---|
| **Working** | Active context for the current task |
| **Episodic** | Past sessions and events with timestamps |
| **Semantic** | Facts, entities, and relationships |
| **Procedural** | Successful workflows and learned strategies |
| **Declarative** | Policies, preferences, and hard constraints |
| **Feedback** | User corrections that override existing memories |

Acumius routes each memory to the correct backend, enforces policies, and returns merged results — invisibly to the agent.

### Governance and transparency

A first-class web UI lets you inspect every memory, view a timeline of changes, approve or reject memories before they become permanent, bulk-redact PII, and audit every read, write, and delete operation.

### Policy engine

A declarative policy file controls what agents are allowed to remember — blocking secrets, expiring PII, enforcing retention rules — before problems happen.

### Memory distillation

A background worker periodically compresses old episodic memories into compact semantic facts. Agents genuinely improve over time instead of just accumulating raw logs.

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Agent Ecosystem                   │
│         LangGraph · AutoGen · CrewAI · Custom       │
└────────────────────────┬────────────────────────────┘
                         │  MCP · REST · AG-UI
┌────────────────────────▼────────────────────────────┐
│                  Acumius Framework                  │
│                                                     │
│   ┌─────────────┐  ┌──────────────┐  ┌──────────┐  │
│   │   Memory    │  │    Policy    │  │Distiller │  │
│   │   Router    │  │    Engine    │  │ Worker   │  │
│   └─────────────┘  └──────────────┘  └──────────┘  │
│                                                     │
│   ┌─────────────────────────────────────────────┐   │
│   │            Storage Routing Layer            │   │
│   │  Working → Valkey  |  Semantic → pgvector   │   │
│   │  Episodic → Postgres  |  Procedural → JSONB │   │
│   └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

**Design principles:**
- **Local-first** — everything runs on-device, nothing leaves the machine unless configured
- **Protocol-neutral** — supports MCP, AG-UI SSE, and REST so agents don't care which framework they're built in
- **Modular backends** — swap Valkey for Redis, pgvector for Chroma, without changing agent code
- **Transparent to agents** — agents call `store` and `retrieve`, routing is invisible

---

## Current Status

**Phase 0 — Alignment (in progress)**

The project is in its foundational phase. Architecture decisions have been made, the shared memory schema is defined, and the first implementation work has begun.

- [x] GitHub organization and repository created
- [x] Architecture Decision Records (ADRs) written and merged
- [x] Shared memory schema defined
- [x] Go service scaffold with health endpoint and CI pipeline — [PR #16](../../pull/16)
- [ ] Docker Compose baseline (Postgres + pgvector + Valkey)
- [ ] Core memory store/retrieve endpoints
- [ ] MCP server wrapping the core API
- [ ] Governance UI v0.1

Follow the [Phase 0 milestone](../../milestone/1) and [v0.1 milestone](../../milestone/2) for live progress.

---

## Repository Structure

```
acumius/
├── core/              # Go memory engine — storage, routing, distillation
├── protocol/          # MCP, AG-UI, and REST protocol servers
├── adapters/          # LangGraph, AutoGen, CrewAI drop-in adapters
├── governance-ui/     # Next.js dashboard — inspect, edit, govern memories
├── sdk/
│   ├── python/        # PyPI package
│   └── typescript/    # npm package
├── bench/             # Benchmark CLI and certification suite
├── docs/              # Architecture docs and ADRs
├── migrations/        # PostgreSQL schema migrations
├── examples/          # Working example agents
└── docker-compose.yml # One-command local development environment
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Core engine | Go |
| Primary storage | PostgreSQL + pgvector |
| Working memory | Valkey |
| Agent protocol | MCP (primary), REST, AG-UI |
| Governance UI | Next.js 15 + shadcn/ui |
| SDKs | Python (PyPI), TypeScript (npm) |

See [`docs/decisions/`](docs/decisions/) for the full Architecture Decision Records explaining each choice.

---

## Getting Started

> Local development setup is coming in Phase 1 once the Docker Compose baseline is merged. Watch [issue #15](../../issues/15) and the [v0.1 milestone](../../milestone/2) for updates.

Once available, the full local environment will start with:

```bash
git clone https://github.com/Acumius/Acumius.git
cd Acumius
make up
```

For now, you can run the core service directly:

```bash
go run ./cmd/acumius
curl http://localhost:8080/health
# {"service":"acumius","status":"ok"}
```

---

## Contributing

Acumius is being built in the open as a team project. Contributions are welcome.

**Before you start:**
1. Read [`CONTRIBUTING.md`](CONTRIBUTING.md) for the full contribution workflow
2. Read [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the shared memory schema and API contracts
3. Browse the [open issues](../../issues) — issues labeled [`good first issue`](../../issues?q=label%3A%22good+first+issue%22) are a good starting point

**Branch and PR conventions:**
- Branch from `main` using `feat/<your-name>/<short-description>`
- Every PR needs at least one review before merging
- Commits follow [Conventional Commits](https://www.conventionalcommits.org/) format
- PRs should reference the issue they close: `Closes #N`

**Track ownership:**

| Track | Scope |
|---|---|
| `track:core` | Go memory engine, storage, distillation |
| `track:protocol` | MCP server, REST API, SDKs |
| `track:adapters` | Framework adapters, DX, examples |
| `track:governance` | React UI, policy engine |
| `track:bench` | Benchmark CLI, CI quality gates |

---

## Roadmap

| Milestone | Goal |
|---|---|
| **v0.1** | Store + retrieve working and episodic memories, MCP server, LangGraph adapter, basic Governance UI |
| **v0.2** | All 6 memory types, AutoGen + CrewAI adapters, TypeScript and Python SDKs published, authentication |
| **v0.3** | Distillation worker, policy editor UI, GDPR tools, full docs site |
| **v1.0** | Production-ready, community launch, certification badge |

Full phase-by-phase plan is tracked via [GitHub Milestones](../../milestones).

---

## Vision

Acumius is building toward becoming the **OpenTelemetry of agent memory** — a vendor-neutral, community-governed infrastructure standard that agent frameworks adopt, enterprises trust, and developers extend.

Memory is the first module. As the agent ecosystem surfaces new infrastructure problems, Acumius will grow to address them — always with the same principles: local-first, protocol-neutral, open, and transparent.

---

## License

MIT — see [`LICENSE`](LICENSE) for details.

---

*Acumius is an open-source initiative. Built for the agent ecosystem, governed by the community.*
## Architecture Decision Records (ADRs)

Acumius records major architectural and process decisions in [`docs/adr/`](docs/adr/).

- [ADR-0001: Core Engine Language (Go)](docs/adr/ADR-0001-core-engine-language.md)
- [ADR Index and Process](docs/adr/README.md)
