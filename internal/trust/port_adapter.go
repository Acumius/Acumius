package trust

import (
	"context"
	"fmt"

	"github.com/Acumius/Acumius/internal/memory"
	"github.com/google/uuid"
)

// PortAdapter implements memory.TrustPort using the actual Trust Layer.
type PortAdapter struct {
	registry    *AgentRegistry
	reputation  *ReputationEngine
	attestation *AttestationStore
}

func NewPortAdapter(registry *AgentRegistry, reputation *ReputationEngine, attestation *AttestationStore) *PortAdapter {
	return &PortAdapter{
		registry:    registry,
		reputation:  reputation,
		attestation: attestation,
	}
}

func (a *PortAdapter) VerifySignature(ctx context.Context, agentDID string, memoryID uuid.UUID, content []byte, signature []byte) error {
	agent, err := a.registry.Get(ctx, agentDID)
	if err != nil {
		return fmt.Errorf("get agent: %w", err)
	}

	if !VerifyMemorySignature(agent.PublicKey, memoryID.String(), content, signature) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (a *PortAdapter) GetAgentReputation(ctx context.Context, agentDID string) (int, error) {
	agent, err := a.registry.Get(ctx, agentDID)
	if err != nil {
		return 0, err
	}
	return agent.ReputationScore, nil
}

func (a *PortAdapter) IsAgentActive(ctx context.Context, agentDID string) (bool, error) {
	_, err := a.registry.Get(ctx, agentDID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *PortAdapter) GetAttestationsForMemory(ctx context.Context, memoryID uuid.UUID) ([]memory.AttestationRecord, error) {
	attestations, err := a.attestation.ListByMemory(ctx, memoryID.String())
	if err != nil {
		return nil, fmt.Errorf("list attestations: %w", err)
	}

	var records []memory.AttestationRecord
	for _, att := range attestations {
		id, _ := uuid.Parse(att.ID)
		memID, _ := uuid.Parse(att.MemoryID)
		records = append(records, memory.AttestationRecord{
			ID:        id,
			AgentDID:  att.AgentDID,
			MemoryID:  memID,
			Signature: att.Signature,
			Timestamp: att.CreatedAt,
		})
	}

	return records, nil
}

func (a *PortAdapter) RecordEvent(ctx context.Context, agentDID string, eventType string, delta int, description string) error {
	return a.reputation.LogEvent(ctx, agentDID, eventType, delta)
}
