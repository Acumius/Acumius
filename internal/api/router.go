package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Acumius/Acumius/internal/audit"
	"github.com/Acumius/Acumius/internal/gdpr"
	"github.com/Acumius/Acumius/internal/mcp"
	"github.com/Acumius/Acumius/internal/memory"
	"github.com/Acumius/Acumius/internal/policy"
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
		// Initialize Trust components (Placeholder)
		auditLogger := audit.NewLogger(pg.DB())
		policyStore := policy.NewPostgresStore(pg.DB())
		policyCache := policy.NewValkeyCache(vk.Client())
		evaluator := policy.NewEvaluator(policyCache, policyStore)
		gdprService := gdpr.NewService(pg.DB())

		// Create Handlers
		memRouter := memory.NewRouter(pgStore, vkStore, nil)
		memHandler := NewMemoryHandler(memRouter)
		policyHandler := NewPolicyHandler(policyStore, evaluator)
		auditHandler := NewAuditHandler(auditLogger)
		gdprHandler := NewGDPRHandler(gdprService)

		// Create Middlewares
		auditMW := AuditMiddleware(auditLogger)
		policyMW := PolicyMiddleware(evaluator)

		// Mount Memory Engine Endpoints with Middleware
		mux.Handle("POST /api/memory", auditMW(policyMW(http.HandlerFunc(memHandler.StoreMemory))))
		mux.Handle("POST /api/memory/search", auditMW(policyMW(http.HandlerFunc(memHandler.SearchMemory))))

		// Mount Policy Endpoints
		mux.Handle("POST /api/policies", auditMW(http.HandlerFunc(policyHandler.CreatePolicy)))
		mux.Handle("POST /api/policies/evaluate", auditMW(http.HandlerFunc(policyHandler.EvaluatePolicy)))

		// Mount Audit Endpoints
		mux.Handle("GET /api/audit", auditMW(http.HandlerFunc(auditHandler.QueryAudit)))

		// Mount GDPR Endpoints
		mux.Handle("POST /api/gdpr/forget", auditMW(http.HandlerFunc(gdprHandler.RightToForget)))
		mux.Handle("GET /api/gdpr/export", auditMW(http.HandlerFunc(gdprHandler.ExportData)))
		mux.Handle("PUT /api/gdpr/rectify", auditMW(http.HandlerFunc(gdprHandler.RectifyData)))

		// Initialize MCP Server
		mcpServer := mcp.NewServer(pgStore, evaluator)
		sseServer := mcpServer.SSEServer()
		// Mark3Labs MCP server usually registers `/sse` and `/message` inside SSEServer.
		// Alternatively, just expose its ServeHTTP to the base path for MCP.
		mux.Handle("/mcp/sse", sseServer.SSEHandler())
		mux.Handle("/mcp/message", sseServer.MessageHandler())
	}
}
