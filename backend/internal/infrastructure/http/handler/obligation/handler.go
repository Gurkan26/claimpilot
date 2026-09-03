package obligation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/obligation/dto"
	"github.com/masterfabric-go/masterfabric/internal/application/obligation/usecase"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
)

// Handler handles obligation HTTP endpoints.
type Handler struct {
	approveUC *usecase.ApproveObligationUseCase
	dismissUC *usecase.DismissObligationUseCase
	listUC    *usecase.ListObligationsUseCase
}

// Config holds dependencies for obligation Handler.
type Config struct {
	ApproveUC *usecase.ApproveObligationUseCase
	DismissUC *usecase.DismissObligationUseCase
	ListUC    *usecase.ListObligationsUseCase
}

// NewHandler creates a new obligation HTTP handler.
func NewHandler(cfg Config) *Handler {
	return &Handler{
		approveUC: cfg.ApproveUC,
		dismissUC: cfg.DismissUC,
		listUC:    cfg.ListUC,
	}
}

// Routes registers obligation routes on the Chi router.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Post("/{id}/approve", h.Approve)
	r.Post("/{id}/dismiss", h.Dismiss)
}

// List handles querying obligations with filters.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	daysAhead, _ := strconv.Atoi(r.URL.Query().Get("days_ahead"))
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	var statuses []oblModel.ObligationStatus
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		for _, s := range strings.Split(statusParam, ",") {
			statuses = append(statuses, oblModel.ObligationStatus(strings.TrimSpace(strings.ToUpper(s))))
		}
	}

	params := usecase.FilterParams{
		UserID:    userID,
		Statuses:  statuses,
		DaysAhead: daysAhead,
		Limit:     limit,
		Offset:    offset,
	}

	resp, err := h.listUC.Execute(r.Context(), params)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Get retrieves a single obligation by ID.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	resp, err := h.listUC.GetByID(r.Context(), idStr)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Approve approves an obligation and optionally executes its MCP action.
func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID := getUserID(r)

	var body struct {
		ExecuteMCP bool           `json:"execute_mcp"`
		Parameters map[string]any `json:"parameters"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	// Default execute_mcp to true if not specified in query or body
	executeMCP := body.ExecuteMCP || r.URL.Query().Get("execute_mcp") != "false"

	req := dto.ApproveObligationRequest{
		ObligationID: idStr,
		UserID:       userID,
		ExecuteMCP:   executeMCP,
		Parameters:   body.Parameters,
	}

	resp, err := h.approveUC.Execute(r.Context(), req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Dismiss marks an obligation as dismissed.
func (h *Handler) Dismiss(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID := getUserID(r)

	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	req := dto.DismissObligationRequest{
		ObligationID: idStr,
		UserID:       userID,
		Reason:       body.Reason,
	}

	resp, err := h.dismissUC.Execute(r.Context(), req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func getUserID(r *http.Request) uuid.UUID {
	if uidStr := r.Header.Get("X-User-ID"); uidStr != "" {
		if parsed, err := uuid.Parse(uidStr); err == nil {
			return parsed
		}
	}
	return uuid.MustParse("00000000-0000-0000-0000-000000000001") // demo fallback user
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}
