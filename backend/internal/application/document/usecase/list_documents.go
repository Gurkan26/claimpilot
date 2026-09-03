package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/document/dto"
	docRepo "github.com/masterfabric-go/masterfabric/internal/domain/document/repository"
)

// ListDocumentsUseCase retrieves documents belonging to a user or organization.
type ListDocumentsUseCase struct {
	docRepo docRepo.DocumentRepository
}

// NewListDocumentsUseCase creates a new ListDocumentsUseCase.
func NewListDocumentsUseCase(docRepo docRepo.DocumentRepository) *ListDocumentsUseCase {
	return &ListDocumentsUseCase{docRepo: docRepo}
}

// Execute retrieves documents for a given user with pagination.
func (uc *ListDocumentsUseCase) Execute(ctx context.Context, userID uuid.UUID, limit, offset int64) (*dto.DocumentListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	docs, err := uc.docRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	items := make([]*dto.DocumentResponse, len(docs))
	for i, d := range docs {
		items[i] = dto.ToDocumentResponse(d)
	}

	return &dto.DocumentListResponse{
		Items:      items,
		TotalCount: len(items),
		Limit:      limit,
		Offset:     offset,
	}, nil
}
