package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Acumius/Acumius/internal/audit"
	"github.com/Acumius/Acumius/internal/policy"
)

// contextKey is a private type for context keys defined in this package, to
// avoid collisions with keys defined in other packages.
type contextKey string

const agentDIDContextKey contextKey = "agent_did"

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// AuditMiddleware logs all incoming requests.
func AuditMiddleware(logger *audit.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, r)
			duration := time.Since(start)

			agentDID := r.Header.Get("X-Agent-DID")
			if agentDID == "" {
				agentDID = "anonymous"
			}

			if logger != nil {
				logger.Log(audit.Event{
					AgentDID: agentDID,
					Action:   r.Method,
					Resource: r.URL.Path,
					Allowed:  rw.status < 400,
					Metadata: map[string]interface{}{
						"duration_ms": duration.Milliseconds(),
						"status":      rw.status,
					},
				})
			}
		})
	}
}

// PolicyMiddleware enforces policies on incoming requests.
func PolicyMiddleware(evaluator *policy.Evaluator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract AgentDID (simplistic auth for now, later Trust Layer)
			agentDID := r.Header.Get("X-Agent-DID")
			if agentDID == "" {
				http.Error(w, "Unauthorized: missing X-Agent-DID", http.StatusUnauthorized)
				return
			}

			// Map HTTP Method and Path to Action and Resource
			action := mapHTTPMethodToAction(r.Method)
			resourceType, namespace := mapPathToResource(r.URL.Path)

			req := policy.Request{
				AgentDID:   agentDID,
				Action:     action,
				MemoryType: resourceType,
				Namespace:  namespace,
			}

			result, err := evaluator.Evaluate(r.Context(), req)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !result.Allowed {
				http.Error(w, "Forbidden: "+result.Reason, http.StatusForbidden)
				return
			}

			// Pass the DID to the context
			ctx := context.WithValue(r.Context(), agentDIDContextKey, agentDID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func mapHTTPMethodToAction(method string) string {
	switch method {
	case http.MethodGet:
		return "read"
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return "write"
	case http.MethodDelete:
		return "delete"
	default:
		return "unknown"
	}
}

func mapPathToResource(path string) (string, string) {
	// e.g. /api/memory -> memory, "" (converted to "default")
	// e.g. /api/trust/agents -> trust, agents
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		ns := strings.Join(parts[2:], "/")
		if ns == "" {
			ns = "default"
		}
		return parts[1], ns
	}
	return "unknown", "default"
}
