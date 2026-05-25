CREATE TABLE agents (
    did TEXT PRIMARY KEY,
    public_key BYTEA NOT NULL,
    reputation_score INTEGER NOT NULL DEFAULT 500,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE reputation_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_did TEXT NOT NULL REFERENCES agents(did) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    score_change INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE attestations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memory_id UUID NOT NULL,
    agent_did TEXT NOT NULL REFERENCES agents(did) ON DELETE CASCADE,
    signature BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_did TEXT NOT NULL REFERENCES agents(did) ON DELETE CASCADE,
    verifier_did TEXT NOT NULL REFERENCES agents(did) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reputation_events_agent_did ON reputation_events(agent_did);
CREATE INDEX idx_attestations_memory_id ON attestations(memory_id);
CREATE INDEX idx_attestations_agent_did ON attestations(agent_did);
CREATE INDEX idx_verifications_target_did ON verifications(target_did);
CREATE INDEX idx_verifications_verifier_did ON verifications(verifier_did);
