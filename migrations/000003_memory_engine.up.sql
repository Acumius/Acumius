-- migrations/000003_memory_engine.up.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- MEMORIES (unified table for 5 persistent types)
CREATE TABLE IF NOT EXISTS memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type TEXT NOT NULL CHECK (type IN ('episodic', 'semantic', 'procedural', 'declarative', 'feedback')),
    namespace TEXT NOT NULL,
    agent_did TEXT NOT NULL REFERENCES agents(did),
    content JSONB NOT NULL,
    embedding VECTOR(1536),                    -- pgvector for semantic search
    metadata JSONB DEFAULT '{}',
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    distilled_from UUID REFERENCES memories(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ                    -- soft delete for GDPR
);

-- Full-text search index on content
CREATE INDEX IF NOT EXISTS idx_memories_fts 
    ON memories USING GIN(to_tsvector('english', content::text));

-- Vector index for semantic search (IVFFlat for balance of speed/recall)
CREATE INDEX IF NOT EXISTS idx_memories_embedding 
    ON memories USING ivfflat (embedding vector_cosine_ops) 
    WITH (lists = 100);                       -- tune based on dataset size

-- Common query indexes
CREATE INDEX IF NOT EXISTS idx_memories_namespace ON memories(namespace);
CREATE INDEX IF NOT EXISTS idx_memories_agent ON memories(agent_did);
CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at);

-- Composite index for filtered queries
CREATE INDEX IF NOT EXISTS idx_memories_ns_type_created 
    ON memories(namespace, type, created_at) 
    WHERE deleted_at IS NULL;

-- Partial index for active (non-deleted) memories
CREATE INDEX IF NOT EXISTS idx_memories_active 
    ON memories(namespace, type, agent_did) 
    WHERE deleted_at IS NULL;

-- Index for temporal queries
CREATE INDEX IF NOT EXISTS idx_memories_valid 
    ON memories(valid_from, valid_until) 
    WHERE deleted_at IS NULL;

-- NAMESPACE ACL
CREATE TABLE IF NOT EXISTS namespace_acl (
    namespace TEXT NOT NULL,
    agent_did TEXT NOT NULL REFERENCES agents(did),
    permission TEXT NOT NULL CHECK (permission IN ('read', 'write', 'admin')),
    granted_by TEXT NOT NULL REFERENCES agents(did),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (namespace, agent_did, permission)
);

CREATE INDEX IF NOT EXISTS idx_acl_namespace ON namespace_acl(namespace);
CREATE INDEX IF NOT EXISTS idx_acl_agent ON namespace_acl(agent_did);

-- PII REGISTRY (GDPR compliance)
CREATE TABLE IF NOT EXISTS pii_registry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    memory_id UUID NOT NULL REFERENCES memories(id),
    pii_type TEXT NOT NULL,
    pii_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pii_memory ON pii_registry(memory_id);
CREATE INDEX IF NOT EXISTS idx_pii_hash ON pii_registry(pii_hash);

-- AUDIT LOG (append-only, partitioned by month)
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    agent_did TEXT NOT NULL,
    action TEXT NOT NULL,
    resource TEXT NOT NULL,
    allowed BOOLEAN NOT NULL,
    policy_id TEXT,
    reason TEXT,
    metadata JSONB DEFAULT '{}',
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create initial monthly partitions
CREATE TABLE IF NOT EXISTS audit_log_y2026m05 PARTITION OF audit_log
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE IF NOT EXISTS audit_log_y2026m06 PARTITION OF audit_log
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE INDEX IF NOT EXISTS idx_audit_agent ON audit_log(agent_did);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp);

-- POLICIES (Phase 3 Policy Engine)
CREATE TABLE IF NOT EXISTS policies (
    id TEXT PRIMARY KEY,
    agent_did TEXT REFERENCES agents(did),
    content JSONB NOT NULL,
    version TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_policies_agent ON policies(agent_did);
