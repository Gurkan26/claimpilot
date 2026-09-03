package usecase

import (
	"context"
	"fmt"

	"github.com/masterfabric-go/masterfabric/internal/application/document/dto"
	docRepo "github.com/masterfabric-go/masterfabric/internal/domain/document/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// GetDocumentUseCase retrieves document details by ID.
type GetDocumentUseCase struct {
	docRepo docRepo.DocumentRepository
}

// NewGetDocumentUseCase creates a new GetDocumentUseCase.
func NewGetDocumentUseCase(docRepo docRepo.DocumentRepository) *GetDocumentUseCase {
	return &GetDocumentUseCase{docRepo: docRepo}
}

// Execute fetches the document by its ObjectID hex string.
func (uc *GetDocumentUseCase) Execute(ctx context.Context, idStr string) (*dto.DocumentResponse, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID format: %w", err)
	}

	doc, err := uc.docRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find document: %w", err)
	}

	return dto.ToDocumentResponse(doc), nil
}
