# ADR-0003: Storage Baseline for v0.1

**Date:** 2026-04-09  
**Status:** Accepted  
**Deciders:** Core Engine track, Quality & Security track, project maintainers

## Context

Acumius needs a default storage stack for local development and the v0.1 milestone that supports:

- fast working-memory operations,
- durable episodic/event storage,
- semantic vector retrieval for memory search,
- local-first deployment with Docker Compose.

Phase 0 discussions in issue [#1](https://github.com/Acumius/Acumius/issues/1) and the project plan converge on a baseline of Redis + PostgreSQL + pgvector.

## Decision

Acumius will standardize on the following storage baseline for v0.1:

- **Redis** for working memory and low-latency ephemeral state
- **PostgreSQL** for durable episodic and relational metadata
- **pgvector (PostgreSQL extension)** for semantic embedding storage and vector similarity search

This baseline is the default in local Docker Compose and reference environments.

## Rationale

- Aligns with project requirements for mixed memory workloads (ephemeral + durable + vector).
- Keeps operational complexity reasonable by using a small, familiar OSS stack.
- Supports local-first self-hosting without managed cloud dependencies.
- Enables tighter integration between relational and vector queries in one primary database engine.
- Matches Phase 0 team direction and external feedback from kickoff discussions.

## Alternatives Considered

- **PostgreSQL-only (without Redis)**
  - Rejected for v0.1 because working-memory hot paths benefit from an in-memory store.
  - Would increase load and latency pressure on the primary SQL path.

- **Redis + SQLite + external vector DB**
  - Rejected due to durability and scale limitations for multi-agent concurrent workloads.
  - Adds fragmentation and migration complexity early in the project.

- **Managed cloud databases only**
  - Rejected for baseline because Acumius is local-first and should run fully self-hosted by default.

## Consequences

### Positive

- Clear and actionable default stack for contributors and CI environments.
- Better latency profile for mixed memory operations.
- Predictable path to scale while preserving local developer ergonomics.

### Tradeoffs

- Requires running multiple services locally.
- Vector index quality and latency depend on pgvector/HNSW parameter tuning.

## Follow-up

- Provide a canonical `docker-compose.yml` baseline with pinned versions.
- Add benchmark checks for vector retrieval latency targets under representative load.
- Document pgvector index tuning guidance (including `ef_construction`, `m`, and query-time settings).
- Keep backend abstractions modular so Valkey/Chroma/Qdrant alternatives can be introduced later without protocol changes.
