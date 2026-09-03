package usecase

import (
	"context"
	"fmt"
	"log/slog"

	mktAgent "github.com/masterfabric-go/masterfabric/internal/agent/marketplace"
	"github.com/masterfabric-go/masterfabric/internal/application/marketplace/dto"
	mktRepo "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/repository"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TriggerRFQUseCase coordinates automated supplier bid collection.
type TriggerRFQUseCase struct {
	rfqEngine *mktAgent.RFQEngine
	mktRepo   mktRepo.MarketplaceRepository
	eventBus  events.EventBus
	logger    *slog.Logger
}

// NewTriggerRFQUseCase creates a new TriggerRFQUseCase.
func NewTriggerRFQUseCase(
	rfqEngine *mktAgent.RFQEngine,
	mktRepo mktRepo.MarketplaceRepository,
	eventBus events.EventBus,
	logger *slog.Logger,
) *TriggerRFQUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &TriggerRFQUseCase{
		rfqEngine: rfqEngine,
		mktRepo:   mktRepo,
		eventBus:  eventBus,
		logger:    logger.With("usecase", "trigger_rfq"),
	}
}

// Execute initiates the RFQ process for an opportunity.
func (uc *TriggerRFQUseCase) Execute(ctx context.Context, oppIDStr string) (*dto.OpportunityResponse, error) {
	id, err := bson.ObjectIDFromHex(oppIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid opportunity ID: %w", err)
	}

	opp, err := uc.mktRepo.FindOpportunityByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find opportunity: %w", err)
	}

	bids, err := uc.rfqEngine.CollectBids(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("collect bids: %w", err)
	}

	// Publish domain event
	if uc.eventBus != nil {
		_ = uc.eventBus.Publish(ctx, events.TopicTenant, map[string]interface{}{
			"type":           "marketplace.rfq_completed",
			"opportunity_id": id.Hex(),
			"bids_count":     len(bids),
		})
	}

	// Reload opportunity
	if updated, err := uc.mktRepo.FindOpportunityByID(ctx, id); err == nil {
		opp = updated
	}

	return dto.ToOpportunityResponse(opp, bids), nil
}
