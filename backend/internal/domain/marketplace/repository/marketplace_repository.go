package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// MarketplaceRepository defines the persistence contract for marketplace entities.
type MarketplaceRepository interface {
	// --- Opportunities ---

	// CreateOpportunity inserts a new marketplace opportunity.
	CreateOpportunity(ctx context.Context, opp *model.MarketplaceOpportunity) error

	// FindOpportunityByID retrieves an opportunity by ID.
	FindOpportunityByID(ctx context.Context, id bson.ObjectID) (*model.MarketplaceOpportunity, error)

	// FindOpportunitiesByUserID retrieves opportunities for a user.
	FindOpportunitiesByUserID(ctx context.Context, userID uuid.UUID, statusFilter []model.OpportunityStatus, limit, offset int64) ([]*model.MarketplaceOpportunity, error)

	// UpdateOpportunityStatus transitions an opportunity to a new status.
	UpdateOpportunityStatus(ctx context.Context, id bson.ObjectID, status model.OpportunityStatus) error

	// --- Bids ---

	// CreateBid inserts a new bid for an opportunity.
	CreateBid(ctx context.Context, bid *model.Bid) error

	// FindBidByID retrieves a bid by ID.
	FindBidByID(ctx context.Context, id bson.ObjectID) (*model.Bid, error)

	// FindBidsByOpportunityID retrieves all bids for a given opportunity.
	FindBidsByOpportunityID(ctx context.Context, oppID bson.ObjectID) ([]*model.Bid, error)

	// UpdateBidStatus transitions a bid to a new status.
	UpdateBidStatus(ctx context.Context, id bson.ObjectID, status model.BidStatus) error

	// UpdateBidRiskScore stores the Verifier Agent's risk assessment.
	UpdateBidRiskScore(ctx context.Context, id bson.ObjectID, score float64, notes string) error

	// --- Transactions ---

	// CreateTransaction records a completed marketplace deal.
	CreateTransaction(ctx context.Context, txn *model.Transaction) error

	// FindTransactionByID retrieves a transaction by ID.
	FindTransactionByID(ctx context.Context, id bson.ObjectID) (*model.Transaction, error)

	// FindTransactionsByUserID retrieves transactions for a user.
	FindTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]*model.Transaction, error)

	// UpdateTransactionStatus transitions a transaction status.
	UpdateTransactionStatus(ctx context.Context, id bson.ObjectID, status model.TransactionStatus) error

	// --- Aggregations (dashboard/reporting) ---

	// TotalGMV calculates total gross merchandise value for a user or org.
	TotalGMV(ctx context.Context, userID uuid.UUID) (*model.Money, error)

	// TotalSavings calculates total savings realized through vendor switches.
	TotalSavings(ctx context.Context, userID uuid.UUID) (*model.Money, error)

	// --- GDPR/KVKK ---

	// DeleteByUserID removes all marketplace data for a user.
	DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
