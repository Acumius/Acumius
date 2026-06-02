package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Event represents an access attempt or action taken by an agent.
type Event struct {
	AgentDID string                 `json:"agent_did"`
	Action   string                 `json:"action"`
	Resource string                 `json:"resource"`
	Allowed  bool                   `json:"allowed"`
	PolicyID string                 `json:"policy_id,omitempty"`
	Reason   string                 `json:"reason,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Logger handles asynchronous, fire-and-forget logging to the database.
type Logger struct {
	db      *sql.DB
	errChan chan error
}

// NewLogger creates a new audit Logger.
func NewLogger(db *sql.DB) *Logger {
	return &Logger{
		db:      db,
		errChan: make(chan error, 100), // Buffer to avoid blocking
	}
}

// Log records an event asynchronously.
func (l *Logger) Log(event Event) {
	// Fire and forget
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var metaBytes []byte
		if event.Metadata != nil {
			metaBytes, _ = json.Marshal(event.Metadata)
		} else {
			metaBytes = []byte("{}")
		}

		query := `
			INSERT INTO audit_log (agent_did, action, resource, allowed, policy_id, reason, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err := l.db.ExecContext(ctx, query,
			event.AgentDID,
			event.Action,
			event.Resource,
			event.Allowed,
			event.PolicyID,
			event.Reason,
			metaBytes,
		)

		if err != nil {
			// Send to error channel or just log to stdout in a real app if the channel is full
			select {
			case l.errChan <- fmt.Errorf("failed to insert audit log: %w", err):
			default:
				fmt.Printf("Audit log insert failed and error channel full: %v\n", err)
			}
		}
	}()
}

// Errors returns a channel that receives logging errors.
func (l *Logger) Errors() <-chan error {
	return l.errChan
}
