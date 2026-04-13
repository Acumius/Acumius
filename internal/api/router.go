package api

import "net/http"

// NewMux creates the base HTTP router for the Acumius service.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	RegisterRoutes(mux)
	return mux
}

// RegisterRoutes wires all currently supported HTTP routes.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", HealthHandler())
}
