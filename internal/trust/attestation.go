package trust

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SignMemory signs the memory metadata (memoryID + content) using the agent's private key.
func SignMemory(privKey ed25519.PrivateKey, memoryID string, content []byte) ([]byte, error) {
	if len(privKey) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	payload := generatePayload(memoryID, content)
	return ed25519.Sign(privKey, payload), nil
}

// VerifyMemorySignature verifies the signature using the agent's public key.
func VerifyMemorySignature(pubKey ed25519.PublicKey, memoryID string, content []byte, signature []byte) bool {
	if len(pubKey) != ed25519.PublicKeySize {
		return false
	}
	payload := generatePayload(memoryID, content)
	return ed25519.Verify(pubKey, payload, signature)
}

func generatePayload(memoryID string, content []byte) []byte {
	return []byte(memoryID + ":" + string(content))
}

type AttestationStore struct {
	db *sql.DB
}

func NewAttestationStore(db *sql.DB) *AttestationStore {
	return &AttestationStore{db: db}
}

// Save stores the attestation in the database and updates the agent's reputation.
func (s *AttestationStore) Save(ctx context.Context, attestation *Attestation) error {
	if attestation.MemoryID == "" {
		return errors.New("memory ID cannot be empty")
	}
	if attestation.AgentDID == "" {
		return errors.New("agent DID cannot be empty")
	}
	if len(attestation.Signature) == 0 {
		return errors.New("signature cannot be empty")
	}

	if attestation.CreatedAt.IsZero() {
		attestation.CreatedAt = time.Now()
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO attestations (memory_id, agent_did, signature, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, query, attestation.MemoryID, attestation.AgentDID, attestation.Signature, attestation.CreatedAt).Scan(&attestation.ID)
	if err != nil {
		return fmt.Errorf("failed to insert attestation: %w", err)
	}

	// Recalculate reputation score
	// Note: Creating an attestation gives +25 reputation points (clamped to 1000)
	var currentScore int
	queryAgent := `SELECT reputation_score FROM agents WHERE did = $1`
	err = tx.QueryRowContext(ctx, queryAgent, attestation.AgentDID).Scan(&currentScore)
	if err != nil {
		return fmt.Errorf("failed to get agent score: %w", err)
	}

	newScore := currentScore + 25
	if newScore > 1000 {
		newScore = 1000
	}

	updateAgent := `
		UPDATE agents
		SET reputation_score = $1, last_active_at = NOW()
		WHERE did = $2
	`
	_, err = tx.ExecContext(ctx, updateAgent, newScore, attestation.AgentDID)
	if err != nil {
		return fmt.Errorf("failed to update agent reputation: %w", err)
	}

	return tx.Commit()
}

// Get retrieves an attestation by its ID.
func (s *AttestationStore) Get(ctx context.Context, id string) (*Attestation, error) {
	query := `
		SELECT id, memory_id, agent_did, signature, created_at
		FROM attestations
		WHERE id = $1
	`
	row := s.db.QueryRowContext(ctx, query, id)

	var att Attestation
	err := row.Scan(&att.ID, &att.MemoryID, &att.AgentDID, &att.Signature, &att.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("attestation not found")
		}
		return nil, fmt.Errorf("failed to get attestation: %w", err)
	}

	return &att, nil
}

// ListByMemory retrieves all attestations for a given memory.
func (s *AttestationStore) ListByMemory(ctx context.Context, memoryID string) ([]Attestation, error) {
	query := `
		SELECT id, memory_id, agent_did, signature, created_at
		FROM attestations
		WHERE memory_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list attestations: %w", err)
	}
	defer rows.Close()

	var attestations []Attestation
	for rows.Next() {
		var att Attestation
		err := rows.Scan(&att.ID, &att.MemoryID, &att.AgentDID, &att.Signature, &att.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attestation: %w", err)
		}
		attestations = append(attestations, att)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return attestations, nil
}
