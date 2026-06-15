package trust

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"
)

type ReputationFactors struct {
	CompletionRate     float64 `json:"completion_rate"`     // 0.0 to 1.0
	PeerVerifications  int     `json:"peer_verifications"`  // Count of verifications completed
	MemoryAttestations int     `json:"memory_attestations"` // Count of attestations signed
	PolicyViolations   int     `json:"policy_violations"`   // Count of violations
	DisputesLost       int     `json:"disputes_lost"`       // Count of lost disputes
	DaysInactive       int     `json:"days_inactive"`       // Days since last active
}

// CalculateReputation computes the score using the formula:
// 500 + completion_rate*200 + peer_verifications*50 + memory_attestations*25 - policy_violations*100 - disputes_lost*150 - days_inactive*1
// Score is clamped between 0 and 1000.
func CalculateReputation(factors ReputationFactors) int {
	score := 500.0

	score += factors.CompletionRate * 200.0
	score += float64(factors.PeerVerifications * 50)
	score += float64(factors.MemoryAttestations * 25)
	score -= float64(factors.PolicyViolations * 100)
	score -= float64(factors.DisputesLost * 150)
	score -= float64(factors.DaysInactive * 1)

	rounded := int(math.Round(score))

	if rounded < 0 {
		return 0
	}
	if rounded > 1000 {
		return 1000
	}
	return rounded
}

type ReputationEngine struct {
	db *sql.DB
}

func NewReputationEngine(db *sql.DB) *ReputationEngine {
	return &ReputationEngine{db: db}
}

// LogEvent records a reputation event in the database, and recalculates/updates the agent's overall score.
func (e *ReputationEngine) LogEvent(ctx context.Context, agentDID string, eventType string, scoreChange int) error {
	// Start transaction
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Insert reputation event
	queryEvent := `
		INSERT INTO reputation_events (agent_did, event_type, score_change, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err = tx.ExecContext(ctx, queryEvent, agentDID, eventType, scoreChange)
	if err != nil {
		return fmt.Errorf("failed to log event: %w", err)
	}

	// Fetch current score
	var currentScore int
	queryAgent := `SELECT reputation_score FROM agents WHERE did = $1`
	err = tx.QueryRowContext(ctx, queryAgent, agentDID).Scan(&currentScore)
	if err != nil {
		return fmt.Errorf("failed to get agent score: %w", err)
	}

	// Update score
	newScore := currentScore + scoreChange
	if newScore < 0 {
		newScore = 0
	}
	if newScore > 1000 {
		newScore = 1000
	}

	updateQuery := `
		UPDATE agents
		SET reputation_score = $1, last_active_at = NOW()
		WHERE did = $2
	`
	_, err = tx.ExecContext(ctx, updateQuery, newScore, agentDID)
	if err != nil {
		return fmt.Errorf("failed to update agent reputation: %w", err)
	}

	return tx.Commit()
}

// FetchFactorsAndRecalculate queries the DB to construct ReputationFactors, computes the score, and updates the agent.
func (e *ReputationEngine) FetchFactorsAndRecalculate(ctx context.Context, agentDID string) (int, error) {
	// Query to fetch counts
	// For completion rate, we check reputation_events for completions vs starts, or mock a default of 1.0.
	// For this phase, we count verifications from verifications table where verifier_did = agentDID and status = 'completed'.
	// We count attestations from attestations table where agent_did = agentDID.
	// We count policy violations from reputation_events where event_type = 'policy_violation'.
	// We count disputes lost from reputation_events where event_type = 'dispute_lost'.
	// We count days inactive from agents table (NOW() - last_active_at).

	var factors ReputationFactors
	var lastActive time.Time

	queryAgent := `SELECT last_active_at FROM agents WHERE did = $1`
	err := e.db.QueryRowContext(ctx, queryAgent, agentDID).Scan(&lastActive)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch agent activity: %w", err)
	}

	// Compute days inactive
	days := int(time.Since(lastActive).Hours() / 24)
	if days < 0 {
		days = 0
	}
	factors.DaysInactive = days

	// Default completion rate to 1.0 (100%) for active agents
	factors.CompletionRate = 1.0

	// Count verifications completed
	queryVerifications := `SELECT COUNT(*) FROM verifications WHERE verifier_did = $1 AND status = 'completed'`
	err = e.db.QueryRowContext(ctx, queryVerifications, agentDID).Scan(&factors.PeerVerifications)
	if err != nil {
		return 0, fmt.Errorf("failed to count verifications: %w", err)
	}

	// Count attestations
	queryAttestations := `SELECT COUNT(*) FROM attestations WHERE agent_did = $1`
	err = e.db.QueryRowContext(ctx, queryAttestations, agentDID).Scan(&factors.MemoryAttestations)
	if err != nil {
		return 0, fmt.Errorf("failed to count attestations: %w", err)
	}

	// Count violations from reputation_events
	queryViolations := `SELECT COUNT(*) FROM reputation_events WHERE agent_did = $1 AND event_type = 'policy_violation'`
	err = e.db.QueryRowContext(ctx, queryViolations, agentDID).Scan(&factors.PolicyViolations)
	if err != nil {
		return 0, fmt.Errorf("failed to count violations: %w", err)
	}

	// Count disputes lost from reputation_events
	queryDisputes := `SELECT COUNT(*) FROM reputation_events WHERE agent_did = $1 AND event_type = 'dispute_lost'`
	err = e.db.QueryRowContext(ctx, queryDisputes, agentDID).Scan(&factors.DisputesLost)
	if err != nil {
		return 0, fmt.Errorf("failed to count disputes: %w", err)
	}

	score := CalculateReputation(factors)

	// Update in DB
	updateQuery := `UPDATE agents SET reputation_score = $1 WHERE did = $2`
	_, err = e.db.ExecContext(ctx, updateQuery, score, agentDID)
	if err != nil {
		return 0, fmt.Errorf("failed to save recalculated score: %w", err)
	}

	return score, nil
}
