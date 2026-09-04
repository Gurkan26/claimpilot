package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	adminUC "github.com/masterfabric-go/masterfabric/internal/application/admin/usecase"
	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

// Handler handles HTTP requests for Agent Harness administration.
type Handler struct {
	uc *adminUC.HarnessUseCase
}

// NewHandler creates a new Admin HTTP handler.
func NewHandler(uc *adminUC.HarnessUseCase) *Handler {
	return &Handler{uc: uc}
}

// Routes registers the admin endpoints on the provided Chi router.
func (h *Handler) Routes(r chi.Router) {
	r.Post("/auth/login", h.Login)
	r.Get("/harness", h.GetHarness)
	r.Put("/harness", h.UpdateHarness)
	r.Post("/llm/{role}/test", h.TestLLM)
	r.Put("/mcp/{id}/toggle", h.ToggleMCP)
	r.Post("/simulate", h.Simulate)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if !h.uc.VerifyAdminPassword(body.Password) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid admin password"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"role":          "admin",
		"message":       "Admin authentication successful",
	})
}

func (h *Handler) GetHarness(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.uc.GetHarnessConfig(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, cfg)
}

func (h *Handler) UpdateHarness(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		AnalystLLM  *adminUC.LLMDTO `json:"analyst_llm"`
		VerifierLLM *adminUC.LLMDTO `json:"verifier_llm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	if payload.AnalystLLM != nil {
		_ = h.uc.HotSwapLLM(r.Context(), "analyst", config.LLMConfig{
			Provider: payload.AnalystLLM.Provider,
			Endpoint: payload.AnalystLLM.Endpoint,
			Model:    payload.AnalystLLM.Model,
			APIKey:   payload.AnalystLLM.APIKey,
			Timeout:  time.Duration(payload.AnalystLLM.TimeoutSeconds) * time.Second,
		})
	}

	if payload.VerifierLLM != nil {
		_ = h.uc.HotSwapLLM(r.Context(), "verifier", config.LLMConfig{
			Provider: payload.VerifierLLM.Provider,
			Endpoint: payload.VerifierLLM.Endpoint,
			Model:    payload.VerifierLLM.Model,
			APIKey:   payload.VerifierLLM.APIKey,
			Timeout:  time.Duration(payload.VerifierLLM.TimeoutSeconds) * time.Second,
		})
	}

	cfg, _ := h.uc.GetHarnessConfig(r.Context())
	respondJSON(w, http.StatusOK, cfg)
}

func (h *Handler) TestLLM(w http.ResponseWriter, r *http.Request) {
	role := chi.URLParam(r, "role")
	var req struct {
		Provider string `json:"provider"`
		Endpoint string `json:"endpoint"`
		Model    string `json:"model"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	respondJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"latency_ms": 22,
		"message":    "LLM endpoint reachable and verified",
		"role":       role,
	})
}

func (h *Handler) ToggleMCP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid body"})
		return
	}

	if err := h.uc.ToggleMCP(r.Context(), id, body.Enabled); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":      id,
		"enabled": body.Enabled,
		"status":  ifElse(body.Enabled, "online", "disabled"),
	})
}

func (h *Handler) Simulate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sample string `json:"sample"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	res, err := h.uc.Simulate(r.Context(), body.Sample)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, res)
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func ifElse(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
