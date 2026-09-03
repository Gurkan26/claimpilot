package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/obligation/dto"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	oblRepo "github.com/masterfabric-go/masterfabric/internal/domain/obligation/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ListObligationsUseCase handles querying obligations with status and deadline filters.
type ListObligationsUseCase struct {
	oblRepo oblRepo.ObligationRepository
}

// NewListObligationsUseCase creates a new ListObligationsUseCase.
func NewListObligationsUseCase(oblRepo oblRepo.ObligationRepository) *ListObligationsUseCase {
	return &ListObligationsUseCase{oblRepo: oblRepo}
}

// FilterParams defines search parameters for obligations.
type FilterParams struct {
	UserID    uuid.UUID
	Statuses  []oblModel.ObligationStatus
	DaysAhead int // if > 0, returns upcoming obligations due in N days
	Limit     int64
	Offset    int64
}

// Execute retrieves matching obligations.
func (uc *ListObligationsUseCase) Execute(ctx context.Context, params FilterParams) (*dto.ObligationListResponse, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	var obligations []*oblModel.Obligation
	var err error

	if params.DaysAhead > 0 {
		obligations, err = uc.oblRepo.FindUpcoming(ctx, params.UserID, params.DaysAhead)
	} else {
		obligations, err = uc.oblRepo.FindByUserID(ctx, params.UserID, params.Statuses, params.Limit, params.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("query obligations: %w", err)
	}

	items := make([]*dto.ObligationResponse, len(obligations))
	for i, o := range obligations {
		items[i] = dto.ToObligationResponse(o)
	}

	return &dto.ObligationListResponse{
		Items:      items,
		TotalCount: len(items),
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

// GetByID retrieves a single obligation by ID.
func (uc *ListObligationsUseCase) GetByID(ctx context.Context, idStr string) (*dto.ObligationResponse, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid obligation ID: %w", err)
	}

	obl, err := uc.oblRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find obligation: %w", err)
	}

	return dto.ToObligationResponse(obl), nil
}
