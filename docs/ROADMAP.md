# Acumius Roadmap

> **Version:** 1.0  
> **Last Updated:** 2026-05-24  
> **Status:** Phase 0 — Foundation (In Progress)

---

## v0.1 — Foundation (Weeks 1-6)

**Goal:** Store and retrieve Working and Episodic memories via REST and MCP.

| Feature | Status | Issue |
|---------|--------|-------|
| Go service scaffold | Done | #16 |
| Docker Compose baseline | In Progress | #18 |
| PostgreSQL + pgvector schema | Planned | |
| Valkey integration | Planned | |
| Memory CRUD REST API | Planned | |
| MCP server (basic) | Planned | |
| LangGraph adapter | Planned | |
| Basic Governance UI | Planned | |
| CI/CD pipeline | Done | #16 |

**Milestone:** [v0.1](https://github.com/Acumius/Acumius/milestone/2)

---

## v0.2 — Full Memory (Weeks 7-10)

**Goal:** All 6 memory types, multi-framework adapters, published SDKs.

| Feature | Status |
|---------|--------|
| Semantic memory with embeddings | Planned |
| Procedural memory | Planned |
| Declarative memory | Planned |
| Feedback memory | Planned |
| Hybrid search (semantic + keyword) | Planned |
| CrewAI adapter | Planned |
| AutoGen adapter | Planned |
| Python SDK (PyPI) | Planned |
| TypeScript SDK (npm) | Planned |
| Authentication (API key + DID) | Planned |

---

## v0.3 — Intelligence (Weeks 11-14)

**Goal:** Memory distillation, policy editor, GDPR tools.

| Feature | Status |
|---------|--------|
| Distillation worker (episodic to semantic) | Planned |
| Policy editor UI | Planned |
| GDPR right-to-forget | Planned |
| Data export | Planned |
| Rectification | Planned |
| Full documentation site | Planned |
| Benchmark suite | Planned |

---

## v0.4 — Trust (Weeks 15-18)

**Goal:** Agent identity, reputation, attestation, delegation.

| Feature | Status |
|---------|--------|
| Ed25519 identity generation | Planned |
| Agent registration | Planned |
| Reputation scoring | Planned |
| Memory attestation | Planned |
| Peer verification | Planned |
| Delegation chains | Planned |
| Trust ceilings | Planned |

---

## v0.5 — Governance (Weeks 19-22)

**Goal:** Advanced policy, audit dashboard, compliance.

| Feature | Status |
|---------|--------|
| Advanced policy engine (Rego/OPA) | Planned |
| Audit dashboard with SIEM export | Planned |
| Compliance mapping (EU AI Act, SOC 2, HIPAA) | Planned |
| Real-time policy violation alerts | Planned |
| Agent lifecycle management | Planned |

---

## v1.0 — Production (Weeks 23-24)

**Goal:** Production-ready, community launch, certification.

| Feature | Status |
|---------|--------|
| Security audit | Planned |
| Performance benchmarks | Planned |
| Load testing | Planned |
| Documentation complete | Planned |
| Community launch | Planned |
| Acumius Certified badge program | Planned |

---

## Post-v1.0

| Version | Focus | Key Features |
|---------|-------|-------------|
| **v1.1** | Scale | Connection pooling, read replicas, horizontal scaling |
| **v1.2** | Advanced Trust | Delegation chains, trust ceilings, dispute arbitration |
| **v1.3** | Marketplace | Agent discovery, capability advertising, task posting |
| **v1.4** | Economics | Payment rails, escrow, micro-transactions between agents |
| **v1.5** | Enterprise | SSO/SAML, RBAC, advanced compliance, SIEM integrations |
| **v2.0** | Agent Kernel | Full OS abstraction — process scheduling, I/O, network stack |

---

## Legend

| Symbol | Meaning |
|--------|---------|
| Done | Complete |
| In Progress | Currently being worked on |
| Planned | Scheduled for upcoming sprint |
| Blocked | Waiting on dependency |

---

Track live progress via GitHub Milestones
