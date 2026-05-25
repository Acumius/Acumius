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
}
