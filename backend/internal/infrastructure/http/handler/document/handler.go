package document

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/agent/orchestrator"
	"github.com/masterfabric-go/masterfabric/internal/application/document/dto"
	"github.com/masterfabric-go/masterfabric/internal/application/document/usecase"
	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	oblRepo "github.com/masterfabric-go/masterfabric/internal/domain/obligation/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Handler handles document HTTP endpoints.
type Handler struct {
	uploadUC     *usecase.UploadDocumentUseCase
	getUC        *usecase.GetDocumentUseCase
	listUC       *usecase.ListDocumentsUseCase
	orchestrator *orchestrator.Orchestrator
	oblRepo      oblRepo.ObligationRepository
}

// Config holds dependencies for document Handler.
type Config struct {
	UploadUC     *usecase.UploadDocumentUseCase
	GetUC        *usecase.GetDocumentUseCase
	ListUC       *usecase.ListDocumentsUseCase
	Orchestrator *orchestrator.Orchestrator
	OblRepo      oblRepo.ObligationRepository
}

// NewHandler creates a new document HTTP handler.
func NewHandler(cfg Config) *Handler {
	return &Handler{
		uploadUC:     cfg.UploadUC,
		getUC:        cfg.GetUC,
		listUC:       cfg.ListUC,
		orchestrator: cfg.Orchestrator,
		oblRepo:      cfg.OblRepo,
	}
}

// Routes registers document routes on the router.
func (h *Handler) Routes(r chi.Router) {
	r.Post("/upload", h.Upload)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Post("/{id}/process", h.Process)
	r.Get("/{id}/obligations", h.GetObligations)
}

// Upload handles multipart document uploads.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form up to 32MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form: " + err.Error()})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "file field is required: " + err.Error()})
		return
	}
	defer file.Close()

	fileTypeStr := strings.ToUpper(r.FormValue("file_type"))
	var fileType docModel.DocumentType
	switch fileTypeStr {
	case "CONTRACT":
		fileType = docModel.DocumentTypeContract
	case "INVOICE":
		fileType = docModel.DocumentTypeInvoice
	case "LICENSE":
		fileType = docModel.DocumentTypeLicense
	case "POLICY":
		fileType = docModel.DocumentTypePolicy
	case "SUBSCRIPTION":
		fileType = docModel.DocumentTypeSubscription
	default:
		fileType = docModel.DocumentTypeOther
	}

	syncProcess := r.FormValue("sync") == "true" || r.URL.Query().Get("sync") == "true"

	// Demo/fallback user ID if not extracted from auth header
	userID := uuid.Nil
	if uidStr := r.Header.Get("X-User-ID"); uidStr != "" {
		if parsed, err := uuid.Parse(uidStr); err == nil {
			userID = parsed
		}
	}
	if userID == uuid.Nil {
		userID = uuid.MustParse("00000000-0000-0000-0000-000000000001") // Default demo user
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	req := dto.UploadDocumentRequest{
		UserID:      userID,
		FileName:    header.Filename,
		FileType:    fileType,
		MimeType:    mimeType,
		FileContent: file,
		SyncProcess: syncProcess,
	}

	resp, err := h.uploadUC.Execute(r.Context(), req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// Get handles fetching single document by ID.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	resp, err := h.getUC.Execute(r.Context(), idStr)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// List handles listing user documents.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := uuid.Nil
	if uidStr := r.Header.Get("X-User-ID"); uidStr != "" {
		if parsed, err := uuid.Parse(uidStr); err == nil {
			userID = parsed
		}
	}
	if userID == uuid.Nil {
		userID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	resp, err := h.listUC.Execute(r.Context(), userID, limit, offset)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Process triggers or re-runs analysis on an existing document.
func (h *Handler) Process(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid document ID"})
		return
	}

	if h.orchestrator == nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "agent orchestrator is not available"})
		return
	}

	syncProcess := r.URL.Query().Get("sync") == "true"
	if syncProcess {
		if err := h.orchestrator.ProcessDocument(r.Context(), id); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		docResp, err := h.getUC.Execute(r.Context(), idStr)
		if err != nil {
			respondJSON(w, http.StatusOK, map[string]string{"status": "processed"})
			return
		}
		respondJSON(w, http.StatusOK, docResp)
		return
	}

	// Asynchronous
	go func(docID bson.ObjectID) {
		_ = h.orchestrator.ProcessDocument(r.Context(), docID)
	}(id)

	respondJSON(w, http.StatusAccepted, map[string]string{
		"status":      "processing",
		"document_id": idStr,
	})
}

// GetObligations lists obligations generated from this document.
func (h *Handler) GetObligations(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid document ID"})
		return
	}

	if h.oblRepo == nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "obligation repository not available"})
		return
	}

	obligations, err := h.oblRepo.FindByDocumentID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, obligations)
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			_ = errors.New(fmt.Sprintf("error encoding json: %v", err))
		}
	}
}
