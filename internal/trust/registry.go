package trust

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrAgentAlreadyExists = errors.New("agent already exists")
)

type AgentRegistry struct {
	db *sql.DB
}

func NewAgentRegistry(db *sql.DB) *AgentRegistry {
	return &AgentRegistry{db: db}
}

func (r *AgentRegistry) Register(ctx context.Context, agent *Agent) error {
	if agent.DID == "" {
		return errors.New("agent DID cannot be empty")
	}
	if len(agent.PublicKey) == 0 {
		return errors.New("agent public key cannot be empty")
	}

	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = time.Now()
	}
	if agent.LastActiveAt.IsZero() {
		agent.LastActiveAt = time.Now()
	}
	if agent.ReputationScore == 0 {
		agent.ReputationScore = 500 // Default score
	}

	query := `
		INSERT INTO agents (did, public_key, reputation_score, created_at, last_active_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, agent.DID, []byte(agent.PublicKey), agent.ReputationScore, agent.CreatedAt, agent.LastActiveAt)
	if err != nil {
		// Postgres unique violation code is 23505
		return fmt.Errorf("failed to register agent: %w", err)
	}

	return nil
}

func (r *AgentRegistry) Get(ctx context.Context, did string) (*Agent, error) {
	query := `
		SELECT did, public_key, reputation_score, created_at, last_active_at
		FROM agents
		WHERE did = $1
	`
	row := r.db.QueryRowContext(ctx, query, did)

	var agent Agent
	var pubKeyBytes []byte

	err := row.Scan(&agent.DID, &pubKeyBytes, &agent.ReputationScore, &agent.CreatedAt, &agent.LastActiveAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	agent.PublicKey = pubKeyBytes

	return &agent, nil
}

func (r *AgentRegistry) Update(ctx context.Context, agent *Agent) error {
	query := `
		UPDATE agents
		SET reputation_score = $1, last_active_at = $2
		WHERE did = $3
	`
	result, err := r.db.ExecContext(ctx, query, agent.ReputationScore, agent.LastActiveAt, agent.DID)
	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrAgentNotFound
	}

	return nil
}

func (r *AgentRegistry) List(ctx context.Context) ([]Agent, error) {
	query := `
		SELECT did, public_key, reputation_score, created_at, last_active_at
		FROM agents
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		var pubKeyBytes []byte
		err := rows.Scan(&agent.DID, &pubKeyBytes, &agent.ReputationScore, &agent.CreatedAt, &agent.LastActiveAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agent.PublicKey = pubKeyBytes
		agents = append(agents, agent)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return agents, nil
}
