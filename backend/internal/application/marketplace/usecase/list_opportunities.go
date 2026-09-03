package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/marketplace/dto"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	mktRepo "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ListOpportunitiesUseCase handles querying opportunities, bids, and transactions.
type ListOpportunitiesUseCase struct {
	mktRepo mktRepo.MarketplaceRepository
}

// NewListOpportunitiesUseCase creates a new ListOpportunitiesUseCase.
func NewListOpportunitiesUseCase(mktRepo mktRepo.MarketplaceRepository) *ListOpportunitiesUseCase {
	return &ListOpportunitiesUseCase{mktRepo: mktRepo}
}

// List retrieves opportunities for a user with bids populated.
func (uc *ListOpportunitiesUseCase) List(
	ctx context.Context,
	userID uuid.UUID,
	statuses []mktModel.OpportunityStatus,
	limit, offset int64,
) ([]*dto.OpportunityResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	opps, err := uc.mktRepo.FindOpportunitiesByUserID(ctx, userID, statuses, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find opportunities: %w", err)
	}

	result := make([]*dto.OpportunityResponse, len(opps))
	for i, opp := range opps {
		bids, _ := uc.mktRepo.FindBidsByOpportunityID(ctx, opp.ID)
		result[i] = dto.ToOpportunityResponse(opp, bids)
	}

	return result, nil
}

// GetByID retrieves a single opportunity by its ID with all bids populated.
func (uc *ListOpportunitiesUseCase) GetByID(ctx context.Context, idStr string) (*dto.OpportunityResponse, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid opportunity ID: %w", err)
	}

	opp, err := uc.mktRepo.FindOpportunityByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find opportunity: %w", err)
	}

	bids, _ := uc.mktRepo.FindBidsByOpportunityID(ctx, id)
	return dto.ToOpportunityResponse(opp, bids), nil
}

// ListTransactions retrieves completed deals for reporting.
func (uc *ListOpportunitiesUseCase) ListTransactions(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int64,
) ([]*dto.TransactionResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	txns, err := uc.mktRepo.FindTransactionsByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find transactions: %w", err)
	}

	result := make([]*dto.TransactionResponse, len(txns))
	for i, t := range txns {
		result[i] = dto.ToTransactionResponse(t)
	}

	return result, nil
}
