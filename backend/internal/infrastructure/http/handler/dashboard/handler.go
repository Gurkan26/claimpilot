package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/dashboard/usecase"
)

// Handler handles dashboard HTTP endpoints.
type Handler struct {
	summaryUC *usecase.GetDashboardSummaryUseCase
}

// NewHandler creates a new dashboard HTTP handler.
func NewHandler(summaryUC *usecase.GetDashboardSummaryUseCase) *Handler {
	return &Handler{summaryUC: summaryUC}
}

// Routes registers dashboard routes on the Chi router.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/summary", h.GetSummary)
}

// GetSummary returns the synthesized "Good Morning" dashboard metrics.
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := uuid.Nil
	if uidStr := r.Header.Get("X-User-ID"); uidStr != "" {
		if parsed, err := uuid.Parse(uidStr); err == nil {
			userID = parsed
		}
	}
	if userID == uuid.Nil {
		userID = uuid.MustParse("00000000-0000-0000-0000-000000000001") // demo fallback
	}

	resp, err := h.summaryUC.Execute(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
