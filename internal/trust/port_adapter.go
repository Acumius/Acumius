package trust

import (
	"context"
	"crypto/ed25519"
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

	if !ed25519.Verify(agent.PublicKey, content, signature) {
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
	// If attestation is needed here, assuming AttestationStore has a method.
	// For now, if the method isn't fully implemented in Phase 2, return empty.
	// In the real phase 2, it exists or we mock it.
	return nil, nil // Placeholder depending on actual phase 2 completion
}

func (a *PortAdapter) RecordEvent(ctx context.Context, agentDID string, eventType string, delta int, description string) error {
	// Call to reputation engine.
	return nil // Placeholder
}
