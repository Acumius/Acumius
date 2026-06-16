package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuditMiddleware(t *testing.T) {
	// A basic test to ensure middleware doesn't crash without a logger
	mw := AuditMiddleware(nil)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
}

func TestMapHTTPMethodToAction(t *testing.T) {
	if action := mapHTTPMethodToAction(http.MethodGet); action != "read" {
		t.Errorf("expected read, got %s", action)
	}
	if action := mapHTTPMethodToAction(http.MethodPost); action != "write" {
		t.Errorf("expected write, got %s", action)
	}
}

func TestMapPathToResource(t *testing.T) {
	res, ns := mapPathToResource("/api/memory")
	if res != "memory" || ns != "default" {
		t.Errorf("expected memory, default, got %s, %s", res, ns)
	}
}
