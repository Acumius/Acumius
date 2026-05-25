package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Acumius/Acumius/internal/memory"
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

	// Trust Layer Endpoints (Placeholder until Trust Layer is fully wired)
	mux.HandleFunc("POST /api/trust/agents", RegisterAgentHandler(pg))
	mux.HandleFunc("GET /api/trust/agents", ListAgentsHandler(pg))
	mux.HandleFunc("GET /api/trust/agents/{did}", GetAgentHandler(pg))
	mux.HandleFunc("POST /api/trust/attestations", CreateAttestationHandler(pg))
	mux.HandleFunc("GET /api/trust/attestations/memory/{memory_id}", ListAttestationsHandler(pg))
	mux.HandleFunc("POST /api/trust/verifications", CreateVerificationHandler(pg))
	mux.HandleFunc("POST /api/trust/verifications/{id}/result", SubmitVerificationResultHandler(pg))

	// Initialize Memory Engine components
	if pg != nil && vk != nil {
		dummyEmbedder := func(ctx context.Context, text string) ([]float32, error) {
			return make([]float32, 1536), nil
		}
		searcher := memory.NewHybridSearcher(pg.DB(), dummyEmbedder)
		pgStore := memory.NewPostgresStore(pg.DB(), searcher)
		vkStore := memory.NewValkeyStore(vk.Client(), 24*time.Hour)

		// Pass nil for TrustPort in Phase 1 (Trust integration fully operational in Phase 2)
		memRouter := memory.NewRouter(pgStore, vkStore, nil)
		memHandler := NewMemoryHandler(memRouter)

		// Memory Engine Endpoints
		mux.HandleFunc("POST /api/memory", memHandler.StoreMemory)
		mux.HandleFunc("POST /api/memory/search", memHandler.SearchMemory)
	}
}
