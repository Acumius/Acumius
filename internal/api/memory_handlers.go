package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Acumius/Acumius/internal/memory"
)

type MemoryHandler struct {
	router *memory.Router
}

func NewMemoryHandler(router *memory.Router) *MemoryHandler {
	return &MemoryHandler{router: router}
}

type StoreMemoryRequest struct {
	AgentDID  string            `json:"agent_did"`
	Type      memory.MemoryType `json:"type"`
	Namespace string            `json:"namespace"`
	Content   json.RawMessage   `json:"content"`
	Metadata  memory.Metadata   `json:"metadata"`
	Signature []byte            `json:"signature"`
	TTL       time.Duration     `json:"ttl,omitempty"`
}

func (h *MemoryHandler) StoreMemory(w http.ResponseWriter, r *http.Request) {
	var req StoreMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	storeReq := memory.StoreRequest{
		Type:      req.Type,
		Namespace: req.Namespace,
		AgentDID:  req.AgentDID,
		Content:   req.Content,
		Metadata:  req.Metadata,
	}

	storedMem, err := h.router.StoreWithVerification(r.Context(), storeReq, req.Signature)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(storedMem)
}

type SearchMemoryRequest struct {
	Query      string              `json:"query"`
	Types      []memory.MemoryType `json:"types,omitempty"`
	Namespaces []string            `json:"namespaces,omitempty"`
	AgentDID   string              `json:"agent_did,omitempty"`
	Limit      int                 `json:"limit,omitempty"`
}

func (h *MemoryHandler) SearchMemory(w http.ResponseWriter, r *http.Request) {
	var req SearchMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}

	searchQuery := memory.SearchQuery{
		Query:      req.Query,
		Types:      req.Types,
		Namespaces: req.Namespaces,
		AgentDID:   req.AgentDID,
		Limit:      req.Limit,
	}

	results, err := h.router.Search(r.Context(), searchQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
