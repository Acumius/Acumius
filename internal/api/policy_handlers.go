package api

import (
	"encoding/json"
	"net/http"

	"github.com/Acumius/Acumius/internal/policy"
)

// PolicyHandler groups HTTP handlers for policy operations.
type PolicyHandler struct {
	store     policy.Store
	evaluator *policy.Evaluator
}

// NewPolicyHandler creates a new PolicyHandler.
func NewPolicyHandler(store policy.Store, evaluator *policy.Evaluator) *PolicyHandler {
	return &PolicyHandler{
		store:     store,
		evaluator: evaluator,
	}
}

// CreatePolicy handles POST /api/policies.
func (h *PolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req policy.Policy
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentDID == "" {
		http.Error(w, "missing agent_did", http.StatusBadRequest)
		return
	}

	if err := h.store.SavePolicy(r.Context(), req); err != nil {
		http.Error(w, "failed to save policy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created", "id": req.ID})
}

// EvaluatePolicy handles POST /api/policies/evaluate.
func (h *PolicyHandler) EvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	var req policy.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentDID == "" || req.Action == "" {
		http.Error(w, "missing agent_did or action", http.StatusBadRequest)
		return
	}

	result, err := h.evaluator.Evaluate(r.Context(), req)
	if err != nil {
		http.Error(w, "evaluation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
