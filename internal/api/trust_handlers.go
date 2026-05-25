package api

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Acumius/Acumius/internal/storage"
	"github.com/Acumius/Acumius/internal/trust"
	"github.com/mr-tron/base58/base58"
)

// RegisterAgentHandler handles POST /api/trust/agents
func RegisterAgentHandler(pg *storage.PostgresStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PublicKey string `json:"public_key"` // base58 representation
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		registry := trust.NewAgentRegistry(pg.DB())

		var pubKey ed25519.PublicKey
		var privKey ed25519.PrivateKey
		var err error

		if req.PublicKey == "" {
			// Auto-generate keypair
			pubKey, privKey, err = trust.GenerateKeypair()
			if err != nil {
				http.Error(w, "Failed to generate keys: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Parse base58 public key
			decoded, err := base58.Decode(req.PublicKey)
			if err != nil {
				http.Error(w, "Invalid base58 public key", http.StatusBadRequest)
				return
			}
			if len(decoded) != ed25519.PublicKeySize {
				http.Error(w, "Invalid public key size", http.StatusBadRequest)
				return
			}
			pubKey = ed25519.PublicKey(decoded)
		}

		did := trust.GenerateDID(pubKey)
		agent := &trust.Agent{
			DID:             did,
			PublicKey:       pubKey,
			ReputationScore: 500,
		}

		if err := registry.Register(r.Context(), agent); err != nil {
			http.Error(w, "Failed to register agent: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		var resp struct {
			DID        string `json:"did"`
			PublicKey  string `json:"public_key"`
			PrivateKey string `json:"private_key,omitempty"`
		}
		resp.DID = did
		resp.PublicKey = base58.Encode(pubKey)
		if privKey != nil {
			resp.PrivateKey = base58.Encode(privKey)
		}

		_ = json.NewEncoder(w).Encode(resp)
	}
}

// GetAgentHandler handles GET /api/trust/agents/{did}
func GetAgentHandler(pg *storage.PostgresStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		did := r.PathValue("did")
		if did == "" {
			http.Error(w, "DID is required", http.StatusBadRequest)
			return
		}

		registry := trust.NewAgentRegistry(pg.DB())
		agent, err := registry.Get(r.Context(), did)
		if err != nil {
			if errors.Is(err, trust.ErrAgentNotFound) {
				http.Error(w, "Agent not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agent)
	}
}

// ListAgentsHandler handles GET /api/trust/agents
func ListAgentsHandler(pg *storage.PostgresStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := trust.NewAgentRegistry(pg.DB())
		agents, err := registry.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agents)
	}
}

// CreateAttestationHandler handles POST /api/trust/attestations
func CreateAttestationHandler(pg *storage.PostgresStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			MemoryID  string `json:"memory_id"`
			AgentDID  string `json:"agent_did"`
			Signature string `json:"signature"` // base64 encoded signature
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		sigBytes, err := base64.StdEncoding.DecodeString(req.Signature)
		if err != nil {
			http.Error(w, "Invalid base64 signature", http.StatusBadRequest)
			return
		}

		store := trust.NewAttestationStore(pg.DB())
		att := &trust.Attestation{
			MemoryID:  req.MemoryID,
			AgentDID:  req.AgentDID,
			Signature: sigBytes,
		}

		if err := store.Save(r.Context(), att); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(att)
	}
}

// ListAttestationsHandler handles GET /api/trust/attestations/memory/{memory_id}
func ListAttestationsHandler(pg *storage.PostgresStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		memoryID := r.PathValue("memory_id")
		if memoryID == "" {
			http.Error(w, "Memory ID is required", http.StatusBadRequest)
			return
		}

		store := trust.NewAttestationStore(pg.DB())
		list, err := store.ListByMemory(r.Context(), memoryID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}
}

// CreateVerificationHandler handles POST /api/trust/verifications
func CreateVerificationHandler(pg *storage.PostgresStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TargetDID string `json:"target_did"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		store := trust.NewVerificationStore(pg.DB())

		// Select a verifier with minimum score of 500
		verifierDID, err := store.SelectVerifier(r.Context(), req.TargetDID, 500)
		if err != nil {
			http.Error(w, "No suitable verifier found: "+err.Error(), http.StatusFailedDependency)
			return
		}

		v, err := store.CreateVerification(r.Context(), req.TargetDID, verifierDID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(v)
	}
}

// SubmitVerificationResultHandler handles POST /api/trust/verifications/{id}/result
func SubmitVerificationResultHandler(pg *storage.PostgresStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Verification ID is required", http.StatusBadRequest)
			return
		}

		var req struct {
			Success bool `json:"success"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		store := trust.NewVerificationStore(pg.DB())
		if err := store.SubmitVerificationResult(r.Context(), id, req.Success); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}
}
