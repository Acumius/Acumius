# Acumius Database Schema

> **Version:** 1.0  
> **Database:** PostgreSQL 16 + pgvector  
> **Migration Tool:** golang-migrate  
> **Last Updated:** 2026-05-24

---

## Extensions

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";
```

---

## Tables

### agents

Agent identity and profile.

```sql
CREATE TABLE agents (
    did TEXT PRIMARY KEY,
    public_key BYTEA NOT NULL,
    name TEXT NOT NULL,
    capabilities TEXT[] DEFAULT '{}',
    reputation_score INT DEFAULT 500,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'revoked')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_agents_reputation ON agents(reputation_score);
CREATE INDEX idx_agents_status ON agents(status);
```

---

### memories

Unified memory table for all 6 types.

```sql
CREATE TABLE memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type TEXT NOT NULL CHECK (type IN ('working', 'episodic', 'semantic', 'procedural', 'declarative', 'feedback')),
    namespace TEXT NOT NULL,
    agent_did TEXT NOT NULL REFERENCES agents(did),
    content JSONB NOT NULL,
    embedding VECTOR(1536),
    metadata JSONB DEFAULT '{}',
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    distilled_from UUID REFERENCES memories(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

**Indexes:**
```sql
-- Full-text search
CREATE INDEX idx_memories_fts ON memories USING GIN(to_tsvector('english', content::text));

-- Vector search (IVFFlat for balance of speed/recall)
CREATE INDEX idx_memories_embedding ON memories USING ivfflat (embedding vector_cosine_ops);

-- Common query patterns
CREATE INDEX idx_memories_namespace ON memories(namespace);
CREATE INDEX idx_memories_agent ON memories(agent_did);
CREATE INDEX idx_memories_type ON memories(type);
CREATE INDEX idx_memories_created ON memories(created_at);
CREATE INDEX idx_memories_deleted ON memories(deleted_at) WHERE deleted_at IS NOT NULL;

-- Composite for filtered search
CREATE INDEX idx_memories_ns_type_created ON memories(namespace, type, created_at);
```

---

### namespace_acl

Access control for shared namespaces.

```sql
CREATE TABLE namespace_acl (
    namespace TEXT NOT NULL,
    agent_did TEXT NOT NULL REFERENCES agents(did),
    permission TEXT NOT NULL CHECK (permission IN ('read', 'write', 'admin')),
    granted_by TEXT NOT NULL REFERENCES agents(did),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (namespace, agent_did, permission)
);
```

---

### attestations

Cryptographic attestations on memories.

```sql
CREATE TABLE attestations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_did TEXT NOT NULL REFERENCES agents(did),
    memory_id UUID NOT NULL REFERENCES memories(id),
    claim TEXT NOT NULL,
    signature BYTEA NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_attestations_memory ON attestations(memory_id);
CREATE INDEX idx_attestations_agent ON attestations(agent_did);
```

---

### audit_log

Append-only audit log, partitioned by month.

```sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    agent_did TEXT NOT NULL,
    action TEXT NOT NULL,
    resource TEXT NOT NULL,
    allowed BOOLEAN NOT NULL,
    policy_id TEXT,
    reason TEXT,
    metadata JSONB DEFAULT '{}'
) PARTITION BY RANGE (timestamp);
```

**Partition Setup:**
```sql
-- Create monthly partitions
CREATE TABLE audit_log_y2026m05 PARTITION OF audit_log
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE audit_log_y2026m06 PARTITION OF audit_log
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
-- ... auto-create future partitions via cron
```

**Indexes:**
```sql
CREATE INDEX idx_audit_agent ON audit_log(agent_did);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_allowed ON audit_log(allowed);
CREATE INDEX idx_audit_timestamp ON audit_log(timestamp);
```

---

### reputation_events

Reputation change history.

```sql
CREATE TABLE reputation_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_did TEXT NOT NULL REFERENCES agents(did),
    event_type TEXT NOT NULL,
    delta INT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_rep_events_agent ON reputation_events(agent_did);
CREATE INDEX idx_rep_events_created ON reputation_events(created_at);
```

---

### policies

Policy documents.

```sql
CREATE TABLE policies (
    id TEXT PRIMARY KEY,
    agent_did TEXT REFERENCES agents(did),
    content JSONB NOT NULL,
    version TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_policies_agent ON policies(agent_did);
```

---

### pii_registry

PII tracking for GDPR right-to-forget.

```sql
CREATE TABLE pii_registry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    memory_id UUID NOT NULL REFERENCES memories(id),
    pii_type TEXT NOT NULL,
    pii_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_pii_memory ON pii_registry(memory_id);
CREATE INDEX idx_pii_hash ON pii_registry(pii_hash);
CREATE INDEX idx_pii_type ON pii_registry(pii_type);
```

---

## Valkey Schema

### Key Patterns

```
acumius:working:{namespace}:{agent_did}:{memory_id}  → JSON memory
acumius:working:{namespace}:{agent_did}:latest       → Sorted set (by timestamp)
acumius:policy:{agent_did}                           → JSON compiled policy
acumius:policy:{agent_did}:version                   → String (policy version hash)
acumius:session:{session_id}                         → JSON session metadata
acumius:rate_limit:{agent_did}:{window}              → Counter
```

### TTL Strategy

| Key Pattern | Default TTL | Configurable |
|-------------|-------------|--------------|
| Working memory | 24h | Per-namespace |
| Policy cache | 5m | Global |
| Session | Session lifetime | Per-session |
| Rate limit | Window size | Per-endpoint |

---

## Migration Strategy

**Tool:** golang-migrate

**Naming:**
```
migrations/
  000001_init.up.sql
  000001_init.down.sql
  000002_add_namespace_acl.up.sql
  000002_add_namespace_acl.down.sql
```

**Process:**
1. Migrations run automatically on service startup
2. `schema_migrations` table tracks applied migrations
3. Down migrations for rollback testing only
4. Never modify existing migration files after merge

---

## Backup & Recovery

**PostgreSQL:**
- Daily pg_dump with `--clean` flag
- Point-in-time recovery via WAL archiving (v1.1+)
- Retention: 30 days

**Valkey:**
- RDB snapshots every 15 minutes
- AOF for durability (v1.1+)
- Retention: 7 days

---

*Schema changes require ADR and migration file.*
