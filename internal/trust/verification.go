package trust

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type VerificationStore struct {
	db *sql.DB
}

func NewVerificationStore(db *sql.DB) *VerificationStore {
	return &VerificationStore{db: db}
}

// SelectVerifier finds a suitable verifier agent excluding the target agent.
// It prioritizes agents with high reputation score and returns the chosen agent's DID.
func (s *VerificationStore) SelectVerifier(ctx context.Context, targetDID string, minReputation int) (string, error) {
	query := `
		SELECT did
		FROM agents
		WHERE did != $1 AND reputation_score >= $2
		ORDER BY reputation_score DESC, RANDOM()
		LIMIT 1
	`
	var verifierDID string
	err := s.db.QueryRowContext(ctx, query, targetDID, minReputation).Scan(&verifierDID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no suitable verifier found")
		}
		return "", fmt.Errorf("failed to select verifier: %w", err)
	}
	return verifierDID, nil
}

// CreateVerification schedules a new peer verification between a verifier and a target.
func (s *VerificationStore) CreateVerification(ctx context.Context, targetDID, verifierDID string) (*Verification, error) {
	if targetDID == "" || verifierDID == "" {
		return nil, errors.New("target and verifier DIDs cannot be empty")
	}
	if targetDID == verifierDID {
		return nil, errors.New("verifier cannot be the same as target agent")
	}

	createdAt := time.Now()
	query := `
		INSERT INTO verifications (target_did, verifier_did, status, created_at)
		VALUES ($1, $2, 'pending', $3)
		RETURNING id
	`
	var id string
	err := s.db.QueryRowContext(ctx, query, targetDID, verifierDID, createdAt).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create verification: %w", err)
	}

	return &Verification{
		ID:          id,
		TargetDID:   targetDID,
		VerifierDID: verifierDID,
		Status:      "pending",
		CreatedAt:   createdAt,
	}, nil
}

// SubmitVerificationResult processes the result of a verification.
// If approved (success=true): Status = "completed", and verifier gets +50 reputation points.
// If rejected (success=false): Status = "failed", and target gets a policy_violation event (-100 reputation score).
func (s *VerificationStore) SubmitVerificationResult(ctx context.Context, id string, success bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Fetch verification details
	var targetDID, verifierDID, currentStatus string
	queryFetch := `
		SELECT target_did, verifier_did, status
		FROM verifications
		WHERE id = $1
	`
	err = tx.QueryRowContext(ctx, queryFetch, id).Scan(&targetDID, &verifierDID, &currentStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("verification not found")
		}
		return fmt.Errorf("failed to fetch verification: %w", err)
	}

	if currentStatus != "pending" {
		return fmt.Errorf("verification is already in status: %s", currentStatus)
	}

	// 2. Update verification status
	newStatus := "failed"
	if success {
		newStatus = "completed"
	}

	queryUpdate := `
		UPDATE verifications
		SET status = $1
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, queryUpdate, newStatus, id)
	if err != nil {
		return fmt.Errorf("failed to update verification status: %w", err)
	}

	// 3. Update agent reputation scores
	repEngine := NewReputationEngine(nil) // Use in-memory calculation logic

	if success {
		// Verifier receives +50 reputation points for a completed verification
		var score int
		err = tx.QueryRowContext(ctx, "SELECT reputation_score FROM agents WHERE did = $1", verifierDID).Scan(&score)
		if err == nil {
			newScore := repEngine.db /* avoid field reference since we just clamp */
			_ = newScore
			verifierScore := score + 50
			if verifierScore > 1000 {
				verifierScore = 1000
			}
			_, err = tx.ExecContext(ctx, "UPDATE agents SET reputation_score = $1, last_active_at = NOW() WHERE did = $2", verifierScore, verifierDID)
			if err != nil {
				return fmt.Errorf("failed to update verifier reputation: %w", err)
			}
		}
	} else {
		// Target agent gets policy violation (-100 reputation score)
		var score int
		err = tx.QueryRowContext(ctx, "SELECT reputation_score FROM agents WHERE did = $1", targetDID).Scan(&score)
		if err == nil {
			targetScore := score - 100
			if targetScore < 0 {
				targetScore = 0
			}
			_, err = tx.ExecContext(ctx, "UPDATE agents SET reputation_score = $1, last_active_at = NOW() WHERE did = $2", targetScore, targetDID)
			if err != nil {
				return fmt.Errorf("failed to update target reputation: %w", err)
			}

			// Log reputation event for target agent
			queryEvent := `
				INSERT INTO reputation_events (agent_did, event_type, score_change, created_at)
				VALUES ($1, 'policy_violation', -100, NOW())
			`
			_, err = tx.ExecContext(ctx, queryEvent, targetDID)
			if err != nil {
				return fmt.Errorf("failed to log policy violation event: %w", err)
			}
		}
	}

	return tx.Commit()
}

// Get retrieves a verification by ID.
func (s *VerificationStore) Get(ctx context.Context, id string) (*Verification, error) {
	query := `
		SELECT id, target_did, verifier_did, status, created_at
		FROM verifications
		WHERE id = $1
	`
	var v Verification
	err := s.db.QueryRowContext(ctx, query, id).Scan(&v.ID, &v.TargetDID, &v.VerifierDID, &v.Status, &v.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("verification not found")
		}
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}
	return &v, nil
}
