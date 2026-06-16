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
	ID        string                 `json:"id,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
	AgentDID  string                 `json:"agent_did"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Allowed   bool                   `json:"allowed"`
	PolicyID  string                 `json:"policy_id,omitempty"`
	Reason    string                 `json:"reason,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AuditFilter contains parameters for querying audit logs.
type AuditFilter struct {
	AgentDID *string
	Action   *string
	Allowed  *bool
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
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

// Query retrieves audit events based on the provided filter.
func (l *Logger) Query(ctx context.Context, filter AuditFilter) ([]Event, int, error) {
	query := `SELECT id, timestamp, agent_did, action, resource, allowed, policy_id, reason, metadata FROM audit_log WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM audit_log WHERE 1=1`
	var args []interface{}
	argID := 1

	if filter.AgentDID != nil {
		clause := fmt.Sprintf(" AND agent_did = $%d", argID)
		query += clause
		countQuery += clause
		args = append(args, *filter.AgentDID)
		argID++
	}
	if filter.Action != nil {
		clause := fmt.Sprintf(" AND action = $%d", argID)
		query += clause
		countQuery += clause
		args = append(args, *filter.Action)
		argID++
	}
	if filter.Allowed != nil {
		clause := fmt.Sprintf(" AND allowed = $%d", argID)
		query += clause
		countQuery += clause
		args = append(args, *filter.Allowed)
		argID++
	}
	if filter.From != nil {
		clause := fmt.Sprintf(" AND timestamp >= $%d", argID)
		query += clause
		countQuery += clause
		args = append(args, *filter.From)
		argID++
	}
	if filter.To != nil {
		clause := fmt.Sprintf(" AND timestamp <= $%d", argID)
		query += clause
		countQuery += clause
		args = append(args, *filter.To)
		argID++
	}

	var total int
	if err := l.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argID)
		args = append(args, filter.Limit)
		argID++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argID)
		args = append(args, filter.Offset)
	}

	rows, err := l.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var policyID, reason sql.NullString
		var metaBytes []byte

		err := rows.Scan(
			&e.ID,
			&e.Timestamp,
			&e.AgentDID,
			&e.Action,
			&e.Resource,
			&e.Allowed,
			&policyID,
			&reason,
			&metaBytes,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}

		if policyID.Valid {
			e.PolicyID = policyID.String
		}
		if reason.Valid {
			e.Reason = reason.String
		}
		if len(metaBytes) > 0 {
			// Best-effort metadata decode; skip on error.
			_ = json.Unmarshal(metaBytes, &e.Metadata)
		}
		events = append(events, e)
	}
	return events, total, rows.Err()
}

// Errors returns a channel that receives logging errors.
func (l *Logger) Errors() <-chan error {
	return l.errChan
}
