# Acumius Implementation Package

This package contains all planning documents, specifications, and AntiGravity prompts for building Acumius — The Agent Collaboration Fabric.

## Quick Start

1. Read `INDEX.md` for package overview
2. Read `README.md` for the refreshed project README
3. Read `PHASES.md` for the 24-week implementation roadmap
4. Feed `prompts/phase_*.md` to AntiGravity (Claude Code) in order

## Package Contents

### Project Documents
- `README.md` — Refreshed project README (industry standard)
- `PHASES.md` — Master implementation phases with deliverables and acceptance criteria
- `CONTRIBUTING.md` — Contribution guidelines
- `SECURITY.md` — Security policy and vulnerability reporting
- `Makefile` — Build, test, and development commands
- `Dockerfile` — Multi-stage Docker build
- `docker-compose.yml` — Full local development environment
- `.env.example` — Configuration template
- `.gitignore` — Git ignore patterns

### Specifications (`docs/`)
- `architecture.md` — Full system architecture with data flows
- `api_spec.md` — REST API endpoint definitions
- `schema.md` — PostgreSQL + Valkey database schema
- `policy_spec.md` — Policy engine language and evaluation semantics
- `trust_spec.md` — Identity, reputation, and attestation specification
- `ROADMAP.md` — Version roadmap and milestones

### AntiGravity Prompts (`prompts/`)
- `phase_0_foundation.md` — Docker Compose, env config, storage connections
- `phase_1_memory_engine.md` — 6-type memory, hybrid search, routing
- `phase_2_trust_layer.md` — Identity, reputation, attestation
- `phase_3_policy_engine.md` — Real-time policy, audit, GDPR
- `phase_4_protocol_layer.md` — MCP, REST, AG-UI, auth, rate limiting
- `phase_5_governance_ui.md` — Next.js dashboard
- `phase_6_sdks_adapters.md` — Python/TS SDKs, framework adapters
- `phase_7_demo_launch.md` — Killer demo, benchmarks, launch

## How to Use with AntiGravity

Each prompt file is self-contained with:
- **Context** — What exists and what's been built
- **Task** — Specific implementation work
- **Acceptance Criteria** — How to know it's done
- **Constraints** — What NOT to do

Feed them in order:
```bash
# Phase 0: Infrastructure
claude prompts/phase_0_foundation.md

# Phase 1: Memory Engine
claude prompts/phase_1_memory_engine.md

# Phase 2: Trust Layer
claude prompts/phase_2_trust_layer.md

# ... and so on
```

## Project Vision

**Acumius is the Agent Collaboration Fabric** — the only infrastructure that combines structured multi-type memory with verifiable identity, reputation, and real-time policy enforcement.

**Tagline:** "The trust layer for the agent internet."

**Differentiation:**
- Microsoft AGT has governance but no memory
- BasedAgents has identity but no memory
- Mem0 has memory but no trust
- Letta has memory but is siloed
- **Acumius has all three**

## Success Metrics

| Metric | Target (6 months post-launch) |
|--------|------------------------------|
| GitHub stars | 5,000+ |
| PyPI downloads | 10,000+/month |
| Registered agents | 1,000+ |
| Framework integrations | 5+ |
| Enterprise pilots | 3+ |

---

*Package generated: 2026-05-24*
*For the Acumius Team*
