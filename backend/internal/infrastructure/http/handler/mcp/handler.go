package mcp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
)

// Handler handles MCP tool HTTP endpoints.
type Handler struct {
	mcpReg *mcp.Registry
}

// NewHandler creates a new MCP HTTP handler.
func NewHandler(mcpReg *mcp.Registry) *Handler {
	return &Handler{mcpReg: mcpReg}
}

// Routes registers MCP routes on Chi router.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/adapters", h.ListAdapters)
	r.Post("/execute", h.ExecuteAction)
}

// AdapterInfo represents adapter description.
type AdapterInfo struct {
	Name             string           `json:"name"`
	Healthy          bool             `json:"healthy"`
	SupportedActions []mcp.ActionType `json:"supported_actions"`
}

// ListAdapters returns all registered adapters and their health status.
func (h *Handler) ListAdapters(w http.ResponseWriter, r *http.Request) {
	if h.mcpReg == nil {
		respondJSON(w, http.StatusOK, []AdapterInfo{})
		return
	}

	names := h.mcpReg.List()
	var list []AdapterInfo
	for _, name := range names {
		if adapter, ok := h.mcpReg.Get(name); ok {
			list = append(list, AdapterInfo{
				Name:             adapter.Name(),
				Healthy:          adapter.HealthCheck(r.Context()) == nil,
				SupportedActions: adapter.SupportedActions(),
			})
		}
	}

	respondJSON(w, http.StatusOK, list)
}

// ExecuteAction allows direct execution of an MCP action.
func (h *Handler) ExecuteAction(w http.ResponseWriter, r *http.Request) {
	var action mcp.Action
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}

	if h.mcpReg == nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "MCP registry not available"})
		return
	}

	adapter, ok := h.mcpReg.Get(action.AdapterID)
	if !ok {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "adapter not found: " + action.AdapterID})
		return
	}

	res, err := adapter.Execute(r.Context(), &action)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, res)
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}
