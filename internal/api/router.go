package api

import (
	"net/http"

	"github.com/Acumius/Acumius/internal/storage"
)

// NewMux creates the base HTTP router for the Acumius service.
func NewMux(pg *storage.PostgresStore, vk *storage.ValkeyStore) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterRoutes(mux, pg, vk)
	return mux
}

// RegisterRoutes wires all currently supported HTTP routes.
func RegisterRoutes(mux *http.ServeMux, pg *storage.PostgresStore, vk *storage.ValkeyStore) {
	mux.HandleFunc("GET /health", HealthHandler(pg, vk))

	// Trust Layer Endpoints
	mux.HandleFunc("POST /api/trust/agents", RegisterAgentHandler(pg))
	mux.HandleFunc("GET /api/trust/agents", ListAgentsHandler(pg))
	mux.HandleFunc("GET /api/trust/agents/{did}", GetAgentHandler(pg))
	mux.HandleFunc("POST /api/trust/attestations", CreateAttestationHandler(pg))
	mux.HandleFunc("GET /api/trust/attestations/memory/{memory_id}", ListAttestationsHandler(pg))
	mux.HandleFunc("POST /api/trust/verifications", CreateVerificationHandler(pg))
	mux.HandleFunc("POST /api/trust/verifications/{id}/result", SubmitVerificationResultHandler(pg))
}
