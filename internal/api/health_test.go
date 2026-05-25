package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewMux(nil, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected JSON response, got %q", contentType)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	if response["status"] != "ok" {
		t.Fatalf("expected status to be ok, got %v", response["status"])
	}

	if response["service"] != "acumius" {
		t.Fatalf("expected service to be acumius, got %v", response["service"])
	}

	deps, ok := response["dependencies"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected dependencies object, got %T", response["dependencies"])
	}

	if deps["postgresql"] != "disconnected" {
		t.Fatalf("expected postgresql to be disconnected since nil was passed, got %v", deps["postgresql"])
	}
}
