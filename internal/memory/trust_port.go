package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TrustPort is the interface the Memory Engine uses from the Trust Layer.
type TrustPort interface {
	// VerifySignature checks if a memory write is signed by the claimed agent.
	VerifySignature(ctx context.Context, agentDID string, memoryID uuid.UUID, content []byte, signature []byte) error

	// GetAgentReputation returns the reputation score for an agent.
	GetAgentReputation(ctx context.Context, agentDID string) (int, error)

	// IsAgentActive checks if an agent is active (not suspended/revoked).
	IsAgentActive(ctx context.Context, agentDID string) (bool, error)

	// GetAttestationsForMemory returns all attestations for a memory.
	GetAttestationsForMemory(ctx context.Context, memoryID uuid.UUID) ([]AttestationRecord, error)

	// RecordEvent records a reputation event.
	RecordEvent(ctx context.Context, agentDID string, eventType string, delta int, description string) error
}

// AttestationRecord is the full attestation data from the trust layer.
type AttestationRecord struct {
	ID        uuid.UUID
	AgentDID  string
	MemoryID  uuid.UUID
	Claim     string
	Signature []byte
	Timestamp time.Time
}
