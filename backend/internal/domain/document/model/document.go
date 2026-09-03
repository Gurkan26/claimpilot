package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// DocumentType categorizes the uploaded document.
type DocumentType string

const (
	DocumentTypeContract     DocumentType = "CONTRACT"
	DocumentTypeInvoice      DocumentType = "INVOICE"
	DocumentTypeLicense      DocumentType = "LICENSE"
	DocumentTypePolicy       DocumentType = "POLICY"
	DocumentTypeSubscription DocumentType = "SUBSCRIPTION"
	DocumentTypeOther        DocumentType = "OTHER"
)

// DocumentStatus tracks the processing state of a document.
type DocumentStatus string

const (
	DocumentStatusUploaded   DocumentStatus = "UPLOADED"
	DocumentStatusProcessing DocumentStatus = "PROCESSING"
	DocumentStatusAnalyzed   DocumentStatus = "ANALYZED"
	DocumentStatusFailed     DocumentStatus = "FAILED"
)

// ExtractionField represents a single extracted key-value from the document.
type ExtractionField struct {
	FieldName  string `bson:"field_name" json:"field_name"`
	Value      string `bson:"value" json:"value"`
	Confidence float64 `bson:"confidence" json:"confidence"`
	PageRef    int     `bson:"page_ref,omitempty" json:"page_ref,omitempty"`
}

// ExtractionResult holds the complete extraction output from the Analyst Agent.
type ExtractionResult struct {
	Fields      []ExtractionField `bson:"fields" json:"fields"`
	RawText     string            `bson:"raw_text,omitempty" json:"raw_text,omitempty"`
	Summary     string            `bson:"summary" json:"summary"`
	ProcessedAt time.Time         `bson:"processed_at" json:"processed_at"`
	ModelUsed   string            `bson:"model_used" json:"model_used"`
}

// Document represents an uploaded document (contract, invoice, license, etc.).
type Document struct {
	ID               bson.ObjectID    `bson:"_id,omitempty" json:"id"`
	UserID           uuid.UUID        `bson:"user_id" json:"user_id"`
	OrganizationID   *uuid.UUID       `bson:"organization_id,omitempty" json:"organization_id,omitempty"`
	FileName         string           `bson:"file_name" json:"file_name"`
	FileType         DocumentType     `bson:"file_type" json:"file_type"`
	MimeType         string           `bson:"mime_type" json:"mime_type"`
	FileSize         int64            `bson:"file_size" json:"file_size"`
	StoragePath      string           `bson:"storage_path" json:"storage_path"`
	Status           DocumentStatus   `bson:"status" json:"status"`
	ExtractionResult *ExtractionResult `bson:"extraction_result,omitempty" json:"extraction_result,omitempty"`
	RawContent       string           `bson:"raw_content,omitempty" json:"-"`
	RedactedContent  string           `bson:"redacted_content,omitempty" json:"-"`
	Tags             []string         `bson:"tags,omitempty" json:"tags,omitempty"`
	CreatedAt        time.Time        `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time        `bson:"updated_at" json:"updated_at"`
}

// IsProcessed checks if the document has been analyzed.
func (d *Document) IsProcessed() bool {
	return d.Status == DocumentStatusAnalyzed
}

// IsFailed checks if document processing failed.
func (d *Document) IsFailed() bool {
	return d.Status == DocumentStatusFailed
}
