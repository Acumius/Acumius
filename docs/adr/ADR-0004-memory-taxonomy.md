# ADR-0004: Memory Taxonomy

**Date:** 2026-04-09  
**Status:** Accepted  
**Deciders:** Core Engine track, Adapters & DX track, project maintainers

## Context

Acumius needs a stable long-term memory model that is expressive enough for advanced agent behavior, while still allowing v0.1 to ship quickly with low implementation complexity.

During Phase 0 discussions, we aligned on keeping the 5-type memory model as the architectural target, but avoiding premature over-specialization in the first prototype.

## Decision

### Architectural Model (official Acumius taxonomy)

Acumius keeps the following **5 memory types** as the long-term model:

- Working
- Episodic
- Semantic
- Procedural
- Declarative

### Operational Model (v0.1 implementation)

v0.1 will start with a **minimal practical subset**:

- Working
- Semantic
- Transitional bucket (`general`) for remaining patterns until boundaries are validated with real usage data

This means the 5-type taxonomy is authoritative at design level, while runtime specialization is phased in.

## Rationale

- Preserves architectural clarity and long-term direction.
- Reduces early complexity so teams can deliver a working system faster.
- Avoids locking incorrect boundaries before real agent traffic reveals usage patterns.
- Supports iterative schema/taxonomy refinement based on observed retrieval quality and user outcomes.

## Alternatives Considered

- **Implement all 5 fully in v0.1**
  - Rejected for v0.1 due to higher complexity, slower delivery, and risk of incorrect early boundaries.

- **Adopt a permanent simplified 2–3 type model**
  - Rejected because it underfits long-term needs for governance, timeline reasoning, and policy-specific memory behavior.

## Consequences

### Positive

- Faster prototype delivery with less schema churn risk.
- Clear migration path from simple runtime behavior to richer specialization.
- Better evidence-based taxonomy evolution.

### Tradeoffs

- Transitional bucket may temporarily mix patterns that later split into Episodic/Procedural/Declarative.
- Requires deliberate migration rules as taxonomy specialization increases.

## Basic Examples (How v0.1 Works)

### Example 1: Current task context (Working)

```json
{
  "type": "working",
  "namespace": "agent:acumius-demo",
  "content": "User is reviewing PR #42 and asked for ADR summary.",
  "metadata": {
    "ttl_minutes": 60
  }
}
```

Expected behavior:
- stored in fast short-lived working memory path
- prioritized for immediate context retrieval

### Example 2: Persistent fact (Semantic)

```json
{
  "type": "semantic",
  "namespace": "agent:acumius-demo",
  "content": "Acumius uses Go for core and JSON for v0.1 memory schema.",
  "metadata": {
    "source": "adr",
    "confidence": 0.95
  }
}
```

Expected behavior:
- indexed for durable retrieval and semantic search
- available across sessions

### Example 3: Transitional pattern (general bucket in v0.1)

```json
{
  "type": "general",
  "namespace": "agent:acumius-demo",
  "content": "Last deployment failed due to missing migration step; rerun with db migrate first.",
  "metadata": {
    "candidate_types": ["episodic", "procedural"]
  }
}
```

Expected behavior:
- accepted in v0.1 without forcing strict categorization
- later reclassified when specialization rules mature

## Evolution Example (Post-v0.1)

A `general` memory like deployment failure history can split into:

- **Episodic**: timestamped incident record
- **Procedural**: reusable runbook step

This preserves early velocity while converging toward the full 5-type architecture.

## Follow-up

- Define metrics to decide when `general` entries should migrate into Episodic/Procedural/Declarative.
- Add periodic taxonomy review checkpoints (start in v0.2).
- Introduce migration tooling to reclassify historical `general` memories safely.
