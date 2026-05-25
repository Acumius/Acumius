# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| v1.0.x | Full support |
| v0.5.x | Security fixes only |
| v0.4.x | End of life |
| < v0.4 | End of life |

## Reporting a Vulnerability

Please do NOT open public issues for security vulnerabilities.

Instead, email: security@acumius.dev

We will:
1. Acknowledge receipt within 24 hours
2. Provide initial assessment within 72 hours
3. Coordinate disclosure timeline
4. Credit you in the advisory (unless you prefer anonymity)

## Security Model

### Threat Boundaries

- UNTRUSTED: Agent code (any framework)
- TRUSTED: Acumius binary (Go) — Policy engine, Authentication, Audit logging
- TRUSTED: PostgreSQL + Valkey (run in same VPC, TLS required)

### OWASP Agentic Top 10 Coverage

| Risk | Acumius Control |
|------|-----------------|
| ASI-01 Agent Goal Hijack | Policy engine blocks unauthorized goal changes |
| ASI-02 Tool Misuse | Namespace ACL + policy enforcement |
| ASI-03 Identity Abuse | Ed25519 DID + reputation |
| ASI-04 Supply Chain | SBOM, signed releases |
| ASI-05 Code Execution | Out of scope — use E2B/Blaxel |
| ASI-06 Memory Poisoning | Attestation + audit trail |
| ASI-07 Unsafe Comms | TLS + signature verification |
| ASI-08 Cascading Failures | v1.1 (SRE layer) |
| ASI-09 Human-Agent Trust | Full flight recorder |
| ASI-10 Rogue Agents | Policy revocation + reputation decay |

### Known Limitations

1. Process isolation: Acumius runs in the same process as agent code. For production, run each agent in a separate container.
2. Network security: Inter-agent communication assumes TLS. Do not run over untrusted networks without TLS.
3. Key management: Agent private keys are managed by the agent, not Acumius. If an agent leaks its key, rotate immediately.
4. Embedding models: Semantic search relies on external embedding models. Verify their outputs before trusting search results.

## Security Tools

| Tool | Coverage |
|------|----------|
| CodeQL | Go SAST |
| Gitleaks | Secret scanning |
| Dependabot | Dependency updates |
| gosec | Go security linter |
| Fuzzing | Policy engine, parsers |

## Disclosure Timeline

1. Day 0: Vulnerability reported
2. Day 1: Acknowledgment + initial assessment
3. Day 7: Fix developed and tested
4. Day 14: Fix released + advisory published
5. Day 21: Public disclosure (if not already)

Last updated: 2026-05-24
