package audit_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/Acumius/Acumius/internal/audit"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("ACUMIUS_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://acumius:acumius@localhost:5432/acumius?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("skipping test; database connection failed: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("skipping test; database ping failed: %v", err)
	}

	return db
}

func TestLogger_Log(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	// Clean table
	_, _ = db.Exec("TRUNCATE audit_log RESTART IDENTITY CASCADE")

	logger := audit.NewLogger(db)

	agentDID := "did:example:audit-123"

	logger.Log(audit.Event{
		AgentDID: agentDID,
		Action:   "read",
		Resource: "memory:123",
		Allowed:  true,
		PolicyID: "pol-123",
		Reason:   "allowed by policy",
		Metadata: map[string]interface{}{"ip": "127.0.0.1"},
	})

	// Wait for goroutine to insert
	time.Sleep(100 * time.Millisecond)

	// Verify insertion
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE agent_did = $1", agentDID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var allowed bool
	var action, resource string
	err = db.QueryRow("SELECT action, resource, allowed FROM audit_log WHERE agent_did = $1 LIMIT 1", agentDID).Scan(&action, &resource, &allowed)
	require.NoError(t, err)
	assert.Equal(t, "read", action)
	assert.Equal(t, "memory:123", resource)
	assert.True(t, allowed)
}
