package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Acumius/Acumius/internal/storage"
)

// HealthHandler returns a lightweight service health response.
func HealthHandler(pg *storage.PostgresStore, vk *storage.ValkeyStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		pgStatus := "disconnected"
		if pg != nil && pg.Ping() == nil {
			pgStatus = "connected"
		}

		vkStatus := "disconnected"
		if vk != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if vk.Ping(ctx) == nil {
				vkStatus = "connected"
			}
		}

		response := map[string]interface{}{
			"service": "acumius",
			"status":  "ok",
			"version": "0.1.0",
			"dependencies": map[string]string{
				"postgresql": pgStatus,
				"valkey":     vkStatus,
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
