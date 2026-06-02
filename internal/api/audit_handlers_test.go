package api_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Acumius/Acumius/internal/api"
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

func TestAuditHandler_QueryAudit(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	// Clean table
	_, _ = db.Exec("TRUNCATE audit_log RESTART IDENTITY CASCADE")

	logger := audit.NewLogger(db)
	handler := api.NewAuditHandler(logger)

	agentDID := "did:test:audit-handler"

	logger.Log(audit.Event{
		AgentDID: agentDID,
		Action:   "read",
		Resource: "memory:test",
		Allowed:  true,
		PolicyID: "pol-1",
	})

	// Wait for async log to insert
	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/api/audit?agent_did="+agentDID+"&action=read", nil)
	rr := httptest.NewRecorder()

	handler.QueryAudit(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	eventsRaw, ok := resp["events"].([]interface{})
	require.True(t, ok)
	assert.Len(t, eventsRaw, 1)

	eventMap := eventsRaw[0].(map[string]interface{})
	assert.Equal(t, agentDID, eventMap["agent_did"])
	assert.Equal(t, "read", eventMap["action"])
	assert.Equal(t, "memory:test", eventMap["resource"])
}
