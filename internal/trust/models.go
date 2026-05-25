package trust

import (
	"crypto/ed25519"
	"time"
)

type Agent struct {
	DID             string            `json:"did"`
	PublicKey       ed25519.PublicKey `json:"public_key"`
	ReputationScore int               `json:"reputation_score"`
	CreatedAt       time.Time         `json:"created_at"`
	LastActiveAt    time.Time         `json:"last_active_at"`
}

type ReputationEvent struct {
	ID          string    `json:"id"`
	AgentDID    string    `json:"agent_did"`
	EventType   string    `json:"event_type"`
	ScoreChange int       `json:"score_change"`
	CreatedAt   time.Time `json:"created_at"`
}

type Attestation struct {
	ID        string    `json:"id"`
	MemoryID  string    `json:"memory_id"`
	AgentDID  string    `json:"agent_did"`
	Signature []byte    `json:"signature"`
	CreatedAt time.Time `json:"created_at"`
}

type Verification struct {
	ID          string    `json:"id"`
	TargetDID   string    `json:"target_did"`
	VerifierDID string    `json:"verifier_did"`
	Status      string    `json:"status"` // e.g., "pending", "completed"
	CreatedAt   time.Time `json:"created_at"`
}
