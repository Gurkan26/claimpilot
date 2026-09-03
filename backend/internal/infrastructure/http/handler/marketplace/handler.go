package marketplace

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/marketplace/dto"
	"github.com/masterfabric-go/masterfabric/internal/application/marketplace/usecase"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
)

// Handler handles marketplace HTTP endpoints.
type Handler struct {
	rfqUC     *usecase.TriggerRFQUseCase
	acceptUC  *usecase.AcceptBidUseCase
	metricsUC *usecase.GetMetricsUseCase
	listUC    *usecase.ListOpportunitiesUseCase
}

// Config holds dependencies for marketplace Handler.
type Config struct {
	RFQUC     *usecase.TriggerRFQUseCase
	AcceptUC  *usecase.AcceptBidUseCase
	MetricsUC *usecase.GetMetricsUseCase
	ListUC    *usecase.ListOpportunitiesUseCase
}

// NewHandler creates a new marketplace HTTP handler.
func NewHandler(cfg Config) *Handler {
	return &Handler{
		rfqUC:     cfg.RFQUC,
		acceptUC:  cfg.AcceptUC,
		metricsUC: cfg.MetricsUC,
		listUC:    cfg.ListUC,
	}
}

// Routes registers marketplace routes on Chi router.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/opportunities", h.ListOpportunities)
	r.Get("/opportunities/{id}", h.GetOpportunity)
	r.Post("/opportunities/{id}/rfq", h.TriggerRFQ)
	r.Post("/opportunities/{id}/accept-bid", h.AcceptBid)
	r.Get("/metrics", h.GetMetrics)
	r.Get("/transactions", h.ListTransactions)
}

// ListOpportunities handles querying marketplace opportunities.
func (h *Handler) ListOpportunities(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	var statuses []mktModel.OpportunityStatus
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		for _, s := range strings.Split(statusParam, ",") {
			statuses = append(statuses, mktModel.OpportunityStatus(strings.TrimSpace(strings.ToUpper(s))))
		}
	}

	opps, err := h.listUC.List(r.Context(), userID, statuses, limit, offset)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, opps)
}

// GetOpportunity retrieves single opportunity with its bids.
func (h *Handler) GetOpportunity(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	opp, err := h.listUC.GetByID(r.Context(), idStr)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, opp)
}

// TriggerRFQ starts automated supplier bid collection.
func (h *Handler) TriggerRFQ(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	resp, err := h.rfqUC.Execute(r.Context(), idStr)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// AcceptBid accepts a bid, generates transaction, and dispatches notifications.
func (h *Handler) AcceptBid(w http.ResponseWriter, r *http.Request) {
	oppID := chi.URLParam(r, "id")
	userID := getUserID(r)

	var body struct {
		BidID string `json:"bid_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "bid_id is required: " + err.Error()})
		return
	}

	req := dto.AcceptBidRequest{
		OpportunityID: oppID,
		BidID:         body.BidID,
		UserID:        userID,
	}

	resp, err := h.acceptUC.Execute(r.Context(), req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetMetrics returns platform GMV, savings, and commission metrics.
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	metrics, err := h.metricsUC.Execute(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, metrics)
}

// ListTransactions returns list of finalized deals.
func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	txns, err := h.listUC.ListTransactions(r.Context(), userID, limit, offset)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, txns)
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
