package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Acumius/Acumius/internal/api"
	"github.com/Acumius/Acumius/internal/gdpr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGDPRHandler_RightToForget(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	service := gdpr.NewService(db)
	handler := api.NewGDPRHandler(service)

	reqBody := map[string]interface{}{
		"agent_did": "did:test:gdpr-1",
		"confirm":   true,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/gdpr/right-to-forget", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.RightToForget(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestGDPRHandler_ExportData(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	service := gdpr.NewService(db)
	handler := api.NewGDPRHandler(service)

	reqBody := map[string]interface{}{
		"agent_did": "did:test:gdpr-2",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/gdpr/export", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ExportData(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Disposition"), "attachment")
}

func TestGDPRHandler_RectifyData(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	// Insert dummy agent to satisfy foreign key constraint
	_, err := db.Exec("INSERT INTO agents (did, public_key, reputation_score) VALUES ('did:test:gdpr-3', 'dummy_key', 100) ON CONFLICT DO NOTHING")
	require.NoError(t, err)

	// Insert dummy memory to rectify
	memoryID := "11111111-1111-1111-1111-111111111111"
	_, err = db.Exec("INSERT INTO memories (id, agent_did, type, namespace, content) VALUES ($1, 'did:test:gdpr-3', 'semantic', 'default', '{}') ON CONFLICT DO NOTHING", memoryID)
	require.NoError(t, err)

	service := gdpr.NewService(db)
	handler := api.NewGDPRHandler(service)

	reqBody := map[string]interface{}{
		"memory_id":  memoryID,
		"correction": map[string]interface{}{"corrected": true},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/gdpr/rectify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.RectifyData(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "rectified", resp["status"])
}
