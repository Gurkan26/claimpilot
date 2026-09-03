package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/agent/orchestrator"
	"github.com/masterfabric-go/masterfabric/internal/application/document/dto"
	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/document/parser"
	docRepo "github.com/masterfabric-go/masterfabric/internal/domain/document/repository"
	"github.com/masterfabric-go/masterfabric/internal/domain/document/storage"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// UploadDocumentUseCase handles file storage, text extraction, and agent pipeline triggering.
type UploadDocumentUseCase struct {
	docRepo      docRepo.DocumentRepository
	storage      storage.StorageService
	parser       parser.Parser
	orchestrator *orchestrator.Orchestrator
	eventBus     events.EventBus
	logger       *slog.Logger
}

// Config holds dependencies for UploadDocumentUseCase.
type UploadConfig struct {
	DocRepo      docRepo.DocumentRepository
	Storage      storage.StorageService
	Parser       parser.Parser
	Orchestrator *orchestrator.Orchestrator
	EventBus     events.EventBus
	Logger       *slog.Logger
}

// NewUploadDocumentUseCase creates a new UploadDocumentUseCase.
func NewUploadDocumentUseCase(cfg UploadConfig) *UploadDocumentUseCase {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &UploadDocumentUseCase{
		docRepo:      cfg.DocRepo,
		storage:      cfg.Storage,
		parser:       cfg.Parser,
		orchestrator: cfg.Orchestrator,
		eventBus:     cfg.EventBus,
		logger:       logger.With("usecase", "upload_document"),
	}
}

// Execute processes the document upload.
func (uc *UploadDocumentUseCase) Execute(ctx context.Context, req dto.UploadDocumentRequest) (*dto.DocumentResponse, error) {
	if req.FileName == "" {
		return nil, fmt.Errorf("file name is required")
	}
	if req.FileContent == nil {
		return nil, fmt.Errorf("file content is required")
	}

	// 1. Save file to storage
	storagePath, size, err := uc.storage.Save(ctx, req.FileName, req.FileContent)
	if err != nil {
		return nil, fmt.Errorf("store document: %w", err)
	}

	// 2. Open saved file to extract text
	fileReader, err := uc.storage.Open(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("open stored document for extraction: %w", err)
	}
	defer fileReader.Close()

	rawContent, err := uc.parser.Parse(ctx, req.MimeType, fileReader)
	if err != nil {
		uc.logger.Warn("text extraction had warnings, proceeding with empty text", "error", err)
		rawContent = ""
	}

	// 3. Detect or refine document type if not specified
	fileType := req.FileType
	if fileType == "" || fileType == docModel.DocumentTypeOther {
		detected := parser.DetectDocumentType(req.FileName, rawContent)
		if detected != docModel.DocumentTypeOther {
			fileType = detected
		} else if fileType == "" {
			fileType = docModel.DocumentTypeOther
		}
	}

	// 4. Create document entity
	doc := &docModel.Document{
		ID:             bson.NewObjectID(),
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		FileName:       req.FileName,
		FileType:       fileType,
		MimeType:       req.MimeType,
		FileSize:       size,
		StoragePath:    storagePath,
		Status:         docModel.DocumentStatusUploaded,
		RawContent:     rawContent,
		Tags:           req.Tags,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := uc.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("save document entity: %w", err)
	}

	uc.logger.Info("document uploaded successfully",
		"document_id", doc.ID.Hex(),
		"file_name", doc.FileName,
		"file_type", doc.FileType,
		"size_bytes", doc.FileSize,
		"content_len", len(rawContent),
	)

	// Publish domain event
	if uc.eventBus != nil {
		_ = uc.eventBus.Publish(ctx, events.TopicTenant, map[string]interface{}{
			"type":        "document.uploaded",
			"document_id": doc.ID.Hex(),
			"user_id":     doc.UserID.String(),
			"file_type":   string(doc.FileType),
		})
	}

	// 5. Trigger Orchestrator pipeline
	if uc.orchestrator != nil && rawContent != "" {
		if req.SyncProcess {
			// Synchronous execution for demo/testing
			if err := uc.orchestrator.ProcessDocument(ctx, doc.ID); err != nil {
				uc.logger.Error("synchronous document processing failed", "error", err)
			} else {
				// Reload updated document
				if updatedDoc, err := uc.docRepo.FindByID(ctx, doc.ID); err == nil {
					doc = updatedDoc
				}
			}
		} else {
			// Asynchronous execution in background
			go func(docID bson.ObjectID) {
				bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				if err := uc.orchestrator.ProcessDocument(bgCtx, docID); err != nil {
					uc.logger.Error("async document processing failed", "error", err, "document_id", docID.Hex())
				}
			}(doc.ID)
		}
	}

	return dto.ToDocumentResponse(doc), nil
}
