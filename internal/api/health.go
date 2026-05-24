package api

import (
	"encoding/json"
	"net/http"
)

// HealthHandler returns a lightweight service health response.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]string{
			"status":  "ok",
			"service": "acumius",
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}
