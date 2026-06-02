package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Acumius/Acumius/internal/audit"
)

// AuditHandler groups HTTP handlers for audit operations.
type AuditHandler struct {
	logger *audit.Logger
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(logger *audit.Logger) *AuditHandler {
	return &AuditHandler{
		logger: logger,
	}
}

// QueryAudit handles GET /api/audit.
func (h *AuditHandler) QueryAudit(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var filter audit.AuditFilter

	if agentDID := q.Get("agent_did"); agentDID != "" {
		filter.AgentDID = &agentDID
	}
	if action := q.Get("action"); action != "" {
		filter.Action = &action
	}
	if allowedStr := q.Get("allowed"); allowedStr != "" {
		if allowed, err := strconv.ParseBool(allowedStr); err == nil {
			filter.Allowed = &allowed
		}
	}
	if fromStr := q.Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.From = &t
		}
	}
	if toStr := q.Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.To = &t
		}
	}

	limit := 100
	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	filter.Limit = limit

	if offsetStr := q.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	events, total, err := h.logger.Query(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to query audit logs", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
