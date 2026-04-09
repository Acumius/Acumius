# ADR-0002: Shared Memory Schema Format

**Date:** 2026-04-09  
**Status:** Accepted  
**Deciders:** Core Engine track, Protocol Layer track, project maintainers

## Context

Acumius needs a shared schema format for memory payloads exchanged across core services, protocols (MCP, AG-UI, REST), adapters, and SDKs.

Phase 0 requires selecting a baseline format before deeper implementation begins. The key tradeoff is between:

- fast iteration and low friction for early development, and
- strict contracts and compact wire formats for long-term high-throughput systems.

Issue [#1](https://github.com/Acumius/Acumius/issues/1) discussion highlighted that the schema will evolve rapidly in v0.1, making early developer velocity critical.

## Decision

Acumius will use **JSON as the shared memory schema format** for v0.1.

## Rationale

- JSON is easy to inspect, debug, and evolve during early-stage architecture changes.
- It works natively across all planned interfaces (MCP, REST, AG-UI) and SDK targets.
- It avoids code-generation overhead while contracts are still moving.
- It reduces contributor friction for open-source collaborators during initial adoption.
- Team and community feedback in Phase 0 converged on JSON-first for practical iteration speed.

## Alternatives Considered

- **Protobuf from day one**
  - Rejected for v0.1 due to schema churn risk and added generation/tooling complexity during rapid iteration.
  - Better fit once message contracts stabilize and throughput optimization becomes a primary concern.

- **Dual format (JSON + Protobuf) in v0.1**
  - Rejected because maintaining two schemas early would increase surface area and testing burden.
  - Would slow core delivery by splitting effort across parallel serialization paths.

## Consequences

### Positive

- Faster iteration cycle for API and memory model evolution.
- Simpler onboarding and contributor experience.
- Easier cross-language interoperability during early SDK and adapter development.

### Tradeoffs

- Less strict compile-time schema guarantees than Protobuf.
- Potentially larger payload sizes and lower serialization efficiency at scale.

## Follow-up

- Define a versioned JSON schema contract for memory store/retrieve payloads.
- Add contract tests for schema compatibility across core and protocol layers.
- Re-evaluate Protobuf adoption in a future ADR once schema stability and throughput requirements justify migration.
