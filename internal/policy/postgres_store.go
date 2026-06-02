package policy

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// PostgresStore implements Store for persistent policies.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a Postgres-backed store.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{
		db: db,
	}
}

// GetPolicies retrieves policies for an agent.
func (s *PostgresStore) GetPolicies(ctx context.Context, agentDID string) ([]Policy, error) {
	query := `SELECT id, agent_did, version, content, created_at, updated_at 
              FROM policies WHERE agent_did = $1`

	rows, err := s.db.QueryContext(ctx, query, agentDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		var p Policy
		var contentJSON []byte
		err := rows.Scan(&p.ID, &p.AgentDID, &p.Version, &contentJSON, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(contentJSON, &p.Content); err != nil {
			return nil, err
		}

		policies = append(policies, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return policies, nil
}

// SavePolicy saves a policy.
func (s *PostgresStore) SavePolicy(ctx context.Context, p Policy) error {
	contentJSON, err := p.Content.ToJSON()
	if err != nil {
		return err
	}

	query := `INSERT INTO policies (id, agent_did, version, content, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6)
              ON CONFLICT (id) DO UPDATE SET 
                  version = EXCLUDED.version,
                  content = EXCLUDED.content,
                  updated_at = EXCLUDED.updated_at`

	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	_, err = s.db.ExecContext(ctx, query, p.ID, p.AgentDID, p.Version, contentJSON, p.CreatedAt, p.UpdatedAt)
	return err
}
