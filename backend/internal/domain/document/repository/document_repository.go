package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// DocumentRepository defines the persistence contract for documents.
type DocumentRepository interface {
	// Create inserts a new document record.
	Create(ctx context.Context, doc *model.Document) error

	// FindByID retrieves a document by its ObjectID.
	FindByID(ctx context.Context, id bson.ObjectID) (*model.Document, error)

	// FindByUserID retrieves all documents belonging to a user.
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]*model.Document, error)

	// FindByOrganizationID retrieves all documents belonging to an organization.
	FindByOrganizationID(ctx context.Context, orgID uuid.UUID, limit, offset int64) ([]*model.Document, error)

	// UpdateStatus updates the processing status of a document.
	UpdateStatus(ctx context.Context, id bson.ObjectID, status model.DocumentStatus) error

	// UpdateExtractionResult stores the AI extraction output.
	UpdateExtractionResult(ctx context.Context, id bson.ObjectID, result *model.ExtractionResult) error

	// UpdateContent stores the extracted raw text and redacted text.
	UpdateContent(ctx context.Context, id bson.ObjectID, rawContent, redactedContent string) error

	// Delete removes a document (GDPR/KVKK right to erasure).
	Delete(ctx context.Context, id bson.ObjectID) error

	// DeleteByUserID removes all documents for a user (account deletion).
	DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
