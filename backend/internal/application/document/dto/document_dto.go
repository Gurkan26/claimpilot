package dto

import (
	"io"
	"time"

	"github.com/google/uuid"
	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
)

// UploadDocumentRequest represents an incoming file upload request.
type UploadDocumentRequest struct {
	UserID         uuid.UUID
	OrganizationID *uuid.UUID
	FileName       string
	FileType       docModel.DocumentType
	MimeType       string
	FileContent    io.Reader
	SyncProcess    bool // whether to synchronously wait for Orchestrator processing
	Tags           []string
}

// ExtractionFieldDTO represents an extracted key-value field.
type ExtractionFieldDTO struct {
	FieldName  string  `json:"field_name"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	PageRef    int     `json:"page_ref,omitempty"`
}

// ExtractionResultDTO represents the AI extraction details.
type ExtractionResultDTO struct {
	Fields      []ExtractionFieldDTO `json:"fields"`
	Summary     string               `json:"summary"`
	ProcessedAt time.Time            `json:"processed_at"`
	ModelUsed   string               `json:"model_used"`
}

// DocumentResponse represents document details returned to the client.
type DocumentResponse struct {
	ID               string               `json:"id"`
	UserID           string               `json:"user_id"`
	OrganizationID   string               `json:"organization_id,omitempty"`
	FileName         string               `json:"file_name"`
	FileType         string               `json:"file_type"`
	MimeType         string               `json:"mime_type"`
	FileSize         int64                `json:"file_size"`
	Status           string               `json:"status"`
	ExtractionResult *ExtractionResultDTO `json:"extraction_result,omitempty"`
	Tags             []string             `json:"tags,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

// DocumentListResponse represents paginated documents list.
type DocumentListResponse struct {
	Items      []*DocumentResponse `json:"items"`
	TotalCount int                 `json:"total_count"`
	Limit      int64               `json:"limit"`
	Offset     int64               `json:"offset"`
}

// ToDocumentResponse converts a domain Document entity to DTO.
func ToDocumentResponse(doc *docModel.Document) *DocumentResponse {
	if doc == nil {
		return nil
	}

	resp := &DocumentResponse{
		ID:        doc.ID.Hex(),
		UserID:    doc.UserID.String(),
		FileName:  doc.FileName,
		FileType:  string(doc.FileType),
		MimeType:  doc.MimeType,
		FileSize:  doc.FileSize,
		Status:    string(doc.Status),
		Tags:      doc.Tags,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}

	if doc.OrganizationID != nil {
		resp.OrganizationID = doc.OrganizationID.String()
	}

	if doc.ExtractionResult != nil {
		fields := make([]ExtractionFieldDTO, len(doc.ExtractionResult.Fields))
		for i, f := range doc.ExtractionResult.Fields {
			fields[i] = ExtractionFieldDTO{
				FieldName:  f.FieldName,
				Value:      f.Value,
				Confidence: f.Confidence,
				PageRef:    f.PageRef,
			}
		}
		resp.ExtractionResult = &ExtractionResultDTO{
			Fields:      fields,
			Summary:     doc.ExtractionResult.Summary,
			ProcessedAt: doc.ExtractionResult.ProcessedAt,
			ModelUsed:   doc.ExtractionResult.ModelUsed,
		}
	}

	return resp
}
