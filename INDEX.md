# Acumius Implementation Package

This package contains all planning documents, specifications, and AntiGravity prompts for building Acumius.

## File Structure

```
acumius/
├── README.md                    # Refreshed project README (industry standard)
├── PHASES.md                    # Master implementation phases document
├── CONTRIBUTING.md              # Contribution guidelines
├── docs/
│   ├── architecture.md          # Full system architecture specification
│   ├── api_spec.md              # REST API specification
│   ├── schema.md                # Database schema reference
│   ├── policy_spec.md           # Policy engine specification
│   └── trust_spec.md            # Trust layer specification
└── prompts/
    ├── phase_0_foundation.md    # Docker Compose, scaffold completion
    ├── phase_1_memory_engine.md # 6-type memory, hybrid search, routing
    ├── phase_2_trust_layer.md   # Identity, reputation, attestation
    ├── phase_3_policy_engine.md # Real-time policy, audit, GDPR
    ├── phase_4_protocol_layer.md # MCP, REST, AG-UI, auth, rate limit
    ├── phase_5_governance_ui.md # Next.js dashboard
    ├── phase_6_sdks_adapters.md # Python/TS SDKs, framework adapters
    └── phase_7_demo_launch.md   # Demo, benchmarks, launch
```

## How to Use

### For Project Planning
Read `PHASES.md` for the full 24-week roadmap with deliverables, acceptance criteria, and handoff prompts.

### For Architecture Decisions
Read `docs/architecture.md` for system design, data flows, and component interactions.

### For API Development
Read `docs/api_spec.md` for endpoint definitions and `docs/schema.md` for database design.

### For AntiGravity (Claude Code)
Feed the prompt files in order:
1. `prompts/phase_0_foundation.md` → Complete infrastructure
2. `prompts/phase_1_memory_engine.md` → Build memory system
3. `prompts/phase_2_trust_layer.md` → Add identity and trust
4. `prompts/phase_3_policy_engine.md` → Add governance
5. `prompts/phase_4_protocol_layer.md` → Add protocols
6. `prompts/phase_5_governance_ui.md` → Build dashboard
7. `prompts/phase_6_sdks_adapters.md` → Build SDKs
8. `prompts/phase_7_demo_launch.md` → Launch

Each prompt is self-contained with context, tasks, acceptance criteria, and constraints.

## Project Vision

Acumius is the Agent Collaboration Fabric — the only infrastructure that combines structured multi-type memory with verifiable identity, reputation, and real-time policy enforcement, enabling agents from different frameworks to actually collaborate.

**Tagline:** *"The trust layer for the agent internet."*

**Positioning:**
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
*Acumius Team*
