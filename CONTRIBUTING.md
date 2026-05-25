# Contributing to Acumius

Thank you for your interest in Acumius! This document explains how to contribute effectively.

## Development Environment

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make
- golang-migrate (for migrations)

### Setup

```bash
git clone https://github.com/Acumius/Acumius.git
cd Acumius
cp .env.example .env
make up
```

Verify:
```bash
curl http://localhost:8080/health
```

## Workflow

### Branch Naming

```
feat/<your-name>/<short-description>
fix/<your-name>/<bug-description>
docs/<your-name>/<doc-change>
adr/<number>-<title>
```

### Commits

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(core): add semantic memory search
fix(policy): resolve cache invalidation race
docs(readme): update installation instructions
```

### Pull Requests

1. Every PR needs at least one review
2. All CI checks must pass
3. Reference issues: `Closes #N`
4. Include tests for new code
5. Update documentation if needed

## Track Ownership

| Track | Scope | Lead |
|-------|-------|------|
| `track:core` | Go memory engine, storage, distillation | @core-lead |
| `track:protocol` | MCP server, REST API, SDKs | @protocol-lead |
| `track:adapters` | Framework adapters, DX, examples | @adapters-lead |
| `track:governance` | React UI, policy engine | @ui-lead |
| `track:bench` | Benchmark CLI, CI quality gates | @bench-lead |

## Testing

```bash
make test          # Run all tests
make test-race     # Run with race detector
make bench         # Run benchmarks
make coverage      # Generate coverage report
```

## Migrations

```bash
make migrate-create name=add_new_table
# Edit migrations/00000X_add_new_table.up.sql
# Edit migrations/00000X_add_new_table.down.sql
make migrate       # Run pending migrations
```

## Code Style

- Go: `gofmt`, `go vet`, `golint`
- Error handling: always check, wrap with context
- No panics in production code
- Comments on exported functions
- Tests: table-driven, use `testify/assert`

## Security

- Report vulnerabilities to security@acumius.dev
- Do not open public issues for security bugs
- See [SECURITY.md](SECURITY.md) for details

## Community

- Join [Discord](https://discord.gg/acumius)
- Participate in [Discussions](https://github.com/Acumius/Acumius/discussions)
- Attend weekly community calls (Fridays 15:00 UTC)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
