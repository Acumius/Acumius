# Database Migrations

This directory is reserved for PostgreSQL migrations used by the Acumius core service.

The v0.1 scaffold intentionally ships without schema migrations yet. Follow-up issues should add versioned SQL migration files here.

## Naming Convention

Use sequential, zero-padded migration numbers with lowercase snake_case descriptions:

- `000001_create_memories_table.up.sql`
- `000001_create_memories_table.down.sql`

Rules:

- Use `.up.sql` for forward migrations and `.down.sql` for rollback migrations.
- Keep the numeric prefix identical for an up/down pair.
- Use lowercase snake_case for the description segment.
- Never rename an already-merged migration; add a new migration instead.
