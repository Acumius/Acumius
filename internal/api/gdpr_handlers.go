package api

import (
	"encoding/json"
	"net/http"

	"github.com/Acumius/Acumius/internal/gdpr"
)

// GDPRHandler groups HTTP handlers for GDPR operations.
type GDPRHandler struct {
	service *gdpr.Service
}

// NewGDPRHandler creates a new GDPRHandler.
func NewGDPRHandler(service *gdpr.Service) *GDPRHandler {
	return &GDPRHandler{
		service: service,
	}
}

// RightToForget handles POST /api/gdpr/right-to-forget.
func (h *GDPRHandler) RightToForget(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentDID string `json:"agent_did"`
		Confirm  bool   `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentDID == "" || !req.Confirm {
		http.Error(w, "missing agent_did or confirmation", http.StatusBadRequest)
		return
	}

	if err := h.service.Forget(r.Context(), req.AgentDID); err != nil {
		http.Error(w, "forget operation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ExportData handles POST /api/gdpr/export.
func (h *GDPRHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentDID string `json:"agent_did"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentDID == "" {
		http.Error(w, "missing agent_did", http.StatusBadRequest)
		return
	}

	data, err := h.service.Export(r.Context(), req.AgentDID)
	if err != nil {
		http.Error(w, "export failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="export.json"`)
	_, _ = w.Write(data)
}

// RectifyData handles POST /api/gdpr/rectify.
func (h *GDPRHandler) RectifyData(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MemoryID   string          `json:"memory_id"`
		Correction json.RawMessage `json:"correction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.MemoryID == "" || len(req.Correction) == 0 {
		http.Error(w, "missing memory_id or correction", http.StatusBadRequest)
		return
	}

	if err := h.service.Rectify(r.Context(), req.MemoryID, req.Correction); err != nil {
		http.Error(w, "rectify failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "rectified"})
}
